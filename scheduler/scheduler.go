package scheduler

import (
	"sync"
	"time"
)

// ScheduledTask 一次性定时任务
type ScheduledTask struct {
	ID         string
	TargetTime time.Time // 目标执行时间
	TaskFunc   func()

	timer    *time.Timer
	stopChan chan struct{}
	running  bool
	mutex    sync.RWMutex
}

// DynamicScheduler 动态调度器（支持全局偏移）
type DynamicScheduler struct {
	tasks        map[string]*ScheduledTask
	globalOffset time.Duration // 全局偏移值
	mutex        sync.RWMutex
}

// NewDynamicScheduler 创建新的调度器
func NewDynamicScheduler() *DynamicScheduler {
	return &DynamicScheduler{
		tasks:        make(map[string]*ScheduledTask),
		globalOffset: 0,
	}
}

// NewScheduledTask 创建新的定时任务
func NewScheduledTask(id string, targetTime time.Time, taskFunc func()) *ScheduledTask {
	return &ScheduledTask{
		ID:         id,
		TargetTime: targetTime,
		TaskFunc:   taskFunc,
		stopChan:   make(chan struct{}),
	}
}

// SetGlobalOffset 设置全局偏移值（影响所有任务）
func (ds *DynamicScheduler) SetGlobalOffset(offset time.Duration) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	oldOffset := ds.globalOffset
	ds.globalOffset = offset

	// 重新调度所有运行中的任务
	for _, task := range ds.tasks {
		if task.IsRunning() {
			task.rescheduleWithNewOffset(offset - oldOffset)
		}
	}
}

// GetGlobalOffset 获取全局偏移值
func (ds *DynamicScheduler) GetGlobalOffset() time.Duration {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.globalOffset
}

// AddTask 添加一次性定时任务
func (ds *DynamicScheduler) AddTask(id string, targetTime time.Time, taskFunc func()) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	task := NewScheduledTask(id, targetTime, taskFunc)
	ds.tasks[id] = task
	task.Start(ds.globalOffset)
}

// RemoveTask 移除任务
func (ds *DynamicScheduler) RemoveTask(taskID string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if task, exists := ds.tasks[taskID]; exists {
		task.Stop()
		delete(ds.tasks, taskID)
	}
}

// GetTaskStatus 获取所有任务状态
func (ds *DynamicScheduler) GetTaskStatus() map[string]interface{} {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	status := make(map[string]interface{})
	for id, task := range ds.tasks {
		adjustedTime := task.TargetTime.Add(ds.globalOffset)
		taskStatus := map[string]interface{}{
			"original_time": task.TargetTime,
			"adjusted_time": adjustedTime,
			"remaining":     time.Until(adjustedTime).Round(time.Second),
			"completed":     !task.IsRunning(),
			"running":       task.IsRunning(),
		}
		status[id] = taskStatus
	}
	return status
}

// GetTaskCount 获取任务数量
func (ds *DynamicScheduler) GetTaskCount() int {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return len(ds.tasks)
}

// CleanupCompletedTasks 清理已完成的任务
func (ds *DynamicScheduler) CleanupCompletedTasks() {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	for id, task := range ds.tasks {
		if !task.IsRunning() {
			delete(ds.tasks, id)
		}
	}
}

// Start 启动定时任务（使用全局偏移）
func (st *ScheduledTask) Start(globalOffset time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.running {
		return
	}

	st.running = true
	go st.run(globalOffset)
}

// Stop 停止定时任务
func (st *ScheduledTask) Stop() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if !st.running {
		return
	}

	close(st.stopChan)
	if st.timer != nil {
		st.timer.Stop()
	}
	st.running = false
}

// IsRunning 检查任务是否在运行
func (st *ScheduledTask) IsRunning() bool {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	return st.running
}

// rescheduleWithNewOffset 重新调度任务（用于全局偏移更新）
func (st *ScheduledTask) rescheduleWithNewOffset(offsetDelta time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if !st.running || st.timer == nil {
		return
	}

	// 停止当前定时器
	if !st.timer.Stop() {
		select {
		case <-st.timer.C:
		default:
		}
	}

	// 重新计算执行时间（考虑新的全局偏移）
	adjustedTime := st.TargetTime.Add(offsetDelta)
	waitDuration := time.Until(adjustedTime)
	if waitDuration <= 0 {
		// 如果已经过期，立即执行
		go st.executeAndStop()
		return
	}

	// 创建新的定时器
	st.timer = time.NewTimer(waitDuration)
	go st.waitForExecution()
}

// run 核心运行逻辑
func (st *ScheduledTask) run(globalOffset time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if !st.running {
		return
	}

	// 计算调整后的执行时间
	adjustedTime := st.TargetTime.Add(globalOffset)
	waitDuration := time.Until(adjustedTime)

	if waitDuration <= 0 {
		// 如果已经过期，立即执行
		go st.executeAndStop()
		return
	}

	// 创建定时器
	st.timer = time.NewTimer(waitDuration)
	go st.waitForExecution()
}

// waitForExecution 等待定时器执行
func (st *ScheduledTask) waitForExecution() {
	select {
	case <-st.timer.C:
		st.executeAndStop()
	case <-st.stopChan:
		// 任务被停止
	}
}

// executeAndStop 执行任务并停止
func (st *ScheduledTask) executeAndStop() {
	st.mutex.Lock()
	if !st.running {
		st.mutex.Unlock()
		return
	}
	st.running = false
	st.mutex.Unlock()

	if st.TaskFunc != nil {
		st.TaskFunc()
	}
}

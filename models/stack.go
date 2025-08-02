package models

import (
	"sync"
)

type Stack[T any] struct {
	data  []T        //用于存储元素的动态数组
	top   uint64     //顶部指针
	cap   uint64     //动态数组的实际空间
	mutex sync.Mutex //并发控制锁
	z     T
}

func (s *Stack[T]) New() *Stack[T] {
	return &Stack[T]{
		data:  make([]T, 1),
		top:   0,
		cap:   1,
		mutex: sync.Mutex{},
	}
}

func (s *Stack[T]) Size() (num uint64) {
	if s == nil {
		s = &Stack[T]{
			data:  make([]T, 1),
			top:   0,
			cap:   1,
			mutex: sync.Mutex{},
		}
	}
	return s.top
}

func (s *Stack[T]) Clear() {
	if s == nil {
		s = &Stack[T]{
			data:  make([]T, 1),
			top:   0,
			cap:   1,
			mutex: sync.Mutex{},
		}
	}
	s.mutex.Lock()
	s.data = make([]T, 1)
	s.top = 0
	s.cap = 1
	s.mutex.Unlock()
}

func (s *Stack[T]) Empty() (b bool) {
	if s == nil {
		return true
	}
	return s.Size() == 0
}

func (s *Stack[T]) Push(e T) {
	if s == nil {
		s = &Stack[T]{
			data:  make([]T, 1),
			top:   0,
			cap:   1,
			mutex: sync.Mutex{},
		}
	}
	s.mutex.Lock()
	if s.top < s.cap {
		//还有冗余,直接添加
		s.data[s.top] = e
	} else {
		//冗余不足,需要扩容
		if s.cap <= 65536 {
			//容量翻倍
			if s.cap == 0 {
				s.cap = 1
			}
			s.cap *= 2
		} else {
			//容量增加2^16
			s.cap += 65536
		}
		//复制扩容前的元素
		tmp := make([]T, s.cap)
		copy(tmp, s.data)
		s.data = tmp
		s.data[s.top] = e
	}
	s.top++
	s.mutex.Unlock()
}

func (s *Stack[T]) Pop() {
	if s == nil {
		s = &Stack[T]{
			data:  make([]T, 1),
			top:   0,
			cap:   1,
			mutex: sync.Mutex{},
		}
		return
	}
	if s.Empty() {
		return
	}
	s.mutex.Lock()
	s.top--
	if s.cap-s.top >= 65536 {
		//容量和实际使用差值超过2^16时,容量直接减去2^16
		s.cap -= 65536
		tmp := make([]T, s.cap)
		copy(tmp, s.data)
		s.data = tmp
	} else if s.top*2 < s.cap {
		//实际使用长度是容量的一半时,进行折半缩容
		s.cap /= 2
		tmp := make([]T, s.cap)
		copy(tmp, s.data)
		s.data = tmp
	}
	s.mutex.Unlock()
}

func (s *Stack[T]) Top() (e T) {

	if s == nil {
		return s.z
	}
	if s.Empty() {
		return s.z
	}
	s.mutex.Lock()
	e = s.data[s.top-1]
	s.mutex.Unlock()
	return e
}

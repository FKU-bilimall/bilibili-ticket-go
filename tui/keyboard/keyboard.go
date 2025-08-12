package keyboard

import (
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/tui/primitives"
	"bilibili-ticket-go/utils"
	"reflect"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type selectItem struct {
	where         int
	obj           tview.Primitive
	previousColor tcell.Color
}

type KeyboardCaptureInstance struct {
	stack       *models.Stack[selectItem]
	app         *tview.Application
	root        tview.Primitive
	selected    selectItem
	isOpenModal bool
	mutex       sync.Mutex
}

var logger = utils.GetLogger(global.GetLogger(), "keyboard", nil)

const highlightColor = tcell.ColorForestGreen

func NewKeyboardCaptureInstance(app *tview.Application, root tview.Primitive) *KeyboardCaptureInstance {
	app.SetFocus(root)
	st := (&models.Stack[selectItem]{}).New()
	st.Push(selectItem{
		where:         -1,
		obj:           root,
		previousColor: tcell.ColorDefault,
	})
	return &KeyboardCaptureInstance{
		stack:    st,
		app:      app,
		root:     root,
		selected: selectItem{where: -1, obj: root},
	}
}

func (k *KeyboardCaptureInstance) Selected() bool {
	return k.selected.obj != nil && k.selected.obj != k.root || k.stack.Size() > 1
}

func (k *KeyboardCaptureInstance) Reset() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	// 恢复当前选中项的颜色（如果有）
	if k.selected.obj != nil {
		setColor(k.selected.obj, k.selected.previousColor)
	}
	var size = int(k.stack.Size() + 1)
	for i := 1; i < size; i++ {
		current := k.stack.Top()
		k.stack.Pop()
		// 恢复 current 的颜色
		setColor(current.obj, current.previousColor)
	}
	k.stack.Clear()
	k.stack.Push(selectItem{
		where:         -1,
		obj:           k.root,
		previousColor: tcell.ColorDefault,
	})
	k.selected = selectItem{where: -1, obj: k.root}
	k.app.SetFocus(k.root)
}

func (k *KeyboardCaptureInstance) InputCapture(event *tcell.EventKey) *tcell.EventKey {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	if k.isOpenModal {
		return event
	}

	if event.Key() == tcell.KeyEscape {
		if k.selected.obj != k.root && k.selected.obj != nil && k.app.GetFocus() == k.selected.obj {
			current := k.stack.Top()
			k.app.SetFocus(current.obj)
			// 恢复当前选中项的颜色
			//setColor(k.selected.obj, k.selected.previousColor)
			return event
		}
		if k.selected.obj != k.root && k.selected.obj != nil && k.app.GetFocus() == k.selected.obj {
			current := k.stack.Top()
			k.app.SetFocus(current.obj)
			// 恢复当前选中项的颜色
			setColor(k.selected.obj, k.selected.previousColor)
			return event
		}
		if k.stack.Size() > 1 {
			current := k.stack.Top()
			k.stack.Pop()
			previous := k.stack.Top()
			// 恢复当前选中项的颜色
			setColor(k.selected.obj, k.selected.previousColor)
			// 恢复 previous 的颜色
			//setColor(previous.obj, previous.previousColor)
			setColor(current.obj, highlightColor)
			k.selected = current
			k.app.SetFocus(previous.obj)
		} else if k.stack.Size() == 1 {
			k.app.SetFocus(k.root)
		}
		return event
	}

	if !allowDeepFocus(k.app.GetFocus()) {
		return event
	}

	if event.Key() == tcell.KeyTab {
		if k.stack.Size() == 0 {
			return event
		}
		item := k.stack.Top()
		k.switchToNextItem(item)
	}

	if event.Key() == tcell.KeyEnter {
		if k.selected.obj != nil && k.selected.obj != k.stack.Top().obj {
			k.app.SetFocus(k.selected.obj)
			switch k.selected.obj.(type) {
			case *tview.Flex:
				if k.selected.where == -1 {
					break
				}
				k.stack.Push(k.selected)
				k.selected = selectItem{
					where: -1,
					obj:   nil,
				}
				return nil
			case *primitives.Pages:
				k.stack.Push(k.selected)
				k.selected = selectItem{
					where: -1,
					obj:   nil,
				}
				return nil
			}
		}
	}
	return event
}

func (k *KeyboardCaptureInstance) switchToNextItem(box selectItem) {
	// 收集所有可选项（递归扁平化）
	items := flattenItems(box.obj)
	if len(items) == 0 {
		return
	}
	// 找到当前选中项在扁平化列表中的索引
	curIdx := -1
	for i, it := range items {
		if it == k.selected.obj {
			curIdx = i
			break
		}
	}
	// 下一个索引，循环
	nextIdx := (curIdx + 1) % len(items)
	// 切换高亮
	setColor(k.selected.obj, k.selected.previousColor)
	c := setColor(items[nextIdx], highlightColor)
	k.selected = selectItem{
		where:         nextIdx,
		obj:           items[nextIdx],
		previousColor: c,
	}
}

func (k *KeyboardCaptureInstance) SetIsOpenModel(stat bool) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	k.isOpenModal = stat
}

// 递归收集所有可选项
func flattenItems(obj tview.Primitive) []tview.Primitive {
	var result []tview.Primitive
	switch o := obj.(type) {
	case *tview.Flex:
		cnt := o.GetItemCount()
		for i := 0; i < cnt; i++ {
			child := o.GetItem(i)
			if selectable(child) {
				v := reflect.ValueOf(child)
				f := reflect.Indirect(v).FieldByName("border")
				if f.IsValid() && f.Kind() == reflect.Bool && !f.Bool() && allowDeepFocus(child) {
					// 只递归无边框的容器
					result = append(result, flattenItems(child)...)
				} else {
					result = append(result, child)
				}
			}
		}
	case *primitives.Pages:
		cur := o.GetCurrentPage()
		if cur != nil {
			result = append(result, flattenItems(cur)...)
		}
	default:
		result = append(result, obj)
	}
	return result
}

func selectable(obj tview.Primitive) bool {
	switch obj.(type) {
	case *tview.List, *tview.Flex, *tview.Grid, *tview.Button, *primitives.Pages, *tview.InputField, *tview.DropDown:
		return true
	default:
		return false
	}
}

func setColor(primitive tview.Primitive, color tcell.Color) tcell.Color {
	var c tcell.Color
	switch obj := primitive.(type) {
	case *tview.List:
		c = obj.GetBackgroundColor()
		obj.SetBackgroundColor(color)
	case *tview.Box:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
	case *tview.Flex:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
	case *tview.Grid:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
	case *tview.Button:
		c = obj.GetBackgroundColor()
		obj.SetStyle(tcell.StyleDefault.Background(color))
	case *primitives.Pages:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
	case *tview.InputField:
		_, c, _ = obj.GetFieldStyle().Decompose()
		obj.SetFieldBackgroundColor(color)
		obj.SetPlaceholderStyle(obj.GetFieldStyle().Background(color))
	case *tview.DropDown:
		c = obj.GetBackgroundColor()
		obj.SetBackgroundColor(color)
	}
	return c
}

func allowDeepFocus(obj tview.Primitive) bool {
	switch obj.(type) {
	case *tview.Grid, *tview.Flex, *primitives.Pages:
		return true
	default:
		return false
	}
}

package keyboard

import (
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/tui"
	"bilibili-ticket-go/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type selectItem struct {
	where         int
	obj           tview.Primitive
	previousColor tcell.Color
}

type KeyboardCaptureInstance struct {
	stack    *models.Stack[selectItem]
	app      *tview.Application
	root     *tview.Flex
	selected selectItem
}

var logger = utils.GetLogger(global.GetLogger(), "keyboard", nil)

const highlightColor = tcell.ColorForestGreen

func NewKeyboardCaptureInstance(app *tview.Application, root *tview.Flex) *KeyboardCaptureInstance {
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
	return k.selected.obj != nil && k.selected.obj != k.root
}

func (k *KeyboardCaptureInstance) Reset() {
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
	if event.Key() == tcell.KeyEscape {
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
			setColor(previous.obj, previous.previousColor)
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
			case *tui.Pages:
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
	switch o := box.obj.(type) {
	case *tview.Flex:
		var nextItemID = k.selected.where + 1
		var cnt = o.GetItemCount()
		var nxtObj tview.Primitive
		for i := 0; i < cnt; i++ {
			var index = nextItemID + i
			if index >= cnt {
				index -= cnt
			}
			obj := o.GetItem(index)
			if selectable(obj) {
				nxtObj = obj
				nextItemID = index
				break
			}
		}
		if nxtObj == nil {
			return
		}
		setColor(k.selected.obj, k.selected.previousColor)
		c := setColor(nxtObj, highlightColor)

		k.selected = selectItem{
			where:         nextItemID,
			obj:           nxtObj,
			previousColor: c,
		}
	case *tui.Pages:
		cur := o.GetCurrentPage()
		k.switchToNextItem(selectItem{
			where: -1,
			obj:   cur,
		})
	}
}

func selectable(obj tview.Primitive) bool {
	switch obj.(type) {
	case *tview.List, *tview.Box, *tview.Flex, *tview.Grid, *tview.Button, *tui.Pages, *tview.InputField:
		return true
	default:
		return false
	}
}
func setColor(primitive tview.Primitive, color tcell.Color) tcell.Color {
	var c tcell.Color
	switch obj := primitive.(type) {
	case *tview.List:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
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
	case *tui.Pages:
		c = obj.GetBorderColor()
		obj.SetBorderColor(color)
	case *tview.InputField:
		_, c, _ = obj.GetFieldStyle().Decompose()
		obj.SetFieldBackgroundColor(color)
		obj.SetPlaceholderStyle(obj.GetFieldStyle().Background(color))
	}
	return c
}

func allowDeepFocus(obj tview.Primitive) bool {
	switch obj.(type) {
	case *tview.Grid, *tview.Flex, *tui.Pages:
		return true
	default:
		return false
	}
}

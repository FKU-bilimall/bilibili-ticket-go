package k

import (
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type selectItem struct {
	where int
	obj   tview.Primitive
}

type KeyboardCaptureInstance struct {
	stack    *models.Stack[selectItem]
	app      *tview.Application
	root     *tview.Flex
	selected selectItem
}

var logger = utils.GetLogger(global.GetLogger(), "keyboard", nil)

func NewKeyboardCaptureInstance(app *tview.Application, root *tview.Flex) *KeyboardCaptureInstance {
	app.SetFocus(root)
	st := (&models.Stack[selectItem]{}).New()
	st.Push(selectItem{
		where: -1,
		obj:   root,
	})
	return &KeyboardCaptureInstance{
		stack:    st,
		app:      app,
		root:     root,
		selected: selectItem{where: -1, obj: root},
	}
}
func (k *KeyboardCaptureInstance) Reset() {
	k.stack.Clear()
	k.stack.Push(selectItem{
		where: 0,
		obj:   k.root,
	})
	k.app.SetFocus(k.root)
}

func (k *KeyboardCaptureInstance) InputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		if k.stack.Size() > 1 {
			current := k.stack.Top()
			k.stack.Pop()
			previous := k.stack.Top()
			switchHighlighted(k.selected.obj, current.obj)
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
		logger.Trace("Tab pressed, current item: ", item.obj, " at position: ", item.where)

		switch o := item.obj.(type) {
		case *tview.Flex:
			var nextItemID = k.selected.where + 1
			if nextItemID >= o.GetItemCount() {
				nextItemID = 0
			}
			nxtObj := o.GetItem(nextItemID)
			switchHighlighted(k.selected.obj, nxtObj)
			k.selected = selectItem{
				where: nextItemID,
				obj:   nxtObj,
			}
		case *tview.Pages:

		}
	}

	if event.Key() == tcell.KeyEnter {
		if k.selected.obj != nil && k.selected.obj != k.stack.Top().obj {
			k.app.SetFocus(k.selected.obj)
			switch obj := k.selected.obj.(type) {
			case *tview.Flex:
				if k.selected.where == -1 {
					break
				}
				k.stack.Push(k.selected)
				switchHighlighted(nil, obj)
				k.selected = selectItem{
					where: -1,
					obj:   nil,
				}
			}
		}
	}
	return event
}

func switchHighlighted(oldObj tview.Primitive, newObj tview.Primitive) {
	colorSwitch(oldObj, tcell.ColorWhite)
	colorSwitch(newObj, tcell.ColorForestGreen)
}

func colorSwitch(primitive tview.Primitive, color tcell.Color) {
	switch old := primitive.(type) {
	case *tview.List:
		old.SetBorderColor(color)
	case *tview.Box:
		old.SetBorderColor(color)
	case *tview.Flex:
		old.SetBorderColor(color)
	case *tview.Grid:
		old.SetBorderColor(color)
	case *tview.Button:
		old.SetStyle(tcell.StyleDefault.Background(color))
	case *tview.Pages:
		old.SetBorderColor(color)
	}
}

func allowDeepFocus(obj tview.Primitive) bool {
	switch obj.(type) {
	case *tview.Grid, *tview.Flex, *tview.Pages:
		return true
	default:
		return false
	}
}

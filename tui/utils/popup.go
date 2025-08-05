package utils

import (
	"bilibili-ticket-go/keyboard"
	"bilibili-ticket-go/tui/primitives"

	"github.com/rivo/tview"
)

// PopupModal displays a modal popup with a message and buttons.
// The buttons are defined in the buttonFuncMap, where the key is the button label
// and the value is a function that returns a boolean indicating whether to close the modal.
// If the function returns false, the modal will remain open.
// The modal is added to the provided mount, which is a tview.Pages instance.
func PopupModal(message string, mount *primitives.Pages, buttonFuncMap map[string]func() bool, instance *keyboard.KeyboardCaptureInstance) {
	instance.SetIsOpenModel(true)
	keys := make([]string, 0, len(buttonFuncMap))
	for k := range buttonFuncMap {
		keys = append(keys, k)
	}
	modal := tview.NewModal().SetText(message).AddButtons(keys)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		var flag = true
		if fn, ok := buttonFuncMap[buttonLabel]; ok {
			flag = fn()
		}
		if flag {
			instance.SetIsOpenModel(false)
			mount.RemovePage("modal")
			mount.SetOthersClickableStat(0)
		}
	})
	mount.AddPage("modal", modal, true, true)
	mount.SetOthersClickableStat(1)
}

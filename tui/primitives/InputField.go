package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputField struct {
	disabledStyle tcell.Style
	fieldStyle    tcell.Style
	disabled      bool
	*tview.InputField
}

func NewInputField() *InputField {
	in := tview.NewInputField()
	return &InputField{
		InputField:    in,
		disabledStyle: tcell.StyleDefault.Background(in.GetBackgroundColor()).Foreground(tview.Styles.SecondaryTextColor),
		fieldStyle:    tcell.StyleDefault.Background(tview.Styles.ContrastBackgroundColor).Foreground(tview.Styles.PrimaryTextColor),
	}
}

func (i *InputField) SetDisabled(disabled bool) *InputField {
	i.InputField.SetDisabled(disabled)
	if disabled {
		i.InputField.SetFieldStyle(i.disabledStyle)
		i.InputField.SetPlaceholderStyle(i.disabledStyle)
	} else {
		i.InputField.SetFieldStyle(i.fieldStyle)
		i.InputField.SetPlaceholderStyle(i.fieldStyle)
	}
	i.disabled = disabled
	return i
}

func (i *InputField) SetFieldStyle(style tcell.Style) *InputField {
	i.fieldStyle = style
	if !i.disabled {
		i.InputField.SetFieldStyle(i.fieldStyle)
		i.InputField.SetPlaceholderStyle(i.fieldStyle)
	}
	return i
}

func (i *InputField) SetDisabledStyle(style tcell.Style) *InputField {
	i.disabledStyle = style
	if i.disabled {
		i.InputField.SetFieldStyle(i.disabledStyle)
		i.InputField.SetPlaceholderStyle(i.disabledStyle)
	}
	return i
}

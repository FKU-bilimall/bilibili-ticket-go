package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DropDown struct {
	fieldStyle tcell.Style
	*tview.DropDown
}

func NewDropDown() *DropDown {
	d := &DropDown{
		fieldStyle: tcell.StyleDefault.Background(tview.Styles.ContrastBackgroundColor).Foreground(tview.Styles.PrimaryTextColor),
		DropDown:   tview.NewDropDown(),
	}
	return d
}

// SetFieldBackgroundColor sets the background color of the selected field.
// This also overrides the prefix background color.
func (d *DropDown) SetFieldBackgroundColor(color tcell.Color) *DropDown {
	d.fieldStyle = d.fieldStyle.Background(color)
	d.DropDown.SetBackgroundColor(color)
	return d
}

// SetFieldTextColor sets the text color of the options area.
func (d *DropDown) SetFieldTextColor(color tcell.Color) *DropDown {
	d.fieldStyle = d.fieldStyle.Foreground(color)
	d.DropDown.SetFieldTextColor(color)
	return d
}

// SetFieldStyle sets the style of the options area.
func (d *DropDown) SetFieldStyle(style tcell.Style) *DropDown {
	d.fieldStyle = style
	d.DropDown.SetFieldStyle(style)
	return d
}

func (d *DropDown) GetFieldStyle() tcell.Style {
	return d.fieldStyle
}

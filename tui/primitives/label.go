package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Label struct {
	text  string
	color tcell.Color
	*tview.Box
}

func NewLabel(text string) *Label {
	box := tview.NewBox()
	return &Label{
		Box:   box,
		color: tview.Styles.PrimaryTextColor,
		text:  text,
	}
}

func (l *Label) SetText(text string) *Label {
	l.text = text
	return l
}

func (l *Label) SetColor(color tcell.Color) *Label {
	l.color = color
	return l
}

func (l *Label) Draw(screen tcell.Screen) {
	x, y, width, height := l.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}
	l.Box.DrawForSubclass(screen, l)
	tview.Print(screen, l.text, x, y, rightLimit-x, tview.AlignLeft, l.color)
	return
}

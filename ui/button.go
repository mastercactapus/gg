package ui

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Button struct {
	X, Y        int
	Width       int
	Text        string
	Enabled     bool
	OnClickFunc func(x, y int)
}

func (b *Button) OnClick(x, y int) {
	if !b.Enabled {
		return
	}
	if b.OnClickFunc == nil {
		return
	}
	b.OnClickFunc(x, y)
}

func (b *Button) Draw(r Renderer) {
	w := b.Width
	minW := len(b.Text) + 2
	if minW > w {
		w = minW
	}
	sw, sh := r.Size()
	x, y, w, _ := StandardSize(b.X, b.Y, w, 2, sw, sh)

	space := strings.Repeat(" ", (w-len(b.Text))/2)

	text := "[" + space + b.Text + space + "]"

	var fg, bg termbox.Attribute
	if !b.Enabled {
		fg = termbox.ColorBlack
	} else {
		bg = termbox.ColorBlue
		fg = termbox.ColorWhite
	}

	putRunesA(r, x, y, []rune(text), fg, bg)
}

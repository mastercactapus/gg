package ui

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type DblButton struct {
	X, Y        int
	Width       int
	Text        string
	Enabled     bool
	OnClickFunc func(x, y int)
}

func (b *DblButton) OnClick(x, y int) {
	if !b.Enabled {
		return
	}
	if b.OnClickFunc == nil {
		return
	}
	b.OnClickFunc(x, y)
}

func (b *DblButton) Draw(r Renderer) {
	w := b.Width
	minW := len([]rune(b.Text)) + 2
	if minW > w {
		w = minW
	}

	sw, sh := r.Size()
	x, y, w, _ := StandardSize(b.X, b.Y, w, 3, sw, sh)

	text1 := "┌" + strings.Repeat(" ", w-2) + "┐"
	text2 := "│" + fitText(b.Text, w-2) + "│"
	text3 := "└" + strings.Repeat(" ", w-2) + "┘"

	var fg, bg termbox.Attribute
	if !b.Enabled {
		fg = termbox.ColorBlack
	} else {
		bg = termbox.ColorBlue
		fg = termbox.ColorWhite
	}

	putRunesA(r, x, y, []rune(text1), fg, bg)
	putRunesA(r, x, y+1, []rune(text2), fg, bg)
	putRunesA(r, x, y+2, []rune(text3), fg, bg)

}

func fitText(s string, w int) string {
	n := w - len([]rune(s))
	if n <= 0 {
		return s
	}

	n = n / 2
	spc := strings.Repeat(" ", n)
	s = spc + s + spc
	if len([]rune(s)) < w {
		s += " "
	}
	return s
}

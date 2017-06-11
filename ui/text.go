package ui

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Text struct {
	X, Y   int
	FG, BG termbox.Attribute
	Lines  []string
}

func (t Text) Draw(r Renderer) {
	w, _ := r.Size()
	space := strings.Repeat(" ", w)
	for i, l := range t.Lines {
		rs := []rune(l + space)
		putRunesA(r, t.X, t.Y+i, rs[:w], t.FG, t.BG)
	}
}

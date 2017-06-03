package ui

import "strings"

type Text struct {
	X, Y  int
	Lines []string
}

func (t Text) Draw(r Renderer) {
	w, _ := r.Size()
	space := strings.Repeat(" ", w)
	for i, l := range t.Lines {
		rs := []rune(l + space)
		putRunes(r, t.X, t.Y+i, rs[:w])
	}
}

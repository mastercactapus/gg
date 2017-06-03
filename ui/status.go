package ui

import (
	"fmt"

	"github.com/mastercactapus/gg/grbl"
	termbox "github.com/nsf/termbox-go"
)

type Status struct {
	X, Y int
	*grbl.Status
}

func (s *Status) Draw(r Renderer) {
	putRunes(r, 0, 0, []rune("hi"))
	if s.Status == nil || len(s.Status.MPos) == 0 {
		return
	}
	xStr := fmt.Sprintf("% 4.3f", s.MPos[0])

	putRunes(r, s.X, s.Y, []rune("Machine Coords   X "))
	putRunesA(r, s.X+19, s.Y, []rune(xStr), termbox.ColorYellow, 0)
}

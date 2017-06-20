package ui

import (
	"fmt"

	"github.com/mastercactapus/gg/grbl"
	termbox "github.com/nsf/termbox-go"
)

type Status struct {
	X, Y int
	*grbl.Status
	*grbl.Settings
}

func (s *Status) Draw(r Renderer) {
	if s.Status == nil || len(s.Status.MPos) == 0 {
		return
	}

	printCoords := func(x, y, w int, label string, c []float64) {
		var coords []string
		for i := 0; i < 3; i++ {
			if len(c) <= i {
				coords = append(coords, fmt.Sprintf(" % 8s", "-"))
			} else {
				coords = append(coords, fmt.Sprintf(" % 8.3f", c[i]))
			}
		}
		putRunes(r, x, y, []rune(label))
		putRunes(r, x, y+1, []rune("Coords"))
		x += w
		putRunesA(r, x, y, []rune{'X'}, termbox.ColorWhite, termbox.ColorBlack)
		putRunesA(r, x+1, y, []rune(coords[0]), termbox.ColorYellow, termbox.ColorBlack)
		y++
		putRunesA(r, x, y, []rune{'Y'}, termbox.ColorWhite, termbox.ColorBlack)
		putRunesA(r, x+1, y, []rune(coords[1]), termbox.ColorYellow, termbox.ColorBlack)
		y++
		putRunesA(r, x, y, []rune{'Z'}, termbox.ColorWhite, termbox.ColorBlack)
		putRunesA(r, x+1, y, []rune(coords[2]), termbox.ColorYellow, termbox.ColorBlack)
	}

	printCoords(s.X, s.Y, 10, "Work", s.WPos)
	printCoords(s.X, s.Y+4, 10, "Machine", s.MPos)
	if s.Settings != nil {
		printCoords(s.X, s.Y+8, 10, "Max", []float64{s.MaxTravel.X.Millimeters(), s.MaxTravel.Y.Millimeters(), s.MaxTravel.Z.Millimeters()})
	}
}

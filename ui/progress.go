package ui

import (
	"strconv"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Progress struct {
	Parent *Rect

	X, Y  int
	Width int

	Value int
	Max   int
	Title string
}

var progRunes = [...]rune{
	' ',
	'\U0000258F',
	'\U0000258E',
	'\U0000258D',
	'\U0000258C',
	'\U0000258B',
	'\U0000258A',
	'\U00002589',
	'\U00002588',
}

func (p *Progress) Draw(r Renderer) {
	sw, sh := r.Size()
	x, y, w, h := StandardSize(p.X, p.Y, p.Width, 2, sw, sh)
	if h < 1 {
		return
	}
	val := p.Value
	if val < 0 {
		val = 0
	}

	title := p.Title + " " + strconv.FormatFloat(float64(val*100)/float64(p.Max), 'f', 2, 64) + "%" + strings.Repeat(" ", w)
	putRunes(r, x, y, []rune(title))
	if h < 2 {
		return
	}

	bar := make([]rune, w)

	val = (val * 8 * w) / p.Max

	rem := val % 8
	val = val / 8

	for i := range bar {
		if i+1 < val {
			bar[i] = progRunes[8]
		} else if i+1 > val {
			bar[i] = progRunes[0]
		} else {
			bar[i] = progRunes[rem]
		}
	}

	putRunesA(r, x, y+1, bar, termbox.ColorRed, termbox.ColorWhite)
}

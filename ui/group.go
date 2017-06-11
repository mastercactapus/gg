package ui

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Group struct {
	Title    string
	Height   int
	Width    int
	X, Y     int
	Clear    bool
	NoBorder bool
	Controls []Control

	rendered []renderedControl
}

func (g *Group) OnClick(x, y int) {
	for _, c := range g.rendered {
		cc, ok := c.c.(Clickable)
		if !ok {
			continue
		}
		if !c.r.Contains(x+g.X, y+g.Y) {
			continue
		}
		cc.OnClick(c.r.Translate(x+g.X, y+g.Y))
	}
}

func (g *Group) Draw(r Renderer) {
	sw, sh := r.Size()
	x, y, w, h := StandardSize(g.X, g.Y, g.Width, g.Height, sw, sh)
	yMx := h + y
	xMx := w + x
	if w < 3 || h < 3 {
		return
	}
	var header string

	if len(g.Title) == 0 {
		header = strings.Repeat("─", w-2)
	} else if w == 3 {
		header = string(g.Title[0])
	} else if w-2 == len(g.Title) {
		header = g.Title
	} else if w-2 < len(g.Title) {
		header = g.Title[:w-3] + "…"
	} else if w-3 == len(g.Title) {
		header = g.Title + "─"
	} else if w-4 == len(g.Title) {
		header = "─" + g.Title + "─"
	} else if w-5 == len(g.Title) {
		header = "─" + g.Title + "──"
	} else {
		header = "─ " + g.Title + " " + strings.Repeat("─", w-5-len(g.Title))
	}
	if !g.NoBorder {
		putRunes(r, x, y, []rune("┌"+header+"┐"))
		putRunes(r, x, yMx-1, []rune("└"+strings.Repeat("─", w-2)+"┘"))

		for row := y + 1; row < yMx-1; row++ {
			r.SetCell(x, row, '│', 0, 0)
			r.SetCell(xMx-1, row, '│', 0, 0)
		}
	} else {
		xMx++
		x--
		y--
		yMx++
	}

	if g.Clear {
		for col := x + 1; col < xMx-1; col++ {
			for row := y + 1; row < yMx-1; row++ {
				r.SetCell(col, row, ' ', 0, 0)
			}
		}
	}

	g.rendered = g.rendered[:0]
	for _, c := range g.Controls {
		if c == nil {
			continue
		}
		g.rendered = append(g.rendered, renderedControl{
			c: c,
			r: r.RenderChild(Rect{Left: x + 1, Top: y + 1, Right: xMx - 2, Bottom: yMx - 1}, c),
		})
	}
}

func putRunes(s CellSetter, x, y int, r []rune) {
	putRunesA(s, x, y, r, 0, 0)

}
func putRunesA(s CellSetter, x, y int, r []rune, fg, bg termbox.Attribute) {
	for i, r := range r {
		s.SetCell(x+i, y, r, fg, bg)
	}
}

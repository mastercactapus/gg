package ui

import termbox "github.com/nsf/termbox-go"

type boundedRenderer struct {
	Screen
	x, y, xm, ym int

	r Rect
}

func newBoundedRenderer(s Screen, x, y, xm, ym int) *boundedRenderer {
	return &boundedRenderer{
		Screen: s,

		x:  x,
		y:  y,
		xm: xm,
		ym: ym,
		r:  Rect{Left: xm, Top: ym},
	}
}

func (b *boundedRenderer) Size() (int, int) {
	return b.xm - b.x, b.ym - b.y
}
func (b *boundedRenderer) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	if x < 0 {
		b.r.Left = b.x
		return
	}
	if y < 0 {
		b.r.Top = b.y
		return
	}
	x += b.x
	y += b.y
	if x > b.xm {
		b.r.Right = b.xm
		return
	}
	if y > b.ym {
		b.r.Bottom = b.ym
		return
	}

	if x < b.r.Left {
		b.r.Left = x
	}
	if x > b.r.Right {
		b.r.Right = x
	}
	if y < b.r.Top {
		b.r.Top = y
	}
	if y > b.r.Bottom {
		b.r.Bottom = y
	}

	b.Screen.SetCell(x, y, ch, fg, bg)
}
func (b *boundedRenderer) RenderChild(x, y, xm, ym int, c Control) Rect {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if xm < 0 {
		xm = 0
	}
	if ym < 0 {
		ym = 0
	}

	x += b.x
	y += b.y
	xm += b.x
	ym += b.y

	if xm > b.xm {
		xm = b.xm
	}
	if ym > b.ym {
		ym = b.ym
	}
	if x > xm {
		x = xm
	}
	if y > ym {
		y = ym
	}

	br := newBoundedRenderer(b.Screen, x, y, xm, ym)
	c.Draw(br)
	return br.r
}

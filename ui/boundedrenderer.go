package ui

import termbox "github.com/nsf/termbox-go"

type boundedRenderer struct {
	Screen
	x, y, xm, ym int

	b        Rect
	drawRect Rect
}

func newBoundedRenderer(s Screen, bounds Rect) *boundedRenderer {
	sw, sh := s.Size()
	if bounds.Right > sw {
		bounds.Right = sw
	}
	if bounds.Left < 0 {
		bounds.Left = 0
	}
	if bounds.Top < 0 {
		bounds.Top = 0
	}
	if bounds.Bottom > sh {
		bounds.Bottom = sh
	}
	if bounds.Left > bounds.Right {
		panic("bounds.Left excedes bounds.Right")
	}
	if bounds.Top > bounds.Bottom {
		panic("bounds.Top excedes bounds.Bottom")
	}

	return &boundedRenderer{
		Screen:   s,
		b:        bounds,
		drawRect: Rect{Left: bounds.Right + 1, Top: bounds.Bottom + 1},
	}
}

func (r *boundedRenderer) Size() (int, int) {
	return r.b.Right - r.b.Left, r.b.Bottom - r.b.Top
}
func (r *boundedRenderer) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	x += r.b.Left
	y += r.b.Top

	var oob bool
	if x < r.b.Left {
		x = r.b.Left
		oob = true
	}
	if x > r.b.Right {
		x = r.b.Right
		oob = true
	}
	if y < r.b.Top {
		y = r.b.Top
		oob = true
	}
	if y > r.b.Bottom {
		y = r.b.Bottom
		oob = true
	}

	if x < r.drawRect.Left {
		r.drawRect.Left = x
	}
	if x > r.drawRect.Right {
		r.drawRect.Right = x
	}
	if y < r.drawRect.Top {
		r.drawRect.Top = y
	}
	if y > r.drawRect.Bottom {
		r.drawRect.Bottom = y
	}
	if oob {
		return
	}

	r.Screen.SetCell(x, y, ch, fg, bg)
}
func (r *boundedRenderer) RenderChild(bounds Rect, c Control) Rect {
	bounds.Left += r.b.Left
	bounds.Right += r.b.Left
	bounds.Top += r.b.Top
	bounds.Bottom += r.b.Top

	br := newBoundedRenderer(r.Screen, bounds)
	c.Draw(br)

	br.drawRect.Left -= r.b.Left
	br.drawRect.Right -= r.b.Left
	br.drawRect.Top -= r.b.Top
	br.drawRect.Bottom -= r.b.Top
	return br.drawRect
}

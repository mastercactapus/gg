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
	w, h := r.Size()
	var oob bool
	if x < 0 {
		x = 0
		oob = true
	} else if x > w {
		x = w
		oob = true
	}

	if y < 0 {
		y = 0
		oob = true
	} else if y > h {
		y = h
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

	r.Screen.SetCell(x+r.b.Left, y+r.b.Top, ch, fg, bg)
}
func (r *boundedRenderer) RenderChild(bounds Rect, c Control) Rect {
	br := newBoundedRenderer(r, bounds)
	c.Draw(br)

	br.drawRect.Left += bounds.Left
	br.drawRect.Right += bounds.Left
	br.drawRect.Top += bounds.Top
	br.drawRect.Bottom += bounds.Top

	return br.drawRect
}

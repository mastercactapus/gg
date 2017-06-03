package ui

type Rect struct{ Top, Bottom, Left, Right int }

func (r Rect) Contains(x, y int) bool {
	return x >= r.Left && x <= r.Right && y >= r.Top && y <= r.Bottom
}
func (r Rect) Translate(x, y int) (int, int) {
	return r.Left + x, r.Top + y
}

func StandardSize(x, y, w, h, sw, sh int) (tx, ty, tw, th int) {
	if w <= 0 {
		w += sw - x
	}
	if h <= 0 {
		h += sh - y
	}

	if sw < x+w {
		w = sw - x
	}
	if sh < y+h {
		h = sh - y
	}

	return x, y, w, h
}

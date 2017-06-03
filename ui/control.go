package ui

import termbox "github.com/nsf/termbox-go"

type DrawFunc func(Renderer)

type RenderedControl struct {
	Control
	Rect
}
type CellSetterFunc func(x, y int, ch rune, fg, bg termbox.Attribute)

func (f CellSetterFunc) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	f(x, y, ch, fg, bg)
}

func (f DrawFunc) Draw(r Renderer) {
	f(r)
}

type Control interface {
	Draw(Renderer)
}

type CellSetter interface {
	SetCell(x, y int, ch rune, fg, bg termbox.Attribute)
}

type Clickable interface {
	OnClick(x, y int)
}
type Scrollable interface {
	OnScroll(n int)
}

type BoundedCellSetter struct {
	CellSetter
	X, Y   int
	Width  int
	Height int

	xMax, yMax int
}

func (b *BoundedCellSetter) Flow() {
	b.Width -= b.xMax - b.X
	b.X = b.xMax

	b.Height -= b.yMax - b.Y
	b.Y = b.yMax
}

func (b *BoundedCellSetter) Size() (int, int) {
	return b.Width, b.Height
}
func (b *BoundedCellSetter) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	if x < 0 || y < 0 {
		return
	}
	if x > b.Width || y > b.Height {
		return
	}
	x += b.X
	y += b.Y

	if y > b.yMax {
		b.yMax = y
		b.xMax = x + 1
	} else if x+1 > b.xMax {
		b.xMax = x + 1
	}

	b.CellSetter.SetCell(x, y, ch, fg, bg)
}

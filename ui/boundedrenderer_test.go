package ui

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
)

// testS implements the Screen interface for testing.
type testS struct {
	SizeW, SizeH int

	SetX, SetY int
}

func (r *testS) Size() (int, int)                                    { return r.SizeW, r.SizeH }
func (r *testS) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) { r.SetX, r.SetY = x, y }

func TestBoundedRenderer_Size(t *testing.T) {
	ts := &testS{
		SizeW: 10, SizeH: 5,
	}

	test := func(x, y, xm, ym, expW, expH int) {
		t.Run("SetCell", func(t *testing.T) {
			b := newBoundedRenderer(ts, x, y, xm, ym)
			w, h := b.Size()
			if w != expW {
				t.Errorf("Width = %d; want %d", w, expW)
			}
			if h != expH {
				t.Errorf("Height = %d; want %d", h, expH)
			}
		})
	}

	data := [][]int{
		{0, 0, 1, 2, 1, 2},
		{1, 0, 1, 2, 0, 2},
		{4, 0, 1, 2, 0, 2},

		{0, 5, 1, 2, 1, 0},
		{0, 1, 1, 2, 1, 1},

		// screen size bounds
		{8, 0, 12, 0, 2, 0},
		{0, 4, 0, 8, 0, 1},
	}

	for _, n := range data {
		test(n[0], n[1], n[2], n[3], n[4], n[5])
	}

}

func TestBoundedRenderer(t *testing.T) {
	ts := &testS{
		SizeW: 10, SizeH: 5,
	}
	b := newBoundedRenderer(ts, 1, 2, 3, 4)

	test := func(x, y, expX, expY int) {
		t.Run("SetCell", func(t *testing.T) {
			b.SetCell(x, y, ' ', 0, 0)
			if ts.SetX != expX {
				t.Errorf("SetX = %d; want %d", ts.SetX, expX)
			}
			if ts.SetY != expY {
				t.Errorf("SetY = %d; want %d", ts.SetY, expY)
			}
		})
	}

	data := [][]int{
		{0, 0, 1, 2},
		{1, 0, 2, 2},
		{2, 0, 3, 2},
		{3, 0, 3, 2},
		{4, 0, 3, 2},

		{0, 1, 1, 3},
		{0, 2, 1, 4},
		{0, 3, 1, 4},
	}

	for _, n := range data {
		test(n[0], n[1], n[2], n[3])
	}
}

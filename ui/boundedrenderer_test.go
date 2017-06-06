package ui

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
)

type pair struct{ x, y int }

// testS implements the Screen interface for testing.
type testS struct {
	SizeW, SizeH int

	SetX, SetY int
}

func (r *testS) Reset()                                              { r.SetX, r.SetY = -1, -1 }
func (r *testS) Size() (int, int)                                    { return r.SizeW, r.SizeH }
func (r *testS) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) { r.SetX, r.SetY = x, y }

func TestBoundedRenderer_Size(t *testing.T) {
	ts := &testS{
		SizeW: 10, SizeH: 5,
	}

	test := func(r Rect, expect pair) {
		t.Run("", func(t *testing.T) {
			ts.Reset()
			b := newBoundedRenderer(ts, r)
			w, h := b.Size()
			if w != expect.x {
				t.Errorf("Width = %d; want %d", w, expect.x)
			}
			if h != expect.y {
				t.Errorf("Height = %d; want %d", h, expect.y)
			}
		})
	}

	data := []struct {
		Set    Rect
		Expect pair
	}{
		{Set: Rect{0, 0, 1, 2}, Expect: pair{1, 2}},
		{Set: Rect{1, 0, 1, 2}, Expect: pair{0, 2}},
		{Set: Rect{4, 0, 4, 2}, Expect: pair{0, 2}},

		{Set: Rect{0, 5, 1, 5}, Expect: pair{1, 0}},
		{Set: Rect{0, 1, 1, 2}, Expect: pair{1, 1}},

		// screen size bounds
		{Set: Rect{8, 0, 12, 0}, Expect: pair{2, 0}},
		{Set: Rect{0, 4, 0, 8}, Expect: pair{0, 1}},
	}

	for _, n := range data {
		test(n.Set, n.Expect)
	}

}

func TestBoundedRenderer_SetCell(t *testing.T) {
	ts := &testS{
		SizeW: 10, SizeH: 5,
	}
	b := newBoundedRenderer(ts, Rect{Left: 1, Top: 2, Right: 3, Bottom: 4})

	test := func(set, expect pair) {
		t.Run("", func(t *testing.T) {
			ts.Reset()
			b.SetCell(set.x, set.y, ' ', 0, 0)
			if ts.SetX != expect.x {
				t.Errorf("SetX = %d; want %d", ts.SetX, expect.x)
			}
			if ts.SetY != expect.y {
				t.Errorf("SetY = %d; want %d", ts.SetY, expect.y)
			}
		})
	}

	data := []struct{ Set, Expect pair }{
		{Set: pair{0, 0}, Expect: pair{1, 2}},
		{Set: pair{1, 0}, Expect: pair{2, 2}},
		{Set: pair{2, 0}, Expect: pair{3, 2}},

		// out of bounds
		{Set: pair{3, 0}, Expect: pair{-1, -1}},
		{Set: pair{4, 0}, Expect: pair{-1, -1}},

		{Set: pair{0, 1}, Expect: pair{1, 3}},
		{Set: pair{0, 2}, Expect: pair{1, 4}},

		// out of bounds
		{Set: pair{0, 3}, Expect: pair{-1, -1}},
	}

	for _, n := range data {
		test(n.Set, n.Expect)
	}
}

func TestBoundedRenderer_Full(t *testing.T) {
	test := func(screen pair, bounds Rect, set []pair, expectChild Rect, expectSize pair) {
		t.Run("", func(t *testing.T) {
			ts := &testS{SizeW: screen.x, SizeH: screen.y}
			ts.Reset()
			b := newBoundedRenderer(ts, Rect{Top: 3, Right: screen.x, Bottom: screen.y})
			var sw, sh int
			result := b.RenderChild(bounds, DrawFunc(func(r Renderer) {
				sw, sh = r.Size()
				for _, p := range set {
					r.SetCell(p.x, p.y, ' ', 0, 0)
				}
			}))

			if sw != expectSize.x {
				t.Errorf("Size.Width = %d; want %d", sw, expectSize.x)
			}
			if sh != expectSize.y {
				t.Errorf("Size.Height = %d; want %d", sh, expectSize.y)
			}

			if result.Left != expectChild.Left {
				t.Errorf("Child.Left = %d; want %d", result.Left, expectChild.Left)
			}
			if result.Top != expectChild.Top {
				t.Errorf("Child.Top = %d; want %d", result.Top, expectChild.Top)
			}
			if result.Right != expectChild.Right {
				t.Errorf("Child.Right = %d; want %d", result.Right, expectChild.Right)
			}
			if result.Bottom != expectChild.Bottom {
				t.Errorf("Child.Bottom = %d; want %d", result.Bottom, expectChild.Bottom)
			}

		})
	}

	data := []struct {
		Screen      pair
		Bounds      Rect
		Set         []pair
		ExpectSize  pair
		ExpectChild Rect
	}{
		{Screen: pair{80, 40}, Bounds: Rect{1, 1, 39, 39}, Set: []pair{{1, 20}, {37, 37}}, ExpectSize: pair{38, 36}, ExpectChild: Rect{2, 21, 38, 37}},
	}

	for _, d := range data {
		test(d.Screen, d.Bounds, d.Set, d.ExpectChild, d.ExpectSize)
	}
}

func TestBoundedRenderer_RenderChild(t *testing.T) {
	ts := &testS{
		SizeW: 10, SizeH: 15,
	}

	b := newBoundedRenderer(ts, Rect{Left: 1, Top: 1, Right: 9, Bottom: 14})
	testBounds := func(r Rect, set, expect pair) {
		t.Run("Bounds", func(t *testing.T) {
			ts.Reset()
			b.RenderChild(r, DrawFunc(func(r Renderer) {
				r.SetCell(set.x, set.y, ' ', 0, 0)
			}))

			if ts.SetX != expect.x {
				t.Errorf("SetX = %d; want %d", ts.SetX, expect.x)
			}
			if ts.SetY != expect.y {
				t.Errorf("SetY = %d; want %d", ts.SetY, expect.y)
			}
		})
	}

	bounds := []struct {
		Bounds Rect
		Set    pair
		Expect pair
	}{
		{Bounds: Rect{1, 2, 3, 4}, Set: pair{1, 2}, Expect: pair{3, 5}},

		// out of bounds, should not get written
		{Bounds: Rect{3, 4, 3, 4}, Set: pair{40, 5}, Expect: pair{-1, -1}},
	}

	for _, b := range bounds {
		testBounds(b.Bounds, b.Set, b.Expect)
	}

	testRect := func(r Rect, set []pair, expect Rect) {
		t.Run("Rect", func(t *testing.T) {
			ts.Reset()
			result := b.RenderChild(r, DrawFunc(func(r Renderer) {
				for _, p := range set {
					r.SetCell(p.x, p.y, ' ', 0, 0)
				}
			}))
			if result.Left != expect.Left {
				t.Errorf("Left = %d; want %d", result.Left, expect.Left)
			}
			if result.Top != expect.Top {
				t.Errorf("Top = %d; want %d", result.Top, expect.Top)
			}
			if result.Right != expect.Right {
				t.Errorf("Right = %d; want %d", result.Right, expect.Right)
			}
			if result.Bottom != expect.Bottom {
				t.Errorf("Bottom = %d; want %d", result.Bottom, expect.Bottom)
			}
		})
	}

	rect := []struct {
		r      Rect
		Set    []pair
		Expect Rect
	}{
		{r: Rect{5, 10, 9, 14}, Set: []pair{{0, 0}, {20, 20}}, Expect: Rect{5, 10, 8, 13}},
		{r: Rect{5, 10, 9, 14}, Set: []pair{{0, 0}, {1, 1}}, Expect: Rect{5, 10, 6, 11}},
		{r: Rect{5, 10, 9, 14}, Set: []pair{{0, 0}}, Expect: Rect{5, 10, 5, 10}},
	}

	for _, b := range rect {
		testRect(b.r, b.Set, b.Expect)
	}
}

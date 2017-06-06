package ui

import "testing"

func TestRect_Translate(t *testing.T) {
	r := Rect{4, 5, 6, 7}

	x, y := r.Translate(6, 7)

	if x != 2 {
		t.Errorf("x = %d; want %d", x, 2)
	}
	if y != 2 {
		t.Errorf("y = %d; want %d", y, 2)
	}
}

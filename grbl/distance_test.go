package grbl

import "testing"

func TestDistance(t *testing.T) {
	test := func(name string, act, exp float64) {
		t.Run(name, func(t *testing.T) {
			if act != exp {
				t.Errorf("%s = %f; want %f", name, act, exp)
			}
		})
	}

	data := []struct {
		name     string
		act, exp float64
	}{
		{"Foot.Millimeters()", Foot.Millimeters(), 304.8},
		{"Inch.Millimeters()", Inch.Millimeters(), 25.4},
		{"Meter.Inches()", (254 * Meter).Inches(), 10000},
	}

	for _, d := range data {
		test(d.name, d.act, d.exp)
	}
}

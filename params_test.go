package gg

import "testing"

func TestParseUnitString(t *testing.T) {

	tests := []struct {
		name, val string
		exp       float64
	}{
		{"mm", "1mm", 1},
		{"mm", "6.3mm", 6.3},
		{"cm", "1cm", 10},
		{"inch", "1in", 25.4},
		{"inch abbrv", "1\"", 25.4},
		{"inch full", "1 inch", 25.4},
		{"inch full", "1 inch.", 25.4},
		{"inch full", "1 in.", 25.4},
		{"inch full", "1 inches", 25.4},
		{"ft", "1ft", 304.8},
		{"ft abbrv", "1'", 304.8},
		{"ft abbrv", "1feet", 304.8},
		{"ft full", "1Foot", 304.8},
		{"compound", "1.6' 2inch", 538.48},
		{"compound", "1foot 2.3inch", 363.22},
		{"compound", "1ft 2inch", 355.6},
		{"compound", "1ft 2inch", 355.6},
		{"compound", "1ft. 2inch", 355.6},
		{"compound", "1feet 2inches", 355.6},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			v, err := parseUnitString(tst.val)
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}

			if v != tst.exp {
				t.Errorf("result = %v; want %v", v, tst.exp)
			}
		})
	}

}

package grbl

import (
	"testing"
	"time"
)

func TestRate(t *testing.T) {
	d := Kilometer * 402336
	r := d.Rate(time.Hour)
	if r.MPH() != 250000 {
		t.Errorf("mph = %f; want 250000", r.MPH())
	}
}

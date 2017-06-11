package grbl

import (
	"math"
	"time"
)

type Accel struct {
	d  Distance
	t  time.Duration
	t2 time.Duration
}

// MMSec2 will return the acceleration in mm/sec^2
func (a Accel) MMSec2() float64 {
	return a.d.Millimeters() / a.t.Seconds() / a.t2.Seconds()
}

func LinearDistanceXY(x, y Distance) Distance {
	return Distance(math.Sqrt(float64(x*x + y*y)))
}
func LinearDistanceXYZ(x, y, z Distance) Distance {
	h := LinearDistanceXY(x, y)
	return Distance(math.Sqrt(float64(h*h + z*z)))
}
func (r Rate) Accel(t time.Duration) Accel {
	return Accel{
		d:  r.d,
		t:  r.t,
		t2: t,
	}
}

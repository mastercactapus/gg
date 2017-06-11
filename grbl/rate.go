package grbl

import "time"

// Rate unit. Stored as nm/hr
type Rate struct {
	d Distance
	t time.Duration
}

func (d Distance) Rate(t time.Duration) Rate {
	return Rate{d: d, t: t}
}
func NewRate(d Distance, t time.Duration) Rate {
	return Rate{t: t, d: d}
}
func (r Rate) MillimetersPerMinute() float64 {
	return r.d.Millimeters() / r.t.Minutes()
}
func (r Rate) InchesPerMinute() float64 {
	return r.d.Inches() / r.t.Minutes()
}
func (r Rate) KPH() float64 {
	return float64(r.d) / float64(Kilometer) / r.t.Hours()
}
func (r Rate) MPH() float64 {
	return float64(r.d) / float64(Mile) / r.t.Hours()
}
func (r Rate) TimeRequired(d Distance) time.Duration {
	return time.Duration(r.d*d) / r.t
}

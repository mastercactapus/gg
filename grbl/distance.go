package grbl

// Distance unit. Stored as nm.
type Distance int64

//	Distance units
const (
	Nanometer  Distance = 1
	Micrometer          = Nanometer * 1000
	Millimeter          = Micrometer * 1000
	Centimeter          = Millimeter * 10
	Meter               = Centimeter * 100
	Kilometer           = Meter * 1000
	Inch                = Micrometer * 25400
	Foot                = Inch * 12
	Yard                = Foot * 3
	Mile                = Foot * 5280
)

func (d Distance) Nanometers() int64 {
	return int64(d)
}
func (d Distance) Inches() float64 {
	return float64(d) / float64(Inch)
}
func (d Distance) Millimeters() float64 {
	return float64(d) / float64(Millimeter)
}

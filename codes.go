package gg

import "github.com/mastercactapus/gg/gcode"

var (
	FeedRateX = 600.0
	FeedRateY = 600.0
	FeedRateZ = 300.0

	setUnits = false
	absMode  = true
	firstAbs = true
	lastNum  = 0.0
	lastFeed = 0.0

	zPos = 0.0
)

var lines []gcode.Line

func CurrentZ() float64 {
	return zPos
}

func feedRate(l gcode.Line) float64 {
	rate := 0.0
	for _, w := range l {
		switch w.Type {
		case 'X':
			if rate == 0.0 || rate > FeedRateX {
				rate = FeedRateX
			}
		case 'Y':
			if rate == 0.0 || rate > FeedRateY {
				rate = FeedRateY
			}
		case 'Z':
			if rate == 0.0 || rate > FeedRateZ {
				rate = FeedRateZ
			}
		}
	}
	return rate
}

func withFeed(l gcode.Line) gcode.Line {
	if l.HasWord('F') {
		return l
	}

	return append(l, F(feedRate(l)))
}

func withoutType(l gcode.Line, t byte) gcode.Line {
	res := l[:0]
	for _, w := range l {
		if w.Type == t {
			continue
		}
		res = append(res, w)
	}
	return res
}

func print(l gcode.Line) {
	if l[0].Type != 'G' {
		lines = append(lines, l)
		return
	}

	if l.HasWord('Z') {
		if absMode {
			zPos = l.Value('Z')
		} else {
			zPos += l.Value('Z')
		}
	}

	switch l[0].Value {
	case 21, 20:
		setUnits = true
	case 90:
		if !firstAbs && absMode {
			return
		}
		absMode = true
		firstAbs = false
	case 91:
		if !firstAbs && !absMode {
			return
		}
		absMode = false
		firstAbs = false

	case 1, 2, 3:
		l = withFeed(l)

		f := l.Value('F')
		if f == lastFeed {
			l = withoutType(l, 'F')
		} else if f != 0 {
			lastFeed = f
		}
	}

	lastNum = l[0].Value

	lines = append(lines, l)
}

// G90 sets distance to absolute mode
func G90() { print(gcode.Line{{Type: 'G', Value: 90}}) }

// G91 sets distance to relative mode
func G91() { print(gcode.Line{{Type: 'G', Value: 91}}) }

// G93 sets the feed rate to inverse time mode instead of units
// per minute.
//
// The feed rate is calculated as 1/F.
//
// For example, in inverse mode a feed rate of 2.0 means a move
// should be completed in 1/2 minute (or 30 seconds).
//
// Feed rates are required for all G1, G2, and G3 commands while
// in this mode.
func G93() { print(gcode.Line{{Type: 'G', Value: 93}}) }

// G94 sets the feed rate to units per minute mode.
//
// The time for a move to complete depends on the total distance
// traveled with the feed rate in consideration.
func G94() { print(gcode.Line{{Type: 'G', Value: 94}}) }

// G0 is for rapid motion.
func G0(words ...gcode.Word) { print(append(gcode.Line{{Type: 'G', Value: 0}}, words...)) }

// G1 is for linear (straight line) motion at a set rate.
func G1(words ...gcode.Word) { print(append(gcode.Line{{Type: 'G', Value: 1}}, words...)) }

// G2 is used to make circular or helical movements *clockwise*.
func G2(words ...gcode.Word) { print(append(gcode.Line{{Type: 'G', Value: 2}}, words...)) }

// G3 is used to make circular or helical movements *counter-clockwise*.
func G3(words ...gcode.Word) { print(append(gcode.Line{{Type: 'G', Value: 3}}, words...)) }

// F controls feed rate
func F(val float64) gcode.Word { return gcode.Word{Type: 'F', Value: val} }

// I is the X Axis offset when doing arcs (G2,G3)
func I(val float64) gcode.Word { return gcode.Word{Type: 'I', Value: val} }

// J is the Y Axis offset when doing arcs (G2,G3)
func J(val float64) gcode.Word { return gcode.Word{Type: 'J', Value: val} }

// K is the Z Axis offset when doing arcs (G2,G3)
func K(val float64) gcode.Word { return gcode.Word{Type: 'K', Value: val} }

// L is used for G10 or canned cycle parameters
func L(val uint8) gcode.Word { return gcode.Word{Type: 'L', Value: float64(val)} }

// N is used to denote line number
func N(val int32) gcode.Word { return gcode.Word{Type: 'N', Value: float64(val)} }

// P is used for G10 and dwell (G4) parameters
func P(val float64) gcode.Word { return gcode.Word{Type: 'P', Value: val} }

// R is the arc radius (G2,G3)
func R(val float64) gcode.Word { return gcode.Word{Type: 'R', Value: val} }

// S controls the spindle speed
func S(val float64) gcode.Word { return gcode.Word{Type: 'S', Value: val} }

// T is used for tool selection
func T(val float64) gcode.Word { return gcode.Word{Type: 'T', Value: val} }

// X is used for the X coordinates
func X(val float64) gcode.Word { return gcode.Word{Type: 'X', Value: val} }

// Y is used for Y coordinates
func Y(val float64) gcode.Word { return gcode.Word{Type: 'Y', Value: val} }

// Z is used for Z coordinates
func Z(val float64) gcode.Word { return gcode.Word{Type: 'Z', Value: val} }

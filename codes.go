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

func G90()                   { print(gcode.Line{{'G', 90}}) }
func G91()                   { print(gcode.Line{{'G', 91}}) }
func G0(words ...gcode.Word) { print(append(gcode.Line{{'G', 0}}, words...)) }
func G1(words ...gcode.Word) { print(append(gcode.Line{{'G', 1}}, words...)) }
func G2(words ...gcode.Word) { print(append(gcode.Line{{'G', 2}}, words...)) }
func G3(words ...gcode.Word) { print(append(gcode.Line{{'G', 3}}, words...)) }

func X(val float64) gcode.Word { return gcode.Word{Type: 'X', Value: val} }
func Y(val float64) gcode.Word { return gcode.Word{Type: 'Y', Value: val} }
func Z(val float64) gcode.Word { return gcode.Word{Type: 'Z', Value: val} }

func F(val float64) gcode.Word { return gcode.Word{Type: 'F', Value: val} }

func I(val float64) gcode.Word { return gcode.Word{Type: 'I', Value: val} }
func J(val float64) gcode.Word { return gcode.Word{Type: 'J', Value: val} }

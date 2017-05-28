package gcode

import "strings"

// A Line is a set of Words to be sent to the CNC as a single unit.
type Line []Word

func (l Line) String() string {
	s := make([]string, len(l))
	for i, w := range l {
		s[i] = w.String()
	}
	return strings.Join(s, "")
}

func (l Line) Modal() string {
	if len(l) == 0 {
		return ""
	}
	if l[0].Type == 'G' || l[0].Type == 'M' {
		return l[0].String()
	}
	return ""
}

func (l Line) HasWord(t byte) bool {
	for _, w := range l {
		if w.Type == t {
			return true
		}
	}
	return false
}

func (l Line) Value(t byte) float64 {
	for _, w := range l {
		if w.Type == t {
			return w.Value
		}
	}
	return 0
}

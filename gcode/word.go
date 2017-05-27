package gcode

import (
	"strconv"
	"strings"
)

// A Word is a byte-number pair. Examples would be `G0` or `X-2`
type Word struct {
	Type  byte
	Value float64
}

func formatFloat(f float64) string {
	s := strings.TrimSuffix(strconv.FormatFloat(f, 'f', 3, 64), ".000")
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
	}
	return s
}

func (w Word) String() string {
	return string(w.Type) + formatFloat(w.Value)
}

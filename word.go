package gg

import (
	"strconv"
	"strings"
)

type WordType byte

const (
	WordTypeX WordType = 'X'
	WordTypeY WordType = 'Y'
	WordTypeZ WordType = 'Z'

	WordTypeF WordType = 'F'

	WordTypeI WordType = 'I'
	WordTypeJ WordType = 'J'
)

type Word struct {
	Type  WordType
	Value float64
}

func (w Word) String() string {

	return string(w.Type) + strings.TrimSuffix(strconv.FormatFloat(w.Value, 'f', 3, 64), ".000")
}

func X(val float64) Word { return Word{Type: WordTypeX, Value: val} }
func Y(val float64) Word { return Word{Type: WordTypeY, Value: val} }
func Z(val float64) Word { return Word{Type: WordTypeZ, Value: val} }

func F(val float64) Word { return Word{Type: WordTypeF, Value: val} }

func I(val float64) Word { return Word{Type: WordTypeI, Value: val} }
func J(val float64) Word { return Word{Type: WordTypeJ, Value: val} }

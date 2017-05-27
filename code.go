package gg

import (
	"fmt"
	"strings"
)

type CodeType byte

const (
	CodeTypeG CodeType = 'G'
	CodeTypeM CodeType = 'M'
)

type Code struct {
	Type   CodeType
	Number float64
	Words  []Word
}

func wordsString(ws []Word) string {
	words := make([]string, len(ws))
	for i, w := range ws {
		words[i] = w.String()
	}
	return strings.Join(words, "")
}

func (c Code) String() string {
	return fmt.Sprintf("%c%v%s", c.Type, c.Number, wordsString(c.Words))
}

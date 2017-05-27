package gg

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

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

func CurrentZ() float64 {
	return zPos
}

func feedRate(words []Word) float64 {
	rate := 0.0
	for _, w := range words {
		switch w.Type {
		case WordTypeX:
			if rate == 0.0 || rate > FeedRateX {
				rate = FeedRateX
			}
		case WordTypeY:
			if rate == 0.0 || rate > FeedRateY {
				rate = FeedRateY
			}
		case WordTypeZ:
			if rate == 0.0 || rate > FeedRateZ {
				rate = FeedRateZ
			}
		}
	}
	return rate
}

func hasWord(words []Word, t WordType) bool {
	for _, w := range words {
		if w.Type == t {
			return true
		}
	}
	return false
}
func withFeed(words []Word) []Word {
	if hasWord(words, WordTypeF) {
		return words
	}

	return append(words, F(feedRate(words)))
}
func val(words []Word, t WordType) float64 {
	for _, w := range words {
		if w.Type == t {
			return w.Value
		}
	}
	return 0
}
func withoutType(words []Word, t WordType) []Word {
	res := words[:0]
	for _, w := range words {
		if w.Type == t {
			continue
		}
		res = append(res, w)
	}
	return res
}
func print(t CodeType, n float64, w []Word) {
	printCode(Code{Type: t, Number: n, Words: w})
}
func printCode(c Code) {

	if !setUnits && (c.Type != CodeTypeG || (c.Number != 21 && c.Number != 20)) {
		logrus.Warnln("Units unspecified, defaulting to mm.")
		print(CodeTypeG, 21, nil)
	}

	if c.Type != CodeTypeG {
		fmt.Println(c.String())
		return
	}

	if hasWord(c.Words, WordTypeZ) {
		if absMode {
			zPos = val(c.Words, WordTypeZ)
		} else {
			zPos += val(c.Words, WordTypeZ)
		}
	}

	switch c.Number {
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
		f := val(c.Words, WordTypeF)
		if f == lastFeed {
			c.Words = withoutType(c.Words, WordTypeF)
		} else if f != 0 {
			lastFeed = f
		}
	}

	// if lastNum == c.Number && (c.Number == 0 || c.Number == 1) {
	// 	fmt.Println(wordsString(c.Words))
	// 	return
	// }

	lastNum = c.Number

	fmt.Println(c.String())
}

func G21()             { print(CodeTypeG, 21, nil) }
func G90()             { print(CodeTypeG, 90, nil) }
func G91()             { print(CodeTypeG, 91, nil) }
func G0(words ...Word) { print(CodeTypeG, 0, words) }
func G1(words ...Word) { print(CodeTypeG, 1, withFeed(words)) }
func G2(words ...Word) { print(CodeTypeG, 2, withFeed(words)) }
func G3(words ...Word) { print(CodeTypeG, 3, withFeed(words)) }

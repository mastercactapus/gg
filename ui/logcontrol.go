package ui

import "strings"

type Logger struct {
	X, Y      int
	Height    int
	Width     int
	MaxBuffer int

	data []byte
}

func (l *Logger) Write(p []byte) (int, error) {
	if l.MaxBuffer == 0 {
		l.data = append(l.data, p...)
		return len(p), nil
	}

	if len(p) > l.MaxBuffer {
		l.data = append(l.data[:0], p[len(p)-l.MaxBuffer:]...)
	} else if len(p)+len(l.data) > l.MaxBuffer {
		l.data = append(l.data[len(p)+len(l.data)-l.MaxBuffer:], p...)
	} else {
		l.data = append(l.data, p...)
	}

	return len(p), nil
}

func (l *Logger) Draw(r Renderer) {
	sw, sh := r.Size()
	x, y, w, h := StandardSize(l.X, l.Y, l.Width, l.Height, sw, sh)

	var lines [][]rune
	var i int
	var line string
	d := string(l.data)
	for len(lines) < h && len(d) > 0 {
		i = strings.LastIndexByte(d, '\n')
		if i != -1 {
			line = d[i+1:]
			d = d[:i]
		} else {
			line = d
			d = d[:0]
		}
		for len(line) > w {
			i = len(line) - w
			lines = append(lines, []rune(line[i:]))
			line = line[:i]
		}
		if len(line) > 0 {
			lines = append(lines, []rune(line))
		}
	}
	if l.MaxBuffer == 0 && len(d) > 0 {
		l.data = l.data[len(d):]
	}

	if len(lines) <= h {
		h = len(lines) - 1
	}
	yMx := h + y
	for i, l := range lines {
		putRunesA(r, x, yMx-i, l, 0, 0)
		if len(l) < w {
			putRunesA(r, x+len(l), yMx-i, []rune(strings.Repeat(" ", w-len(l))), 0, 0)
		}
	}
}

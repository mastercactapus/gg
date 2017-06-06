package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mastercactapus/gg/gcode"
	termbox "github.com/nsf/termbox-go"
)

type gcodeState int

const (
	gcodeStateReady gcodeState = iota
	gcodeStateSent
	gcodeStateRunning
	gcodeStateDone
)

func renderGcode(s CellSetter, x, y, w int, g gcode.Line, state gcodeState) {
	stat := ' '
	var fg, bg termbox.Attribute
	switch state {
	case gcodeStateSent:
		fg = termbox.ColorYellow
		stat = 'S'
	case gcodeStateRunning:
		bg = termbox.ColorBlue
		fg = termbox.ColorYellow
		stat = 'R'
	case gcodeStateDone:
		fg = termbox.ColorGreen
		stat = 'D'
	}

	line := fmt.Sprintf("[%c] %-5s %s", stat, g[0].String(), g[1:].String())
	line += strings.Repeat(" ", w-len(line))

	putRunesA(s, x, y, []rune(line), fg, bg)
}

type GCodeViewer struct {
	Lines        []gcode.Line
	Active, Sent int

	X, Y   int
	Width  int
	Height int
	Top    int
}

func (g *GCodeViewer) Draw(r Renderer) {
	sw, sh := r.Size()
	x, y, w, h := StandardSize(g.X, g.Y, g.Width, g.Height, sw, sh)
	if h < 1 {
		return
	}
	if h == 1 {
		putRunes(r, x, y, []rune("-- Screen too small to display GCode"))
		return
	}
	h--
	top := g.Active - 5
	if top < 0 {
		top = 0
	}
	l := g.Lines[top:]
	if h < len(l) {
		l = l[:h]
	}

	header := "-- Showing lines " + strconv.Itoa(top+1) + "-" + strconv.Itoa(top+len(l)) + " of " + strconv.Itoa(len(g.Lines))
	putRunes(r, x, y, []rune(header))
	y++

	for i, line := range l {
		var state gcodeState
		ln := i + top + 1
		if ln < g.Active {
			state = gcodeStateDone
		} else if ln == g.Active {
			state = gcodeStateRunning
		} else if ln <= g.Sent {
			state = gcodeStateSent
		} else {
			state = gcodeStateReady
		}
		renderGcode(r, x, y+i, w, line, state)
	}
}

package ui

import (
	"log"
	"time"

	"github.com/mastercactapus/gg/gcode"
	"github.com/mastercactapus/gg/grbl"
	"github.com/mastercactapus/gg/shuttlexpress"
)

func (j *JobUI) handleShuttleEvent(e shuttlexpress.Event) {
	switch e.Type {
	case shuttlexpress.EventTypeButton:
		switch e.Value {
		case 0:
			if j.jogStep > 0.001 {
				j.jogStep /= 10
			}
		case 4:
			if j.jogStep < 10 {
				j.jogStep *= 10
			}
		case 1:
			j.shuttleAxis = 'X'
		case 2:
			j.shuttleAxis = 'Y'
		case 3:
			j.shuttleAxis = 'Z'
		}
	case shuttlexpress.EventTypeConnection:
		switch e.Value {
		case shuttlexpress.ConnectionFailed:
			j.shuttleConnected = false
			j.shuttleBusted = true
		case shuttlexpress.ConnectionLost:
			j.shuttleConnected = false
			j.shuttleBusted = false
		case shuttlexpress.ConnectionSuccess:
			j.shuttleConnected = true
			j.shuttleBusted = false
		}
	case shuttlexpress.EventTypeWheel:
		j.shuttleMove(e.Value)
	case shuttlexpress.EventTypeRing:
		j.shuttleRing = e.Value
		if e.Value == 0 {
			j.c.JogCancel()
		}
	}
}
func (j *JobUI) shuttleMove(val int) {
	log.Println("Move", val, j.s.State, j.shuttleAxis)
	if !j.shuttleConnected {
		log.Println("not connected")
		return
	}
	if j.s.State != grbl.StateIdle && j.s.State != grbl.StateJog {
		log.Println("bad state")
		return
	}
	if j.shuttleAxis == 0 {
		log.Println("axis busted")
		return
	}
	s := time.Now()
	j.c.Jog(gcode.Line{
		{Type: 'G', Value: 91},
		{Type: j.shuttleAxis, Value: j.jogStep * float64(val)},
		{Type: 'F', Value: 10000},
	})
	log.Println("move conf", time.Since(s))
}

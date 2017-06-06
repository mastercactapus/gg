package ui

import (
	"regexp"
	"time"

	"github.com/mastercactapus/gg/gcode"
	"github.com/mastercactapus/gg/grbl"
)

type action int

const (
	actionCheckCode action = iota
	actionRunJob
	actionStopJob
)

type checkStatus struct {
	line     int
	err      error
	complete bool
}

type JobUI struct {
	c  *grbl.Grbl
	ui *UI
	g  []gcode.Line

	checked bool
	v       GCodeViewer

	recv        chan grbl.Response
	s           grbl.Status
	recvStatus  chan grbl.Status
	checkStatus chan checkStatus

	actionCh chan action
	renderCh chan struct{}
	closeCh  chan struct{}
}

func (j *JobUI) renderSync() {
	<-j.renderCh
}

func (j *JobUI) loop() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	defer close(j.renderCh)

	for {
		select {
		case j.renderCh <- struct{}{}:
			j.renderCh <- struct{}{}
			continue
		case <-t.C:
			j.c.Status()
		case check := <-j.checkStatus:
			j.v.Active = check.line
			j.v.Sent = check.line
		case j.s = <-j.recvStatus:
		case a := <-j.actionCh:
			j.handleAction(a)
		}
		j.ui.Render()
	}
}

var lineRx = regexp.MustCompile("^N([0-9]+)")

func (j *JobUI) handleAction(a action) {
	switch a {
	case actionCheckCode:
		j.performCheck()
	case actionStopJob:
		j.performStop()
	}
}
func (j *JobUI) performStop() {
	switch j.s.State {
	case grbl.StateCheck:
		j.v.Active = 0
		j.v.Sent = 0
		j.c.SoftReset()
	}
}

func NewJobUI(c *grbl.Grbl, g []gcode.Line) (*JobUI, error) {
	j := &JobUI{
		c:           c,
		g:           g,
		renderCh:    make(chan struct{}),
		actionCh:    make(chan action, 1),
		recvStatus:  c.Status(),
		checkStatus: make(chan checkStatus),
	}
	for i, l := range j.g {
		j.g[i] = append(gcode.Line{gcode.Word{Type: 'N', Value: float64(i + 1)}}, l...)
	}
	ui, err := NewUI(j.render)
	if err != nil {
		return nil, err
	}
	j.ui = ui

	j.v = GCodeViewer{
		Lines: j.g,
		Y:     6,
		X:     1,
	}

	go j.loop()
	return j, nil
}

func (j *JobUI) isRunning() bool {
	return j.s.State == grbl.StateJog || j.s.State == grbl.StateCheck || j.s.State == grbl.StateRun
}
func (j *JobUI) statusText() string {
	switch {
	case !j.checked && !j.isRunning():
		return "Unchecked"
	case j.checked && !j.isRunning():
		return "Idle"
	case !j.checked && j.isRunning():
		return "Checking"
	case j.checked && j.isRunning():
		return "Running"
	}

	return "UI Broken"
}
func (j *JobUI) status() Control {
	switch {
	default:
		return &Text{X: 1, Y: 3, Lines: []string{"No job running.", ""}}
	case !j.checked && !j.isRunning():
		return &Text{X: 1, Y: 3, Lines: []string{"Perform 'Check' to get started.", ""}}
	case !j.checked && j.isRunning():
		return &Progress{
			X:     1,
			Y:     3,
			Width: -1,
			Title: "Checking...",
			Max:   len(j.v.Lines),
			Value: j.v.Active - 1,
		}
	case j.checked && j.isRunning():
		return &Text{
			X: 1, Y: 3,
			Lines: []string{
				"Est. Duration: " + j.duration().String(),
				"    Remaining: " + j.remaining().String(),
			},
		}
	}

}

func (j *JobUI) performCheck() {
	resp := j.c.CheckGCode(j.g)
	go func() {
		ln := 1
		for stat := range resp {
			j.checkStatus <- checkStatus{
				line: ln,
				err:  stat.Err,
			}
			ln++
		}
		j.checkStatus <- checkStatus{complete: true}
	}()
}

func (j *JobUI) duration() time.Duration {
	return time.Minute + time.Second*30
}
func (j *JobUI) remaining() time.Duration {
	return time.Second * 19
}

func (j *JobUI) machineStatusText() string {
	if j.s.State == grbl.StateUnknown {
		return "Machine Status -- Connecting"
	}
	return "Machine Status -- " + string(j.s.State)
}

func (j *JobUI) Start() error {
	j.ui.MainLoop()
	j.ui.Close()
	return nil
}

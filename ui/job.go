package ui

import (
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

type JobUI struct {
	c  *grbl.Conn
	ui *UI
	g  []gcode.Line

	running bool
	checked bool
	v       GCodeViewer

	recv       chan grbl.Response
	s          *grbl.Status
	recvStatus chan *grbl.Status

	actionCh chan action
	renderCh chan struct{}
}

func (j *JobUI) renderSync() {
	<-j.renderCh
}

func (j *JobUI) loop() {
	t := time.NewTicker(time.Millisecond * 100)
	defer t.Stop()
	defer close(j.renderCh)

	for {
		select {
		case j.renderCh <- struct{}{}:
			j.renderCh <- struct{}{}
			continue
		case <-t.C:
			j.c.Status()
		case j.s = <-j.recvStatus:
		case r := <-j.recv:
			j.handleResponse(r)
		case a := <-j.actionCh:
			j.handleAction(a)
		}
		j.ui.Render()
	}
}

func (j *JobUI) handleResponse(r grbl.Response) {

}

func (j *JobUI) handleAction(a action) {
	switch a {
	case actionCheckCode:
		j.checked = false
		j.running = true
	}
}

func NewJobUI(c *grbl.Conn, g []gcode.Line) (*JobUI, error) {

	j := &JobUI{
		c:          c,
		g:          g,
		renderCh:   make(chan struct{}),
		actionCh:   make(chan action, 1),
		recv:       c.Recv(),
		recvStatus: c.RecvStatus(),
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
func (j *JobUI) statusText() string {
	switch {
	case !j.checked && !j.running:
		return "Unchecked"
	case j.checked && !j.running:
		return "Idle"
	case !j.checked && j.running:
		return "Checking"
	case j.checked && j.running:
		return "Running"
	}

	return "UI Broken"
}
func (j *JobUI) status() Control {
	switch {
	default:
		return &Text{X: 1, Y: 3, Lines: []string{"No job running.", ""}}
	case !j.checked && !j.running:
		return &Text{X: 1, Y: 3, Lines: []string{"Perform 'Check' to get started."}}
	case !j.checked && j.running:
		return &Progress{
			X:     1,
			Y:     3,
			Width: -1,
			Title: "Checking...",
			Max:   len(j.v.Lines),
			Value: j.v.Active - 1,
		}
	case j.checked && j.running:
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
	j.running = true
	j.checked = false

}

func (j *JobUI) duration() time.Duration {
	return time.Minute + time.Second*30
}
func (j *JobUI) remaining() time.Duration {
	return time.Second * 19
}

func (j *JobUI) render() []Control {
	j.renderSync()
	defer j.renderSync()

	return []Control{
		&Group{
			Title: "GCODE -- " + j.statusText(),
			Width: 40,
			X:     10,
			Controls: []Control{
				&Button{
					X:           5,
					Y:           1,
					Text:        "Check",
					Disabled:    j.running,
					OnClickFunc: func(x, y int) { j.actionCh <- actionCheckCode },
				},
				&Button{
					X:           16,
					Y:           1,
					Text:        "Run",
					Disabled:    j.running || !j.checked,
					OnClickFunc: func(x, y int) { j.actionCh <- actionRunJob },
				},
				&Button{
					X:           25,
					Y:           1,
					Text:        "Stop",
					Disabled:    !j.running,
					OnClickFunc: func(x, y int) { j.actionCh <- actionStopJob },
				},
				j.status(),
				&j.v,
			},
		},
		&Group{
			Title:  j.machineStatusText(),
			Width:  40,
			Height: 10,
			X:      60,
			Controls: []Control{
				&Text{Lines: []string{"hi"}},
				&Status{Status: j.s},
			},
		},
	}
}
func (j *JobUI) machineStatusText() string {
	if j.s == nil {
		return "Machine Status -- Connecting"
	}
	return "Machine Status -- " + string(j.s.State)
}

func (j *JobUI) Start() error {
	j.ui.MainLoop()
	j.ui.Close()
	return nil
}

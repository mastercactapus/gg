package ui

import (
	"log"
	"regexp"
	"time"

	"github.com/mastercactapus/gg/gcode"
	"github.com/mastercactapus/gg/grbl"
	"github.com/mastercactapus/gg/shuttlexpress"
)

type action int

var jogSteps = []float64{0.001, 0.01, 0.1, 1, 10}

const jogStepIncr = -1
const jogStepDecr = -2

const (
	actionCheckCode action = iota
	actionRunJob
	actionStopJob
)

type gcodeStatus struct {
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
	jogStep float64

	recv         chan grbl.Response
	s            grbl.Status
	settings     grbl.Settings
	recvStatus   chan grbl.Status
	recvSettings chan grbl.Settings
	checkStatus  chan gcodeStatus
	jobStatus    chan gcodeStatus

	setJogStep chan float64
	jogStepCh  chan byte
	zeroAxis   chan byte
	goZeroAxis chan byte

	actionCh chan action
	renderCh chan struct{}
	closeCh  chan struct{}

	shuttleEvents    chan shuttlexpress.Event
	shuttleConnected bool
	shuttleBusted    bool
	shuttleRing      int
	shuttleAxis      byte

	l *Logger
}

func NewJobUI(c *grbl.Grbl, g []gcode.Line) (*JobUI, error) {
	l := &Logger{}
	s := shuttlexpress.NewDevice(log.New(l, "ShuttleXpress: ", 0))
	log.SetOutput(l)
	log.SetFlags(0)
	c.SetLogger(log.New(l, "Grbl", 0))
	j := &JobUI{
		c:            c,
		g:            g,
		renderCh:     make(chan struct{}),
		actionCh:     make(chan action, 1),
		recvStatus:   c.Status(),
		recvSettings: c.Settings(),
		checkStatus:  make(chan gcodeStatus),
		jobStatus:    make(chan gcodeStatus),
		setJogStep:   make(chan float64),
		jogStep:      0.01,
		jogStepCh:    make(chan byte),
		goZeroAxis:   make(chan byte),
		zeroAxis:     make(chan byte),

		shuttleEvents: s.Events(),

		l: l,
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

func (j *JobUI) renderSync() {
	<-j.renderCh
}

func (j *JobUI) loop() {
	t := time.NewTicker(time.Millisecond * 100)
	defer t.Stop()
	defer close(j.renderCh)

	for {
		select {
		case s := <-j.recvSettings:
			j.settings = s
		case e := <-j.shuttleEvents:
			j.handleShuttleEvent(e)
		case w := <-j.zeroAxis:
			if w == '_' {
				j.c.ExecLine(gcode.Line{
					gcode.Word{Type: 'G', Value: 92},
					gcode.Word{Type: 'X'},
					gcode.Word{Type: 'Y'},
					gcode.Word{Type: 'Z'},
				})
				continue
			}
			j.c.ExecLine(gcode.Line{
				gcode.Word{Type: 'G', Value: 92},
				gcode.Word{Type: w},
			})
		case w := <-j.goZeroAxis:
			j.s.State = grbl.StateJog
			if w == '_' {
				j.c.Jog(gcode.Line{
					gcode.Word{Type: 'G', Value: 90},
					gcode.Word{Type: 'X'},
					gcode.Word{Type: 'Y'},
					gcode.Word{Type: 'F', Value: 10000},
				})

				j.c.Jog(gcode.Line{
					gcode.Word{Type: 'G', Value: 90},
					gcode.Word{Type: 'Z'},
					gcode.Word{Type: 'F', Value: 10000},
				})
				continue
			}
			if w == 'H' {
				j.s.State = grbl.StateHome
				go func() {
					j.c.Home()
					j.c.Settings()
				}()
				continue
			}
			j.c.Jog(gcode.Line{
				gcode.Word{Type: 'G', Value: 90},
				gcode.Word{Type: w},
				gcode.Word{Type: 'F', Value: 10000},
			})
		case w := <-j.jogStepCh:
			j.s.State = grbl.StateJog
			var m gcode.Word
			if w > 'a' {
				m.Type = w - 32
				m.Value = -j.jogStep
			} else {
				m.Type = w
				m.Value = j.jogStep
			}

			j.c.Jog(gcode.Line{
				gcode.Word{Type: 'G', Value: 91},
				gcode.Word{Type: 'G', Value: 21},
				m,
				gcode.Word{Type: 'F', Value: 10000},
			})
		case v := <-j.setJogStep:
			j.jogStep = v
		case j.renderCh <- struct{}{}:
			j.renderCh <- struct{}{}
			continue
		case <-t.C:
			j.c.Status()
		case stat := <-j.jobStatus:
			if stat.complete {
				j.v.Active = -1
				j.v.Sent = -1
				continue
			}
			j.v.Active = stat.line - 16
			j.v.Sent = stat.line
		case check := <-j.checkStatus:
			if check.complete {
				j.v.Active = -1
				j.v.Sent = -1
				j.s.State = ""
				j.checked = true
				continue
			}
			j.v.Active = check.line
			j.v.Sent = check.line
		case j.s = <-j.recvStatus:
		case a := <-j.actionCh:
			j.handleAction(a)
		}
		j.ui.Render()
	}
}

func (j *JobUI) JogStep(a byte) {
	select {
	case j.jogStepCh <- a:
	default:
	}
}

var lineRx = regexp.MustCompile("^N([0-9]+)")

func (j *JobUI) handleAction(a action) {
	switch a {
	case actionCheckCode:
		j.performCheck()
	case actionStopJob:
		j.performStop()
	case actionRunJob:
		j.performRun()
	}
}
func (j *JobUI) performRun() {
	resp := j.c.RunGCode(j.g)
	go func() {
		ln := 1
		for stat := range resp {
			j.jobStatus <- gcodeStatus{
				line: ln,
				err:  stat.Err,
			}
			ln++
		}
		j.jobStatus <- gcodeStatus{complete: true}
	}()
}
func (j *JobUI) performStop() {
	switch j.s.State {
	case grbl.StateCheck:
		j.v.Active = 0
		j.v.Sent = 0
		j.c.SoftReset()
	case grbl.StateJog:
		j.c.JogCancel()
		j.s.State = grbl.StateIdle
	case grbl.StateIdle, grbl.StateRun:
		j.c.FeedHold()
	}
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
	case !j.checked && j.s.State == grbl.StateIdle:
		return &Text{X: 1, Y: 3, Lines: []string{"Perform 'Check' to get started.", ""}}
	case j.s.State == grbl.StateCheck:
		return &Progress{
			X:     1,
			Y:     3,
			Width: -1,
			Title: "Checking...",
			Max:   len(j.v.Lines),
			Value: j.v.Active - 1,
		}
	case j.checked && j.s.State == grbl.StateRun:
		return &Text{
			X: 1, Y: 3,
			Lines: []string{
				"Est. Duration: " + j.duration().String(),
				"    Remaining: " + j.remaining().String(),
			},
		}
	case j.s.State == grbl.StateAlarm:
		return &Group{
			X: 1, Y: 3, Height: 3,
			Width: -1,
			Title: "Alarm Mode",
			Clear: true,
			Controls: []Control{
				&Button{X: 1, Text: "Home", Enabled: true,
					OnClickFunc: func(int, int) { j.goZeroAxis <- 'H' },
				},
				&Button{X: 12, Text: "Unlock", Enabled: true,
					OnClickFunc: func(int, int) { j.c.Unlock() },
				},
			},
		}
	case j.s.State == grbl.StateHoldComplete:
		return &Group{
			X: 1, Y: 3, Height: 3,
			Width: -1,
			Title: "Feed Hold Active",
			Clear: true,
			Controls: []Control{
				&Button{X: 1, Text: "Resume", Enabled: true,
					OnClickFunc: func(int, int) { j.c.StartResume() },
				},
				&Button{X: 12, Text: "Reset", Enabled: true,
					OnClickFunc: func(int, int) { j.c.SoftReset() },
				},
			},
		}
	}
}

func (j *JobUI) performCheck() {
	j.checked = false
	resp := j.c.CheckGCode(j.g)
	go func() {
		ln := 1
		for stat := range resp {
			j.checkStatus <- gcodeStatus{
				line: ln,
				err:  stat.Err,
			}
			ln++
		}
		j.checkStatus <- gcodeStatus{complete: true}
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

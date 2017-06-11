package grbl

import (
	"errors"
	"io"
	"log"

	"github.com/mastercactapus/gg/gcode"
)

type Grbl struct {
	c *Client

	s        Status
	settings Settings

	statusCh   chan Status
	settingsCh chan Settings
}

func NewGrbl(rwc io.ReadWriteCloser) *Grbl {
	return NewGrblClient(NewClient(rwc, ModeCharacterCount))
}
func NewGrblClient(c *Client) *Grbl {
	g := &Grbl{
		c: c,

		statusCh:   make(chan Status),
		settingsCh: make(chan Settings, 1),
	}
	go g.loop()
	return g
}

func (g *Grbl) SerialMode() ClientMode {
	return g.c.Mode()
}
func (g *Grbl) SetSerialMode(m ClientMode) {
	g.c.SetMode(m)
}

func (g *Grbl) loop() {
	pCh := g.c.PushMessages()
	for {
		select {
		case data := <-pCh:
			if data[0] == '<' {
				s, err := parseMachineStatus(string(data))
				if err != nil {
					panic(err)
				}
				g.mergeStatus(s)
				g.statusCh <- g.s
				continue
			}

			if data[0] == '$' {
				g.settings.parseSetting(data)
				continue
			}

			switch string(data) {
			case "[MSG:Enabled]":
				g.s.State = StateCheck
				g.statusCh <- g.s
				continue
			}
		}
	}
}

func (g *Grbl) mergeStatus(s *Status) {
	g.s.State = s.State
	var mpos, wpos bool
	for _, f := range s.Fields {
		switch f {
		case "MPos":
			mpos = true
			g.s.MPos = s.MPos
			g.makeWPos()
		case "WPos":
			wpos = true
			g.s.WPos = s.WPos
			g.makeMPos()
		case "WCO":
			g.s.WCO = s.WCO
			if mpos && !wpos {
				g.makeWPos()
			} else if wpos && !mpos {
				g.makeMPos()
			}
		case "Ov":
			g.s.FieldOverrides = s.FieldOverrides
		}
	}
}

func (g *Grbl) makeWPos() {
	if g.s.WCO == nil {
		return
	}
	g.s.WPos = []float64{
		g.s.MPos[0] - g.s.WCO[0],
		g.s.MPos[1] - g.s.WCO[1],
		g.s.MPos[2] - g.s.WCO[2],
	}
}
func (g *Grbl) makeMPos() {
	if g.s.WCO == nil {
		return
	}
	g.s.MPos = []float64{
		g.s.WPos[0] + g.s.WCO[0],
		g.s.WPos[1] + g.s.WCO[1],
		g.s.WPos[2] + g.s.WCO[2],
	}
}
func (g *Grbl) JogCancel() {
	<-g.c.Execute([]byte{byte(rtJogCancel)})
	g.Status()
}
func (g *Grbl) FeedHold() {
	<-g.c.Execute([]byte{byte(rtFeedHold)})
	g.Status()
}
func (g *Grbl) Unlock() {
	<-g.c.Execute([]byte("$X\n"))
}
func (g *Grbl) Home() {
	ch := g.c.Execute([]byte("$H\n"))
	g.Status()
	<-ch
	g.Status()
}
func (g *Grbl) ExecLine(l gcode.Line) {
	<-g.c.Execute([]byte(l.String() + "\n"))
}
func (g *Grbl) Jog(l gcode.Line) {
	g.c.Execute([]byte("$J=" + l.String() + "\n"))
}
func (g *Grbl) Status() chan Status {
	g.c.Execute([]byte{byte(rtStatus)})
	return g.statusCh
}
func (g *Grbl) SoftReset() {
	<-g.c.Execute([]byte{byte(rtSoftReset)})
	g.Status()
}
func (g *Grbl) StartResume() {
	<-g.c.Execute([]byte{byte(rtStartResume)})
	g.Status()
}

type CheckStatus struct {
	Line int
	Err  error
}

func (g *Grbl) Settings() chan Settings {
	resp := g.c.Execute([]byte("$$\n"))
	go func() {
		d := <-resp
		if d.Err != nil {
			log.Println("failed to get settings:", d.Err)
			return
		}
		g.settingsCh <- g.settings
	}()
	return g.settingsCh
}

func (g *Grbl) RunGCode(lines []gcode.Line) chan CheckStatus {
	var cmds [][]byte
	for _, l := range lines {
		cmds = append(cmds, []byte(l.String()+"\n"))
	}
	ch := make(chan CheckStatus, len(lines))
	go func() {
		resp := g.c.ExecuteMany(cmds)
		var r *Response
		for i := range cmds {
			r = <-resp
			if r.Err != nil {
				ch <- CheckStatus{Line: i, Err: r.Err}
			} else if r.Data[0] == 'e' {
				ch <- CheckStatus{Line: i, Err: errors.New(string(r.Data))}
			} else {
				ch <- CheckStatus{Line: i}
			}
		}
		close(ch)
	}()
	return ch
}
func (g *Grbl) CheckGCode(lines []gcode.Line) chan CheckStatus {
	var cmds [][]byte
	cmds = append(cmds, []byte("$C\n"))
	cmds = append(cmds, []byte("G92X0Y0Z0\n"))
	for _, l := range lines {
		cmds = append(cmds, []byte(l.String()+"\n"))
	}
	cmds = append(cmds, []byte("$C\n"))

	ch := make(chan CheckStatus, len(lines))
	go func() {
		max := len(cmds)
		resp := g.c.ExecuteMany(cmds)
		var r *Response
		for i := -1; i < max; i++ {
			r = <-resp
			if i <= 0 || i == max-1 {
				continue
			}
			if r.Err != nil {
				ch <- CheckStatus{Line: i, Err: r.Err}
			} else if r.Data[0] == 'e' {
				ch <- CheckStatus{Line: i, Err: errors.New(string(r.Data))}
			} else {
				ch <- CheckStatus{Line: i}
			}
		}
		close(ch)
	}()
	return ch
}

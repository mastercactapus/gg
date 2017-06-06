package grbl

import (
	"errors"
	"io"

	"github.com/mastercactapus/gg/gcode"
)

type Grbl struct {
	c *Client

	s Status

	statusCh chan Status
}

func NewGrbl(rwc io.ReadWriteCloser) *Grbl {
	return NewGrblClient(NewClient(rwc, ModeCharacterCount))
}
func NewGrblClient(c *Client) *Grbl {
	g := &Grbl{
		c: c,

		statusCh: make(chan Status),
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
				g.s = *s
				g.statusCh <- g.s
			}

			switch string(data) {
			case "[MSG:Enabled]":
				g.s.State = StateCheck
				g.statusCh <- g.s
			}
		}
	}
}

func (g *Grbl) Status() chan Status {
	go g.c.Execute([]byte{byte(rtStatus)})
	return g.statusCh
}
func (g *Grbl) SoftReset() {
	g.c.Execute([]byte{byte(rtSoftReset)})
}

type CheckStatus struct {
	Line int
	Err  error
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

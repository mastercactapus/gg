package grbl

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mastercactapus/gg/gcode"
	"github.com/pkg/errors"
)

const grblBufSize = 127

type Conn struct {
	rwc        io.ReadWriteCloser
	grblBuffer []string

	queue []string

	respPos int

	sendQueueCh     chan string
	recvQueueCh     chan string
	errs            chan error
	ioErr           chan error
	realtimeQueueCh chan byte
	recv            chan string
	machineStatusCh chan *Status
	closeCh         chan struct{}
}

// A FatalError is one that the connection cannot recover from (e.g. i/o error)
type FatalError struct {
	Cause error
}

func (f FatalError) Error() string {
	return f.Cause.Error()
}

func NewConn(rwc io.ReadWriteCloser) *Conn {
	c := &Conn{
		rwc: rwc,

		ioErr:           make(chan error, 5),
		recvQueueCh:     make(chan string, 100),
		sendQueueCh:     make(chan string, 10000),
		errs:            make(chan error, 100),
		recv:            make(chan string, 10000),
		realtimeQueueCh: make(chan byte),
		machineStatusCh: make(chan *Status),

		closeCh: make(chan struct{}),
	}
	go c.readLoop()
	go c.loop()
	return c
}

func (c *Conn) readLoop() {
	r := bufio.NewReader(c.rwc)
	var err error
	var s string
	defer close(c.recvQueueCh)
	for {
		s, err = r.ReadString('\n')
		if err != nil {
			c.ioErr <- err
			return
		}
		c.recvQueueCh <- strings.TrimSpace(s)
	}
}

func (c *Conn) fillGrbl() {
	s := 0
	for _, n := range c.grblBuffer {
		s += len(n)
	}

	gap := grblBufSize - s
	var err error
	for len(c.queue) > 0 && gap > len(c.queue[0]) {
		c.recv <- "W:" + strconv.Quote(c.queue[0])
		_, err = c.rwc.Write([]byte(c.queue[0]))
		if err != nil {
			c.ioErr <- err
			return
		}

		gap -= len(c.queue[0])
		c.grblBuffer = append(c.grblBuffer, c.queue[0])

		copy(c.queue, c.queue[1:])
		c.queue = c.queue[:len(c.queue)-1]
	}

}

func parseFloats(s string) (n []float64, err error) {
	parts := strings.Split(s, ",")
	n = make([]float64, len(parts))
	for i, p := range parts {
		n[i], err = strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}
func parseInts(s string) (n []int, err error) {
	parts := strings.Split(s, ",")
	n = make([]int, len(parts))
	for i, p := range parts {
		n[i], err = strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (c *Conn) parseMachineStatus(data string) {
	data = strings.TrimPrefix(data, "<")
	data = strings.TrimSuffix(data, ">")
	parts := strings.Split(data, "|")

	var s Status
	var idx int
	var val string
	var err error
	for i, p := range parts {
		if i == 0 {
			s.State = State(p)
			continue
		}
		idx = strings.IndexByte(p, ':')
		if idx == -1 {
			continue
		}
		s.Fields = append(s.Fields, p[:idx])
		val = p[idx+1:]
		switch p[:idx] {
		case "MPos":
			s.MPos, err = parseFloats(val)
		case "WPos":
			s.WPos, err = parseFloats(val)
		case "Bf":
			var ints []int
			ints, err = parseInts(val)
			if len(ints) == 2 {
				s.BlockBufferAvailable = ints[0]
				s.SerialBufferAvailable = ints[1]
			}
		case "Ln":
			s.Line, err = strconv.Atoi(val)
		case "FS":
			var floats []float64
			floats, err = parseFloats(val)
			if len(floats) == 2 {
				s.FeedSpeed = floats[0]
				s.SpindleSpeed = floats[1]
			}
		case "F":
			s.FeedSpeed, err = strconv.ParseFloat(val, 64)
		case "Pn":
			for _, c := range val {
				switch c {
				case 'P':
					s.Pins.Probe = true
				case 'X':
					s.Pins.LimitX = true
				case 'Y':
					s.Pins.LimitY = true
				case 'Z':
					s.Pins.LimitZ = true
				case 'D':
					s.Pins.Door = true
				case 'H':
					s.Pins.FeedHold = true
				case 'S':
					s.Pins.CycleStart = true
				}
			}
		case "WCO":
			s.WCO, err = parseFloats(val)
		case "Ov":
			var ints []int
			ints, err = parseInts(val)
			if len(ints) == 3 {
				s.FieldOVerrides.F = ints[0]
				s.FieldOVerrides.R = ints[1]
				s.FieldOVerrides.SpindleSpeed = ints[2]
			}
		case "A":
			for _, c := range val {
				switch c {
				case 'S':
					s.Aux.SpindleDirection = SpindleDirectionCW
					s.Aux.SpindleOn = true
				case 'C':
					s.Aux.SpindleOn = true
					s.Aux.SpindleDirection = SpindleDirectionCCW
				case 'F':
					s.Aux.CoolantFlood = true
				case 'M':
					s.Aux.CoolantMist = true
				}
			}
		}
		if err != nil {
			c.errs <- errors.Wrap(err, "parse machine status")
			return
		}
	}

	select {
	case c.machineStatusCh <- &s:
	default:
	}
}

func (c *Conn) loop() {
	var err error
	var data string
	var rt byte
	defer close(c.errs)
	for {
		select {
		case <-c.closeCh:
			close(c.recv)
			return
		case rt = <-c.realtimeQueueCh:
			_, err = c.rwc.Write([]byte{rt})
			if err != nil {
				c.errs <- &FatalError{Cause: err}
				return
			}
			// TODO: deal with effects (like jog cancel - flushing the jog buffer)
		case err = <-c.ioErr:
			c.errs <- &FatalError{Cause: err}
			return
		case data = <-c.recvQueueCh:
			if len(data) == 0 {
				continue
			}

			if data[0] == '<' {
				c.parseMachineStatus(data)
				continue
			}

			if data == "ok" && len(c.grblBuffer) > 0 {
				copy(c.grblBuffer, c.grblBuffer[1:])
				c.grblBuffer = c.grblBuffer[:len(c.grblBuffer)-1]
			}

			c.recv <- "R:" + strconv.Quote(data)
			c.fillGrbl()
		case data = <-c.sendQueueCh:
			c.queue = append(c.queue, data)
			c.fillGrbl()
		}
	}
}

func (c *Conn) Status() *Status {
	c.realtimeQueueCh <- '?'
	return <-c.machineStatusCh
}

func (c *Conn) Recv() chan string {
	go func() {
		c.realtimeQueueCh <- '\n'
		time.Sleep(time.Second * 3)
		c.realtimeQueueCh <- '?'
		c.realtimeQueueCh <- '?'
		time.Sleep(time.Second)
		c.sendQueueCh <- "$\n"
	}()
	return c.recv
}

func (c *Conn) WriteGCode(l gcode.Line) {
	c.sendQueueCh <- l.String()
}

func (c *Conn) Close() error {
	c.closeCh <- struct{}{}
	return c.rwc.Close()
}

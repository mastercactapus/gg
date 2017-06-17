package shuttlexpress

import (
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/mastercactapus/gg/hidraw"
)

type Device struct {
	eventCh chan Event
	l       *log.Logger
}

type Event struct {
	Type  EventType
	Value int
}
type EventType int

const (
	EventTypeConnection EventType = iota
	EventTypeButton
	EventTypeWheel
	EventTypeRing
)

const (
	ConnectionSuccess = iota
	ConnectionLost
	ConnectionFailed
)

func NewDevice(l *log.Logger) *Device {
	if l == nil {
		l = log.New(ioutil.Discard, "", 0)
	}
	d := &Device{
		eventCh: make(chan Event),
		l:       l,
	}
	go d.loop()
	return d
}

func (d *Device) loop() {
	delay := time.Duration(0)
	for {
		time.Sleep(delay)
		delay = time.Second
		dev, err := hidraw.OpenInputDevice(vendor, product)
		if err == hidraw.ErrNoDevice {
			d.eventCh <- Event{Type: EventTypeConnection, Value: ConnectionLost}
			continue
		}
		if err != nil {
			d.l.Println(err)
			d.eventCh <- Event{Type: EventTypeConnection, Value: ConnectionFailed}
			continue
		}

		d.handleDevice(dev)
	}
}

func (d *Device) Events() chan Event {
	return d.eventCh
}

func (d *Device) handleDevice(dev io.ReadCloser) {
	defer dev.Close()
	se := newShuttleXpress(dev)

	for e := range se.e {
		d.eventCh <- e
	}
}

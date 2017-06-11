package shuttlexpress

import (
	"io"
	"log"
	"time"

	"github.com/mastercactapus/gg/hidraw"
)

type Device struct {
	eventCh chan Event
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

func NewDevice() *Device {
	d := &Device{
		eventCh: make(chan Event),
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
			log.Println("ERR:", err)
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

package shuttlexpress

import (
	"encoding/binary"
	"io"
	"time"
)

var btnMap = [...]uint{12, 13, 14, 15, 0}

const (
	vendor  = 0x0b33
	product = 0x0020

	// RingMax is the max value for an EventTypeRing
	RingMax = 7

	// RingMin is the min value for an EventTypeRing
	RingMin = -7
)

type shuttleXpress struct {
	r io.Reader

	e          chan Event
	old        shuttleXpressState
	firstEvent bool //need to track first event, because we have no idea where the 'jog' position will start
	bState     [5]struct {
		pressed bool
		t       time.Time
	}
}

type shuttleXpressState struct {
	Ring    int8
	Jog     byte
	_       byte
	Buttons uint16
}

func newShuttleXpress(r io.Reader) *shuttleXpress {
	s := &shuttleXpress{
		r: r,
		e: make(chan Event),
	}
	go s.loop()
	return s
}

func (s *shuttleXpress) updateOneBtn(i int, state bool) (changed bool, dur time.Duration) {
	if s.bState[i].pressed && !state {
		changed = true
		dur = time.Since(s.bState[i].t)
	} else if !s.bState[i].pressed && state {
		changed = true
		s.bState[i].t = time.Now()
	}
	s.bState[i].pressed = state
	return
}
func (s *shuttleXpress) updateButtons(st shuttleXpressState) int {
	var btn = -1
	var pressed bool
	var changed bool
	for i := range btnMap {
		pressed = st.Buttons&(1<<btnMap[i]) != 0
		changed, _ = s.updateOneBtn(i, pressed)
		if changed && !pressed {
			if btn == -1 {
				btn = i
			} else {
				btn = -2
			}
		}
	}

	return btn
}

func (s *shuttleXpress) updateState(state shuttleXpressState) {
	s.old = state
}

func (s *shuttleXpress) loop() {
	defer close(s.e)

	s.e <- Event{Type: EventTypeConnection, Value: ConnectionSuccess}
	defer func() { s.e <- Event{Type: EventTypeConnection, Value: ConnectionLost} }()

	var state shuttleXpressState
	var err error
	for {
		s.updateState(state)
		err = binary.Read(s.r, binary.BigEndian, &state)
		if err != nil {
			return
		}
		btn := s.updateButtons(state)
		if s.old.Ring != state.Ring {
			s.e <- Event{Type: EventTypeRing, Value: int(state.Ring)}
			continue
		}

		if s.old.Jog != state.Jog {
			diff := int(state.Jog) - int(s.old.Jog)
			if diff > 127 {
				diff -= 256
			} else if diff < -127 {
				diff += 256
			}

			if s.firstEvent == true {
				s.e <- Event{Type: EventTypeWheel, Value: diff}
				continue
			} else {
				s.firstEvent = true
			}
		}

		if btn != -1 {
			s.e <- Event{Type: EventTypeButton, Value: btn}
		}
	}
}

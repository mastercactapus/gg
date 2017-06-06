package grbl

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// State represents a possible state of the machine, it is returned as the first part
// of the `?` realtime command.
type State string

// Grbl machine states
const (
	StateUnknown      State = ""
	StateIdle         State = "Idle"
	StateRun          State = "Run"
	StateJog          State = "Jog"
	StateHoldActive   State = "Hold:1"
	StateHoldComplete State = "Hold:0"
	StateHome         State = "Home"
	StateAlarm        State = "Alarm"
	StateCheck        State = "Check"
	StateDoorAjar     State = "Door:1"
	StateDoorClosed   State = "Door:0"
	StateDoorClosing  State = "Door:3"
	StateDoorOpening  State = "Door:2"
	StateSleep        State = "Sleep"
)

type SpindleDirection int

const (
	SpindleDirectionCW SpindleDirection = iota
	SpindleDirectionCCW
)

type Status struct {
	State State

	Fields []string

	MPos []float64
	WPos []float64

	BlockBufferAvailable  int
	SerialBufferAvailable int

	Line int

	FeedSpeed    float64
	SpindleSpeed float64

	Pins struct {
		Probe      bool
		LimitX     bool
		LimitY     bool
		LimitZ     bool
		Door       bool
		Reset      bool
		FeedHold   bool
		CycleStart bool
	}

	WCO []float64

	FieldOVerrides struct {
		F            int
		R            int
		SpindleSpeed int
	}

	Aux struct {
		SpindleOn        bool
		SpindleDirection SpindleDirection
		CoolantFlood     bool
		CoolantMist      bool
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

func parseMachineStatus(data string) (*Status, error) {
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
			return nil, errors.Wrap(err, "parse machine status")
		}
	}

	return &s, nil
}

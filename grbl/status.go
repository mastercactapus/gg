package grbl

// State represents a possible state of the machine, it is returned as the first part
// of the `?` realtime command.
type State string

// Grbl machine states
const (
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

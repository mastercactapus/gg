package log

import "github.com/mastercactapus/gg/gcode"

//go:generate stringer -type Direction

// Direction is used to differentiate the data-flow direction of SerialData.
type Direction int

// Directions can be send (tx) or receive (rx)
const (
	DirectionSend Direction = iota
	DirectionRecv
)

// The Node interface is implemented by all node types.
type Node interface {
	Pos() Pos
	End() Pos
}
type node struct {
	tok      Token
	lit      string
	pos, end Pos
}

// Pos returns the starting position of the Node.
func (n node) Pos() Pos {
	return n.pos
}

// End returns the ending position of the Node.
func (n node) End() Pos {
	return n.end
}

// A Comment is a non-functional annotation in a log.
//
// Note: Parser.Parse() will only return a Comment if
// PreserveComments is set to true in the ParserConfig.
type Comment struct {
	Node
	Value string
}

// A Flag captures program settings and options.
type Flag struct {
	Node

	Name  string
	Value string
}

// GCode represents a single line of GCode words.
type GCode struct {
	Node

	Line gcode.Line
}

// Coordinates are used to log positions of the machine.
//
// The most common use is the `ZERO` position, that logs the
// work-zero position in machine coords.
// This is used to resume (after re-homing) in certain cases (e.g. breaker flipped).
type Coordinates struct {
	Node

	ID     string
	Values []float64
}

// SerialData is a log of functional data sent over the wire between the CNC controller and software.
//
// Generally, only GCode and confirmations are logged, and stateful data, like mode or jogging, is omitted.
type SerialData struct {
	Node

	Direction Direction
	Data      string
}

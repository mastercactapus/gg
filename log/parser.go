package log

import (
	"fmt"
	"io"
	"strconv"
)

// Parser uses a Scanner to provide a stream of Nodes via
// the Parse() method.
type Parser struct {
	s   *Scanner
	buf node
	n   int
	c   ParserConfig
}

// ParserConfig allows setting options for a Parser.
type ParserConfig struct {
	PreserveComments bool // If true, the Parse() method will emit rather than skip Comments.
}

// IllegalTokenError is returned if an IllegalToken is encountered.
type IllegalTokenError struct {
	Token   Token
	Literal string
	Pos     Pos
}

// UnexpectedTokenError is returned when the Token type does not match what is expected.
//
// For instance, GCodes must have a letter, and numeric value (e.g. for `G21` the numeric
// value would be `21`). If in the log `GG` was on a line, the second `G` would cause an
// UnexpectedTokenError.
type UnexpectedTokenError struct {
	Token    Token
	Expected string
	Literal  string
	Pos      Pos
}

// SyntaxError is returned when a token is unable to be parsed into a value.
//
// Consider the example: `G0 X12.3.1`
// The numeric value tied to `X` would be `12.3.1`, however this is not a valid number.
// In this case `12.3.1` would cause a SyntaxError.
type SyntaxError struct {
	Pos     Pos
	Token   Token
	Literal string
	Cause   error
}

// Error implements the error interface.
func (i IllegalTokenError) Error() string {
	return fmt.Sprintf(
		"illegal token '%s' at line %d col %d",
		i.Literal, i.Pos.Line, i.Pos.Col,
	)
}

// Error implements the error interface.
func (s SyntaxError) Error() string {
	return fmt.Sprintf(
		"syntax error at line %d col %d parsing '%s' as %s: %s",
		s.Pos.Line, s.Pos.Col, s.Literal, s.Token.String(), s.Cause.Error(),
	)
}

// Error implements the error interface.
func (e UnexpectedTokenError) Error() string {
	return fmt.Sprintf(
		"unexpected %s '%s' at line %d col %d -- expected %s",
		e.Token.String(), e.Literal, e.Pos.Line, e.Pos.Col, e.Expected,
	)
}

func (n node) illegalErr() error {
	return &IllegalTokenError{Token: n.tok, Literal: n.lit, Pos: n.pos}
}
func (n node) unexpectedErr(expected string) error {
	return &UnexpectedTokenError{Token: n.tok, Literal: n.lit, Pos: n.pos, Expected: expected}
}
func (n node) syntaxErr(cause error) error {
	return &SyntaxError{Token: n.tok, Literal: n.lit, Pos: n.pos, Cause: cause}
}

// NewParser creates a new Parser consuming the io.Reader, and using the optional ParserConfig.
func NewParser(r io.Reader, c *ParserConfig) *Parser {
	p := &Parser{s: NewScanner(r)}
	if c != nil {
		p.c = *c
	}
	return p
}

func (p *Parser) scan() node {
	if p.n == 0 {
		p.buf.pos = p.s.Pos()
		p.buf.tok, p.buf.lit = p.s.Scan()
		p.buf.end = p.s.Pos()
	} else {
		p.n = 0
	}

	return p.buf
}

func (p *Parser) unscan() { p.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (n node) {
	n = p.scan()
	if n.tok == TokenWhitespace {
		n = p.scan()
	}
	return n
}

func (p *Parser) scanFlag() (*Flag, error) {
	n := p.scan()
	start := n.pos
	var f Flag
	f.Name = n.lit[1:]
	n = p.scanIgnoreWhitespace()
	if n.tok != TokenEquals {
		return nil, n.unexpectedErr("'='")
	}
	n = p.scanIgnoreWhitespace()
	if n.tok != TokenString {
		return nil, n.unexpectedErr("a quoted string")
	}
	s, err := strconv.Unquote(n.lit)
	if err != nil {
		return nil, n.syntaxErr(err)
	}

	f.Value = s
	f.Node = node{pos: start, end: n.end}
	return &f, nil
}

func (p *Parser) scanGCode() (*GCode, error) {
	n := p.scan()
	start := n.pos
	var end Pos
	words := make([]Word, 0, 10)
	var w Word
	var err error
	for n.tok == TokenWord {
		w.Type = n.lit[0]
		n = p.scanIgnoreWhitespace()
		if n.tok != TokenNumber {
			return nil, n.unexpectedErr("a numeric value")
		}

		w.Value, err = strconv.ParseFloat(n.lit, 64)
		if err != nil {
			return nil, n.syntaxErr(err)
		}
		words = append(words, w)
		end = n.end
		n = p.scanIgnoreWhitespace()
	}
	p.unscan()

	return &GCode{
		Node:  node{pos: start, end: end},
		Words: words,
	}, nil
}

func (p *Parser) scanCoordinates() (*Coordinates, error) {
	n := p.scan()
	name := n.lit[1:]
	start := n.pos
	var end Pos
	n = p.scanIgnoreWhitespace()
	if n.tok != TokenLBrace {
		return nil, n.unexpectedErr("'{'")
	}
	vals := make([]float64, 0, 10)
	var v float64
	var err error
	for {
		n = p.scanIgnoreWhitespace()
		if n.tok != TokenNumber {
			return nil, n.unexpectedErr("a numeric value")
		}

		v, err = strconv.ParseFloat(n.lit, 64)
		if err != nil {
			return nil, n.syntaxErr(err)
		}
		vals = append(vals, v)

		n = p.scanIgnoreWhitespace()
		if n.tok == TokenRBrace {
			end = n.end
			break
		} else if n.tok == TokenComma {
			continue
		}

		return nil, n.unexpectedErr("'}' or ','")
	}

	return &Coordinates{
		ID:     name,
		Values: vals,
		Node:   node{pos: start, end: end},
	}, nil
}

func (p *Parser) scanSerial() (*SerialData, error) {
	n := p.scan()
	var d Direction
	switch n.tok {
	case TokenGT:
		d = DirectionSend
	case TokenLT:
		d = DirectionRecv
	default:
		panic("scanSerial got unknown token: " + n.tok.String())
	}
	start := n.pos
	n = p.scanIgnoreWhitespace()
	if n.tok != TokenString {
		return nil, n.unexpectedErr("a quoted string")
	}
	s, err := strconv.Unquote(n.lit)
	if err != nil {
		return nil, n.syntaxErr(err)
	}
	return &SerialData{
		Data:      s,
		Direction: d,
		Node:      node{pos: start, end: n.end},
	}, nil
}

// Parse will return the next Node, or an error.
func (p *Parser) Parse() (Node, error) {
	n := p.scan()
	for (n.tok == TokenWhitespace || n.tok == TokenNewLine) || (n.tok == TokenComment && !p.c.PreserveComments) {
		n = p.scan()
	}
	switch n.tok {
	case TokenComment:
		return &Comment{Node: n, Value: n.lit[1:]}, nil
	case TokenFlag:
		p.unscan()
		return p.scanFlag()
	case TokenWord:
		p.unscan()
		return p.scanGCode()
	case TokenIdentifier:
		p.unscan()
		return p.scanCoordinates()
	case TokenGT, TokenLT:
		p.unscan()
		return p.scanSerial()
	case TokenEOF:
		return nil, io.EOF
	case TokenIllegal:
		return nil, n.illegalErr()
	}

	return nil, n.unexpectedErr("flag, gcode, send, recv, comment, or EOF")
}

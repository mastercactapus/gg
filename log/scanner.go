package log

import (
	"bufio"
	"bytes"
	"io"
)

var eof = rune(0)

// Pos represents the position within a log file
type Pos struct{ Line, Col int }

// A Scanner consumes an io.Reader of log data, and provides a stream
// of Tokens via the Scan() method.
type Scanner struct {
	r         *bufio.Reader
	line, col int
	lCol      int
	last      rune
	err       error
}

// NewScanner will create a Scanner consuming the io.Reader
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	if ch == '\n' {
		s.line++
		s.lCol = s.col
		s.col = 0
	} else {
		s.col++
	}
	s.last = ch

	return ch
}
func (s *Scanner) unread() {
	err := s.r.UnreadRune()
	if err != nil {
		return
	}
	if s.last == '\n' {
		s.line--
		s.col = s.lCol
	} else {
		s.col--
	}

	s.last = eof
}

// Pos will return the current position of the Scanner.
// To know the start and end Pos of a given token, call Pos before
// and after the Scan() method.
func (s *Scanner) Pos() Pos {
	return Pos{Line: s.line + 1, Col: s.col + 1}
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\r' || r == '\t'
}
func isLowerLetter(ch rune) bool {
	return ch >= 'a' && ch <= 'z'
}
func isUpperLetter(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}
func isLetter(ch rune) bool {
	return isLowerLetter(ch) || isUpperLetter(ch)
}
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}
func isID(ch rune) bool {
	return isUpperLetter(ch) || isDigit(ch)
}
func isNumber(ch rune) bool {
	return isDigit(ch) || ch == '.' || ch == '-'
}
func isFlag(ch rune) bool {
	return isLowerLetter(ch) || isDigit(ch) || ch == '-'
}
func isAlphaNum(ch rune) bool {
	return isLetter(ch) || isNumber(ch)
}
func isLine(ch rune) bool {
	return ch != '\n' && ch != '\r' && ch != eof
}

func (s *Scanner) scanType(t Token, check func(rune) bool) (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	var ch rune
	for {
		ch = s.read()
		if ch == eof {
			break
		} else if !check(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return t, buf.String()
}

func (s *Scanner) scanFlag() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	var ch rune
	ch = s.read()
	buf.WriteRune(ch)
	if !isLowerLetter(ch) && !isDigit(ch) {
		return TokenIllegal, buf.String()
	}

	for {
		ch = s.read()
		if ch == eof {
			break
		} else if !isFlag(ch) {
			s.unread()
			break
		}

		buf.WriteRune(ch)
	}

	return TokenFlag, buf.String()
}

func (s *Scanner) scanString() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	var ch rune
	var back bool
	for {
		back = !back && ch == '\\'
		ch = s.read()
		if ch == eof || !isLine(ch) {
			break
		}

		buf.WriteRune(ch)

		if !back && ch == '"' {
			break
		}
	}
	if ch != '"' {
		s.unread()
		return TokenUnterminatedString, buf.String()
	}
	return TokenString, buf.String()
}

// Scan will return the next Token and its literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanType(TokenWhitespace, isWhitespace)
	}

	if isLetter(ch) {
		return TokenWord, string(ch)
	}

	if isNumber(ch) {
		s.unread()
		return s.scanType(TokenNumber, isNumber)
	}

	switch ch {
	case '\n':
		return TokenNewLine, "\n"
	case '@':
		s.unread()
		return s.scanFlag()
	case '"':
		s.unread()
		return s.scanString()
	case '=':
		return TokenEquals, "="
	case ',':
		return TokenComma, ","
	case '{':
		return TokenLBrace, "{"
	case '}':
		return TokenRBrace, "}"
	case '>':
		return TokenGT, ">"
	case '<':
		return TokenLT, "<"
	case ';':
		s.unread()
		return s.scanType(TokenComment, isLine)
	case '_':
		s.unread()
		return s.scanType(TokenIdentifier, isID)
	case eof:
		return TokenEOF, ""
	}

	return TokenIllegal, string(ch)
}

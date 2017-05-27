package log

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mastercactapus/gg/gcode"
)

type FormatError struct {
	Type   string
	Value  string
	Reason string
}

func (f FormatError) Error() string {
	return fmt.Sprintf(
		"bad format for %s; '%s' was invalid because %s",
		f.Type, f.Value, f.Reason,
	)
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func commentString(value string) string {
	if value == "" {
		return ""
	}
	return " ; " + strings.Replace(value, "\n", " ", -1)
}

func (w *Writer) Comment(value string) error {
	_, err := io.WriteString(w.w, strings.TrimSpace(commentString(value))+"\n")
	return err
}

func (w *Writer) Flag(name, value, comment string) error {
	if len(name) == 0 {
		return &FormatError{Type: "Flag", Value: name + "=" + value, Reason: "name must not be empty"}
	}
	for _, ch := range name {
		if !isFlag(ch) {
			return &FormatError{Type: "Flag", Value: name + "=" + value, Reason: "name must only contain digits, lower-case letters, and hyphens"}
		}
	}
	if !isLowerLetter([]rune(name)[0]) && !isDigit([]rune(name)[0]) {
		return &FormatError{Type: "Flag", Value: name + "=" + value, Reason: "name must begin with lower-case letter or digit"}
	}

	_, err := io.WriteString(w.w, "@"+name+"="+strconv.Quote(value)+commentString(comment)+"\n")
	return err
}

func (w *Writer) GCode(l gcode.Line) error {
	if len(l) == 0 {
		return nil
	}

	_, err := io.WriteString(w.w, l.String()+"\n")
	return err
}

func (w *Writer) Coordinates(id string, coords []float64) error {
	for _, ch := range id {
		if !isID(ch) {
			return &FormatError{Type: "Coordinates", Value: id, Reason: "id must consist of only upper-case letters and digits"}
		}
	}

	if len(coords) == 0 {
		return &FormatError{Type: "Coordinates", Value: "coords", Reason: "must contain at least one coordinate"}
	}

	s := make([]string, len(coords))
	for i, v := range coords {
		s[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	_, err := io.WriteString(w.w, "_"+id+"{"+strings.Join(s, ",")+"}\n")
	return err
}

func (w *Writer) SerialSend(data string) error {
	_, err := io.WriteString(w.w, ">"+strconv.Quote(data)+"\n")
	return err
}
func (w *Writer) SerialRecv(data string) error {
	_, err := io.WriteString(w.w, "<"+strconv.Quote(data)+"\n")
	return err
}

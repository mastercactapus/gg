package log

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestScanner(t *testing.T) {
	data := `
; hi
@foo="bar";ok
G21
G0X1 Y2Z-3.4
_ZERO{1,2}
>"a"
<"b"
`
	s := NewScanner(bytes.NewBufferString(data))

	test := func(index, line, col int, tok Token, lit string) {
		t.Run(fmt.Sprintf("tokenL%dC%d", line, col), func(t *testing.T) {
			p := s.Pos()
			if p.Line != line {
				t.Errorf("line = %d; want %d", p.Line, line)
			}

			if p.Col != col {
				t.Errorf("col = %d; want %d", p.Col, col)
			}

			tk, l := s.Scan()
			if tk != tok {
				t.Errorf("token = %s; want %s", tk.String(), tok.String())
			}

			if l != lit {
				t.Errorf("literal = %s; want %s", strconv.Quote(l), strconv.Quote(lit))
			}
		})
	}

	expected := []struct {
		Line, Col int
		Token     Token
		Literal   string
	}{
		{1, 1, TokenNewLine, "\n"},
		{2, 1, TokenComment, "; hi"},
		{2, 5, TokenNewLine, "\n"},
		{3, 1, TokenFlag, "@foo"},
		{3, 5, TokenEquals, "="},
		{3, 6, TokenString, `"bar"`},
		{3, 11, TokenComment, ";ok"},
		{3, 14, TokenNewLine, "\n"},
		{4, 1, TokenWord, "G"},
		{4, 2, TokenNumber, "21"},
		{4, 4, TokenNewLine, "\n"},
		{5, 1, TokenWord, "G"},
		{5, 2, TokenNumber, "0"},
		{5, 3, TokenWord, "X"},
		{5, 4, TokenNumber, "1"},
		{5, 5, TokenWhitespace, " "},
		{5, 6, TokenWord, "Y"},
		{5, 7, TokenNumber, "2"},
		{5, 8, TokenWord, "Z"},
		{5, 9, TokenNumber, "-3.4"},
		{5, 13, TokenNewLine, "\n"},
		{6, 1, TokenIdentifier, "_ZERO"},
		{6, 6, TokenLBrace, "{"},
		{6, 7, TokenNumber, "1"},
		{6, 8, TokenComma, ","},
		{6, 9, TokenNumber, "2"},
		{6, 10, TokenRBrace, "}"},
		{6, 11, TokenNewLine, "\n"},
		{7, 1, TokenGT, ">"},
		{7, 2, TokenString, `"a"`},
		{7, 5, TokenNewLine, "\n"},
		{8, 1, TokenLT, "<"},
		{8, 2, TokenString, `"b"`},
		{8, 5, TokenNewLine, "\n"},
		{9, 1, TokenEOF, ""},
	}

	for i, e := range expected {
		test(i, e.Line, e.Col, e.Token, e.Literal)
	}
}

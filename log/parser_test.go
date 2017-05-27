package log

import (
	"bytes"
	"io"
	"strconv"
	"testing"
)

func TestParser(t *testing.T) {

	test := func(name, data string, f func(t *testing.T, n Node)) {
		t.Run(name, func(t *testing.T) {
			t.Logf("data = %s", strconv.Quote(data))
			p := NewParser(bytes.NewBufferString(data))
			n, err := p.Parse()
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
			f(t, n)
		})
	}

	test("GCode", `G21`, func(t *testing.T, n Node) {
		g, ok := n.(*GCode)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &GCode{})
		}
		if len(g.Words) != 1 {
			t.Fatalf("len(Words) = %d; want 1", len(g.Words))
		}
		if g.Words[0].Type != 'G' {
			t.Errorf("Words[0].Type = %c; want G", g.Words[0].Type)
		}
		if g.Words[0].Value != 21 {
			t.Errorf("Words[0].Value = %f; want 21", g.Words[0].Value)
		}
	})
	test("GCode", `G21 (hi) Y2`, func(t *testing.T, n Node) {
		g, ok := n.(*GCode)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &GCode{})
		}
		if len(g.Words) != 2 {
			t.Fatalf("len(Words) = %d; want 2", len(g.Words))
		}
		if g.Words[0].Type != 'G' {
			t.Errorf("Words[0].Type = %c; want G", g.Words[0].Type)
		}
		if g.Words[0].Value != 21 {
			t.Errorf("Words[0].Value = %f; want 21", g.Words[0].Value)
		}
		if g.Words[1].Type != 'Y' {
			t.Errorf("Words[1].Type = %c; want Y", g.Words[0].Type)
		}
		if g.Words[1].Value != 2 {
			t.Errorf("Words[1].Value = %f; want 2", g.Words[0].Value)
		}
	})

	test("Flag", `@foo="bar"`, func(t *testing.T, n Node) {
		f, ok := n.(*Flag)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Flag{})
		}
		if f.Name != "foo" {
			t.Errorf("Name = %s; want foo", f.Name)
		}
		if f.Value != "bar" {
			t.Errorf("Value = %s; want bar", f.Value)
		}
	})
	test("Flag", `@foo ="bar" ; what`, func(t *testing.T, n Node) {
		f, ok := n.(*Flag)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Flag{})
		}
		if f.Name != "foo" {
			t.Errorf("Name = %s; want foo", f.Name)
		}
		if f.Value != "bar" {
			t.Errorf("Value = %s; want bar", f.Value)
		}
	})
	test("Flag", `@foo = "1ft 2\"" ; what`, func(t *testing.T, n Node) {
		f, ok := n.(*Flag)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Flag{})
		}
		if f.Name != "foo" {
			t.Errorf("Name = %s; want foo", f.Name)
		}
		if f.Value != `1ft 2"` {
			t.Errorf("Value = %s; want %s", strconv.Quote(f.Value), strconv.Quote(`1ft 2"`))
		}
	})

	test("Coordinates", `_ZERO{1,3 , 4.5}`, func(t *testing.T, n Node) {
		c, ok := n.(*Coordinates)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Coordinates{})
		}
		if len(c.Values) != 3 {
			t.Fatalf("len(Values) = %d; want 2", len(c.Values))
		}
		if c.Values[0] != 1 {
			t.Errorf("Values[0] = %f; want 1", c.Values[0])
		}
		if c.Values[1] != 3 {
			t.Errorf("Values[1] = %f; want 3", c.Values[1])
		}
		if c.Values[2] != 4.5 {
			t.Errorf("Values[2] = %f; want 4.5", c.Values[1])
		}
	})

	test("SerialData", `>"foobar"`, func(t *testing.T, n Node) {
		d, ok := n.(*SerialData)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Coordinates{})
		}
		if d.Data != "foobar" {
			t.Errorf("Data = %s; want foobar", d.Data)
		}
		if d.Direction != DirectionSend {
			t.Errorf("Direction = %s; want DirectionSend", d.Direction.String())
		}
	})
	test("SerialData", `>"foobar;baz"`, func(t *testing.T, n Node) {
		d, ok := n.(*SerialData)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Coordinates{})
		}
		if d.Data != "foobar;baz" {
			t.Errorf("Data = %s; want foobar;baz", d.Data)
		}
		if d.Direction != DirectionSend {
			t.Errorf("Direction = %s; want DirectionSend", d.Direction.String())
		}
	})
	test("SerialData", `<"foobar"`, func(t *testing.T, n Node) {
		d, ok := n.(*SerialData)
		if !ok {
			t.Fatalf("type = %T; want %T", n, &Coordinates{})
		}
		if d.Data != "foobar" {
			t.Errorf("Data = %s; want foobar", d.Data)
		}
		if d.Direction != DirectionRecv {
			t.Errorf("Direction = %s; want DirectionRecv", d.Direction.String())
		}
	})

}

func TestParser_EOF(t *testing.T) {
	p := NewParser(bytes.NewBufferString(""))
	_, err := p.Parse()
	if err != io.EOF {
		t.Errorf("err = %v; want %v", err, io.EOF)
	}

	p = NewParser(bytes.NewBufferString("   \n \t"))
	_, err = p.Parse()
	if err != io.EOF {
		t.Errorf("err = %v; want %v", err, io.EOF)
	}
}

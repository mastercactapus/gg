package ui

import termbox "github.com/nsf/termbox-go"

const (
	radioChecked   = "◈"
	radioUnchecked = "◇"

	checked   = "☑"
	unchecked = "☐"
)

type Checkbox struct {
	X, Y        int
	Width       int
	Text        string
	Enabled     bool
	Checked     bool
	Radio       bool
	OnClickFunc func(x, y int, newState bool)
}

func (c *Checkbox) OnClick(x, y int) {
	if !c.Enabled {
		return
	}
	if c.OnClickFunc == nil {
		return
	}
	c.OnClickFunc(x, y, !c.Checked)
}
func (c *Checkbox) Draw(r Renderer) {
	w := c.Width
	minW := len(c.Text) + 2
	if minW > w {
		w = minW
	}
	sw, sh := r.Size()
	x, y, w, _ := StandardSize(c.X, c.Y, w, 1, sw, sh)

	var icon string
	switch {
	case c.Checked && c.Radio:
		icon = radioChecked
	case c.Checked && !c.Radio:
		icon = checked
	case !c.Checked && c.Radio:
		icon = radioUnchecked
	case !c.Checked && !c.Radio:
		icon = radioChecked
	}

	var fg, bg termbox.Attribute
	if !c.Enabled {
		fg = termbox.ColorBlack
	} else {
		if c.Checked {
			fg = termbox.ColorYellow
		} else {
			fg = termbox.ColorWhite
		}
		bg = termbox.ColorBlue
	}

	rs := []rune(icon + " " + c.Text)
	putRunesA(r, x, y, rs, fg, bg)
}

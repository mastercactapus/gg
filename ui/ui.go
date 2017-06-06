package ui

import (
	"sync"

	termbox "github.com/nsf/termbox-go"
)

type RenderFunc func() []Control

type UI struct {
	mx       sync.Mutex
	r        RenderFunc
	renderCh chan struct{}
	quitCh   chan struct{}
	eventCh  chan termbox.Event

	rendered []renderedControl
}
type renderedControl struct {
	r Rect
	c Control
}

func NewUI(r RenderFunc) (*UI, error) {
	err := termbox.Init()
	if err != nil {
		return nil, err
	}
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	u := &UI{
		r:        r,
		renderCh: make(chan struct{}, 1),
		quitCh:   make(chan struct{}),
		eventCh:  make(chan termbox.Event, 1),
	}
	go u.eventLoop()
	return u, nil
}
func (ui *UI) Render() {
	select {
	case ui.renderCh <- struct{}{}:
	default:
	}
}
func (ui *UI) eventLoop() {
	for {
		select {
		case <-ui.quitCh:
			return
		default:
		}
		ui.eventCh <- termbox.PollEvent()
	}
}
func (ui *UI) MainLoop() {
	termbox.Clear(0, 0)
	ui.renderControls()
	defer close(ui.quitCh)
	for {
		select {
		case <-ui.renderCh:
		case ev := <-ui.eventCh:
			switch ev.Type {
			case termbox.EventMouse:
				if ev.Key != termbox.MouseLeft {
					continue
				}
				ui.processClick(ev.MouseX, ev.MouseY)
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyCtrlC:
					return
				case termbox.KeyCtrlL:
					termbox.Clear(0, 0)
				}

			case termbox.EventResize:
				termbox.Clear(0, 0)
			}
		}

		ui.renderControls()
	}
}
func (ui *UI) processClick(x, y int) {
	for _, c := range ui.rendered {
		cc, ok := c.c.(Clickable)
		if !ok {
			continue
		}
		if !c.r.Contains(x, y) {
			continue
		}
		cc.OnClick(c.r.Translate(x, y))
	}
}
func (ui *UI) Lock() {
	ui.mx.Lock()
}
func (ui *UI) Unlock() {
	ui.mx.Unlock()
}
func (ui *UI) Close() error {
	termbox.Close()
	return nil
}

func (UI) Size() (int, int) {
	return termbox.Size()
}
func (ui *UI) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x, y, ch, fg, bg)
}

func (ui *UI) renderControls() {
	sw, sh := termbox.Size()
	bounds := Rect{Left: 0, Top: 0, Right: sw, Bottom: sh}
	b := newBoundedRenderer(ui, bounds)
	r := ui.r()
	ui.rendered = ui.rendered[:0]
	for _, c := range r {
		ui.rendered = append(ui.rendered, renderedControl{
			c: c,
			r: b.RenderChild(bounds, c),
		})
	}
	termbox.Flush()
}

type Screen interface {
	Size() (int, int)
	SetCell(x, y int, ch rune, fg, bg termbox.Attribute)
}
type Renderer interface {
	Screen
	RenderChild(bounds Rect, c Control) Rect
}

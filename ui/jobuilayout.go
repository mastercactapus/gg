package ui

import (
	"github.com/mastercactapus/gg/grbl"
	termbox "github.com/nsf/termbox-go"
)

func (j *JobUI) render() []Control {
	serialMode := j.c.SerialMode()
	j.renderSync()
	defer j.renderSync()

	return []Control{
		&Group{
			Title: "GCODE -- " + j.statusText(),
			Width: 40,
			Controls: []Control{
				&Button{
					X:           5,
					Y:           1,
					Text:        "Check",
					Enabled:     j.s.State == grbl.StateIdle,
					OnClickFunc: func(x, y int) { j.actionCh <- actionCheckCode },
				},
				&Button{
					X:           16,
					Y:           1,
					Text:        "Run",
					Enabled:     j.s.State == grbl.StateIdle && j.checked,
					OnClickFunc: func(x, y int) { j.actionCh <- actionRunJob },
				},
				&Button{
					X:           25,
					Y:           1,
					Text:        "Stop",
					Enabled:     j.isRunning(),
					OnClickFunc: func(x, y int) { j.actionCh <- actionStopJob },
				},
				j.status(),
				&j.v,
			},
		},
		&Group{
			Title:  j.machineStatusText(),
			Width:  60,
			Height: 20,
			X:      40,
			Controls: []Control{
				&Status{X: 1, Y: 3, Status: &j.s, Settings: &j.settings},

				&Group{
					Width:  40,
					Height: 20,
					Clear:  true,
					X:      21,
					Controls: []Control{
						j.shuttle(),
						&Button{Y: 0, Text: "All 0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.zeroAxis <- '_' },
						},
						&Button{Y: 2, X: 10, Text: "X0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.zeroAxis <- 'X' },
						},
						&Button{Y: 2, X: 0, Text: "Go X0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.goZeroAxis <- 'X' },
						},
						&Button{Y: 3, X: 10, Text: "Y0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.zeroAxis <- 'Y' },
						},
						&Button{Y: 3, X: 0, Text: "Go Y0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.goZeroAxis <- 'Y' },
						},
						&Button{Y: 4, X: 10, Text: "Z0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.zeroAxis <- 'Z' },
						},
						&Button{Y: 4, X: 0, Text: "Go Z0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.goZeroAxis <- 'Z' },
						},

						&Button{Y: 6, Text: "Go ⌂", Enabled: j.s.State == grbl.StateIdle || j.s.State == grbl.StateAlarm,
							OnClickFunc: func(int, int) { j.goZeroAxis <- 'H' },
						},
						&Button{Y: 7, Text: "Go 0", Enabled: j.s.State == grbl.StateIdle,
							OnClickFunc: func(int, int) { j.goZeroAxis <- '_' },
						},

						&DblButton{
							Y:           10,
							Text:        "⬅ X",
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							OnClickFunc: func(int, int) { j.JogStep('x') },
						},
						&DblButton{
							Y:           10,
							X:           14,
							Text:        "X ➡",
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							OnClickFunc: func(int, int) { j.JogStep('X') },
						},
						&Button{
							Y:           10,
							X:           6,
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							Text:        "Y ⬆",
							OnClickFunc: func(int, int) { j.JogStep('Y') },
						},
						&Button{
							Y:           12,
							X:           6,
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							Text:        "Y ⬇",
							OnClickFunc: func(int, int) { j.JogStep('y') },
						},
						&Button{
							Y:           10,
							X:           20,
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							Text:        "Z⊕",
							OnClickFunc: func(int, int) { j.JogStep('Z') },
						},
						&Button{
							Y:           12,
							X:           20,
							Enabled:     j.s.State == grbl.StateIdle || j.s.State == grbl.StateJog,
							Text:        "Z⊖",
							OnClickFunc: func(int, int) { j.JogStep('z') },
						},

						&Text{
							Y:     14,
							Lines: []string{"Move By:"},
						},
						&Checkbox{
							Radio:       true,
							Enabled:     true,
							Y:           14,
							X:           9,
							Text:        "10",
							Checked:     j.jogStep == 10,
							OnClickFunc: func(int, int, bool) { j.setJogStep <- 10 },
						},
						&Checkbox{
							Radio:       true,
							Enabled:     true,
							Y:           14,
							X:           14,
							Text:        "1",
							Checked:     j.jogStep == 1,
							OnClickFunc: func(int, int, bool) { j.setJogStep <- 1 },
						},
						&Checkbox{
							Radio:       true,
							Enabled:     true,
							Y:           14,
							X:           18,
							Text:        "0.1",
							Checked:     j.jogStep == 0.1,
							OnClickFunc: func(int, int, bool) { j.setJogStep <- 0.1 },
						},
						&Checkbox{
							Radio:       true,
							Enabled:     true,
							Y:           15,
							X:           9,
							Text:        "0.01",
							Checked:     j.jogStep == 0.01,
							OnClickFunc: func(int, int, bool) { j.setJogStep <- 0.01 },
						},
						&Checkbox{
							Radio:       true,
							Enabled:     true,
							Y:           15,
							X:           16,
							Text:        "0.001",
							Checked:     j.jogStep == 0.001,
							OnClickFunc: func(int, int, bool) { j.setJogStep <- 0.001 },
						},
					},
				},
			},
		},
		&Group{
			Title:  "Configuration",
			Width:  20,
			Height: 6,
			X:      100,
			Controls: []Control{
				&Text{Lines: []string{"Serial Mode"}},
				&Checkbox{
					X:           1,
					Y:           1,
					Text:        "Send-Response",
					Enabled:     true,
					Radio:       true,
					Checked:     serialMode == grbl.ModeSendResponse,
					OnClickFunc: func(x, y int, newState bool) { j.c.SetSerialMode(grbl.ModeSendResponse) },
				},
				&Checkbox{
					X:           1,
					Y:           2,
					Text:        "Character Count",
					Enabled:     true,
					Radio:       true,
					Checked:     serialMode == grbl.ModeCharacterCount,
					OnClickFunc: func(x, y int, newState bool) { j.c.SetSerialMode(grbl.ModeCharacterCount) },
				},
			},
		},
		&Group{
			Title:    "Logs",
			X:        40,
			Y:        20,
			Controls: []Control{j.l},
		},
	}
}

func (j *JobUI) shuttle() Control {
	if j.shuttleBusted {
		return &Group{
			X: 17, Y: 0,
			Height:   3,
			Width:    17,
			NoBorder: true,
			Controls: []Control{
				&Text{X: 3, Lines: []string{"ShuttleXpress"}, FG: termbox.ColorRed | termbox.AttrBold},
			},
		}
	}
	if !j.shuttleConnected {
		return nil
	}

	var axis string
	if j.shuttleAxis == 0 {
		axis = " "
	} else {
		axis = string(j.shuttleAxis)
	}

	return &Group{
		X: 17, Y: 0,
		Height:   3,
		Width:    17,
		NoBorder: true,
		Controls: []Control{
			&Text{X: 3, Lines: []string{"ShuttleXpress"}, FG: termbox.ColorCyan | termbox.AttrBold},
			&Text{X: 3, Y: 1, Lines: []string{"Axis:"}},
			&Text{X: 9, Y: 1, Lines: []string{"   " + axis},
				FG: termbox.ColorWhite | termbox.AttrBold,
				BG: termbox.ColorMagenta,
			},
		},
	}
}

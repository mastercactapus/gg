package ui

import "github.com/mastercactapus/gg/grbl"

func (j *JobUI) render() []Control {
	serialMode := j.c.SerialMode()
	j.renderSync()
	defer j.renderSync()

	return []Control{
		&Group{
			Title: "GCODE -- " + j.statusText(),
			Width: 40,
			Y:     1,
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
			Width:  80,
			Height: 10,
			X:      40,
			Controls: []Control{
				&Status{X: 1, Y: 1, Status: &j.s},
			},
		},
		&Group{
			Title:  "Configuration",
			Width:  20,
			Height: 6,
			X:      40,
			Y:      10,
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
	}
}

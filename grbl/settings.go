package grbl

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	StepPulse           time.Duration
	StepIdleDelay       time.Duration
	StepPortInvert      PortInvertMask
	DirectionPortInvert PortInvertMask
	StepEnableInvert    bool
	LimitPinsInvert     bool
	ProbePinInvert      bool
	StatusReport        struct {
		MPos       bool
		BufferData bool
		Inches     bool
	}
	JunctionDeviation     Distance
	ArcTolerance          Distance
	SoftLimits            bool
	HardLimits            bool
	Homing                bool
	HomingDirectionInvert PortInvertMask
	HomingFeed            Rate
	HomingSeek            Rate
	HomingDebounce        time.Duration
	HomingPullOff         Distance
	MaxSpindleSpeed       int
	MinSpindleSpeed       int
	LaserMode             bool
	StepsPerMillimeter    struct {
		X, Y, Z float64
	}
	MaxRate         struct{ X, Y, Z Rate }
	MaxAcceleration struct{ X, Y, Z Accel }
	MaxTravel       struct{ X, Y, Z Distance }
}

type PortInvertMask struct{ X, Y, Z bool }

type settingsParser struct {
	err error
}

func (s *settingsParser) parseInt(val string) int {
	if s.err != nil {
		return 0
	}
	var v int64
	v, s.err = strconv.ParseInt(val, 10, 64)
	if s.err != nil {
		return 0
	}
	return int(v)
}
func (s *settingsParser) parseFloat(val string) float64 {
	if s.err != nil {
		return 0
	}
	var v float64
	v, s.err = strconv.ParseFloat(val, 64)
	if s.err != nil {
		return 0
	}
	return v
}
func (s *settingsParser) parsePIM(val string) PortInvertMask {
	v := s.parseInt(val)
	return PortInvertMask{
		X: v&(1<<0) != 0,
		Y: v&(1<<1) != 0,
		Z: v&(1<<2) != 0,
	}
}
func (s *settingsParser) parseDur(val string, unit time.Duration) time.Duration {
	v := s.parseFloat(val)
	return time.Duration(float64(unit) * v)
}
func (s *settingsParser) parseDist(val string, unit Distance) Distance {
	v := s.parseFloat(val)
	return Distance(float64(unit) * v)
}
func (s *settingsParser) parseRate(val string, d Distance, t time.Duration) Rate {
	return s.parseDist(val, d).Rate(t)
}
func (s *settingsParser) parseAccel(val string, d Distance, t1, t2 time.Duration) Accel {
	return s.parseRate(val, d, t1).Accel(t2)
}

func (s *Settings) parseSetting(data []byte) {
	log.Println(string(data))
	l := strings.TrimSpace(string(data))
	p := &settingsParser{}
	v := strings.SplitN(l, "=", 2)
	switch v[0] {
	case "$0":
		s.StepPulse = p.parseDur(v[1], time.Microsecond)
	case "$1":
		s.StepIdleDelay = p.parseDur(v[1], time.Millisecond)
	case "$2":
		s.StepPortInvert = p.parsePIM(v[1])
	case "$3":
		s.DirectionPortInvert = p.parsePIM(v[1])
	case "$4":
		s.StepEnableInvert = v[1] == "1"
	case "$5":
		s.LimitPinsInvert = v[1] == "1"
	case "$6":
		s.ProbePinInvert = v[1] == "1"
	case "$10":
		i := p.parseInt(v[1])
		s.StatusReport.MPos = i&(1<<0) != 0
		s.StatusReport.BufferData = i&(1<<1) != 0
	case "$11":
		s.JunctionDeviation = p.parseDist(v[1], Millimeter)
	case "$12":
		s.ArcTolerance = p.parseDist(v[1], Millimeter)
	case "$13":
		s.StatusReport.Inches = v[1] == "1"
	case "$20":
		s.SoftLimits = v[1] == "1"
	case "$21":
		s.HardLimits = v[1] == "1"
	case "$22":
		s.Homing = v[1] == "1"
	case "$23":
		s.HomingDirectionInvert = p.parsePIM(v[1])
	case "$24":
		s.HomingFeed = p.parseRate(v[1], Millimeter, time.Minute)
	case "$25":
		s.HomingSeek = p.parseRate(v[1], Millimeter, time.Minute)
	case "$26":
		s.HomingDebounce = p.parseDur(v[1], time.Millisecond)
	case "$27":
		s.HomingPullOff = p.parseDist(v[1], Millimeter)
	case "$30":
		s.MaxSpindleSpeed = p.parseInt(v[1])
	case "$31":
		s.MinSpindleSpeed = p.parseInt(v[1])
	case "$32":
		s.LaserMode = v[1] == "1"
	case "$100":
		s.StepsPerMillimeter.X = p.parseFloat(v[1])
	case "$101":
		s.StepsPerMillimeter.Y = p.parseFloat(v[1])
	case "$102":
		s.StepsPerMillimeter.Z = p.parseFloat(v[1])
	case "$110":
		s.MaxRate.X = p.parseRate(v[1], Millimeter, time.Minute)
	case "$111":
		s.MaxRate.Y = p.parseRate(v[1], Millimeter, time.Minute)
	case "$112":
		s.MaxRate.Z = p.parseRate(v[1], Millimeter, time.Minute)
	case "$120":
		s.MaxAcceleration.X = p.parseAccel(v[1], Millimeter, time.Second, time.Second)
	case "$121":
		s.MaxAcceleration.Y = p.parseAccel(v[1], Millimeter, time.Second, time.Second)
	case "$122":
		s.MaxAcceleration.Z = p.parseAccel(v[1], Millimeter, time.Second, time.Second)
	case "$130":
		s.MaxTravel.X = p.parseDist(v[1], Millimeter)
	case "$131":
		s.MaxTravel.Y = p.parseDist(v[1], Millimeter)
	case "$132":
		s.MaxTravel.Z = p.parseDist(v[1], Millimeter)
	}
	if p.err != nil {
		log.Printf("settings parse error for '%s': %v\n", l, p.err)
	}
}

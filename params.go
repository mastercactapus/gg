package gg

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type unitFlag struct {
	set   bool
	value float64
}

var paramNames []string

var (
	unitRx         = regexp.MustCompile(`^(\d+)\s*(mm|cm|inch|inches|in|"|'|ft|foot|feet)\.?$`)
	compoundUnitRx = regexp.MustCompile(`^(\d+)\s*(?:ft|foot|feet|')\.?\s*(\d+)\s*(?:in|inch|inches|")\.?$`)
)

// parseUnitString will attempt to parse a string value and convert it to mm.
func parseUnitString(s string) (float64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	m := compoundUnitRx.FindStringSubmatch(s)
	if m != nil {
		ft, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, errors.Wrap(err, "parse measurement value "+m[1])
		}
		in, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return 0, errors.Wrap(err, "parse measurement value "+m[2])
		}
		return ft*Foot + in*Inch, nil
	}

	m = unitRx.FindStringSubmatch(s)
	if m != nil {
		val, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, errors.Wrap(err, "parse measurement value "+m[1])
		}
		switch m[2] {
		case "mm":
			return val, nil
		case "cm":
			return val * CM, nil
		case "in", "\"", "inch", "inches":
			return val * Inch, nil
		case "ft", "'", "foot", "feet":
			return val * Foot, nil
		}
	}

	return 0, fmt.Errorf("failed to parse measurement '%s'", s)
}

func (sf *unitFlag) Set(s string) error {
	val, err := parseUnitString(s)
	if err != nil {
		return err
	}
	sf.set = true
	sf.value = val
	return nil
}
func (sf unitFlag) String() string {
	if !sf.set {
		return ""
	}

	return strconv.FormatFloat(sf.value, 'f', -1, 64) + "mm"
}

// ParamUnit will define a new unit/measurement program parameter with the given name and description.
// The parameter is requried to be set before the job can run.
func ParamUnit(name, description string) *float64 {
	f := &unitFlag{}
	flag.Var(f, name, description)
	paramNames = append(paramNames, name)
	return &f.value
}

// ParamUnitD will define a new unit/measurement program parameter with the given name and description.
// The default value will be used unless configured.
func ParamUnitD(name, description string, defaultValue float64) *float64 {
	f := &unitFlag{
		set:   true,
		value: defaultValue,
	}
	flag.Var(f, name, description)
	paramNames = append(paramNames, name)
	return &f.value
}

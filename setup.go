package gg

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/mastercactapus/gg/gcode"
	"github.com/mastercactapus/gg/log"
)

var (
	logFile = flag.String("log", "job.log", "Log file to use when in job mode -- will not overwrite unless -f is set.")
	run     = flag.Bool("run", false, "Run a job. Instead of printing GCode, launches the UI to execute it.")
	resume  = flag.Bool("resume", false, "Resume an existing log (implies -run).")
	l       *log.Writer
)

func failf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
	fmt.Fprintln(os.Stderr)
	flag.Usage()
	os.Exit(1)
}

func Run(f func()) {
	if *resume {
		return
	}
	err := l.Comment("Run(): Generate GCode")
	if err != nil {
		failf("failed to write to log: %v", err)
	}
	print(gcode.Line{{Type: 'G', Value: 21}})
	f()

	if *run {
		for _, line := range lines {
			err = l.GCode(line)
			if err != nil {
				failf("failed to write gcode to log: %v", err)
			}
		}
	} else {
		for _, l := range lines {
			fmt.Println(l.String())
		}
	}

}

func resumeState(r io.Reader) error {
	p := log.NewParser(r)

	node, err := p.Parse()
	for err == nil {
		switch n := node.(type) {
		case *log.Flag:
			err = flag.Set(n.Name, n.Value)
		case *log.GCode:
			lines = append(lines, n.Line)
		}
		if err != nil {
			return err
		}
		node, err = p.Parse()
	}
	if err != io.EOF {
		return err
	}

	return l.Comment("Resume " + time.Now().String())
}

// Setup will parse parameters, ask for input (where required) and make things
// ready to start processing G-Code commands.
func Setup(c Config) {
	flag.Parse()

	var flags int
	if *resume {
		flag.Set("run", "true")
		flags = os.O_RDWR
	} else if *run {
		_, err := os.Stat(*logFile)
		if err == nil {
			failf("cannot start new job over existing log: %s exists", *logFile)
		}
		if err != nil && !os.IsNotExist(err) {
			failf("error checking log file: %v", err)
		}
		flags = os.O_RDWR | os.O_CREATE
	}

	if *resume {
		flag.Visit(func(f *flag.Flag) {
			if _, ok := f.Value.(SavableValue); !ok {
				return
			}
			failf("%s was set; do not set flags when resuming, re-generate GCode instead", f.Name)
		})
	}

	if *run {
		fd, err := os.OpenFile(*logFile, flags, 0666)
		if err != nil {
			failf("failed to open log file: %v", err)
		}
		l = log.NewWriter(fd)
		if *resume {
			err = resumeState(fd)
			if err != nil {
				failf("failed to resume state: %v", err)
			}
		}
	} else {
		l = log.NewWriter(ioutil.Discard)
	}

	var missing bool
	for _, p := range paramNames {
		if !flag.Lookup(p).Value.(*unitFlag).set {
			missing = true
			fmt.Fprintf(os.Stderr, "required flag was not set: -%s\n", p)
		}
	}
	if missing {
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	err := l.Comment("Setup(): " + c.Name)
	if err != nil {
		failf("failed to log to file: %v", err)
	}

	if !*resume {
		// save current parameter values
		flag.VisitAll(func(f *flag.Flag) {
			if _, ok := f.Value.(SavableValue); !ok {
				return
			}
			err = l.Flag(f.Name, f.Value.String(), "default: "+f.DefValue)
			if err != nil {
				failf("failed to log param: %v", err)
			}
		})
	}
}

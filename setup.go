package gg

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	ll "log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"time"

	serial "go.bug.st/serial.v1"

	"github.com/mastercactapus/gg/gcode"
	"github.com/mastercactapus/gg/grbl"
	"github.com/mastercactapus/gg/log"
	"github.com/mastercactapus/gg/ui"
	termbox "github.com/nsf/termbox-go"
)

var (
	logFile = flag.String("log", "job.log", "Log file to use when in job mode -- will not overwrite unless -f is set.")
	run     = flag.Bool("run", false, "Run a job. Instead of printing GCode, launches the UI to execute it.")
	port    = flag.String("port", "", "Serial port to use.")
	rate    = flag.Int("b", 115200, "Baudrate of the serial port.")
	resume  = flag.Bool("resume", false, "Resume an existing log (implies -run).")
	l       *log.Writer
)

func init() {
	c, err := net.Dial("tcp", ":3006")
	if err != nil {
		panic(err)
	}
	ll.SetOutput(c)
	ll.Println("START")
}

func failf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
	fmt.Fprintln(os.Stderr)
	flag.Usage()
	os.Exit(1)
}

type logger struct {
	io.ReadWriteCloser
	rw io.Writer
	r  io.Writer
	w  io.Writer
}

var rep = strings.NewReplacer("\n", "\\n", "\r", "\\r")

func fmtLine(str string) string {
	return rep.Replace(str) + "\n"
}

func (l *logger) Write(p []byte) (int, error) {
	io.WriteString(l.rw, "W: "+fmtLine(string(p)))
	io.WriteString(l.w, "W: "+fmtLine(string(p)))
	return l.ReadWriteCloser.Write(p)
}
func (l *logger) Read(p []byte) (int, error) {
	n, err := l.ReadWriteCloser.Read(p)
	io.WriteString(l.rw, "R: "+fmtLine(string(p[:n])))
	io.WriteString(l.r, "R: "+fmtLine(string(p[:n])))
	return n, err
}

func Run(f func()) {
	if *resume {
		p, err := serial.Open(*port, &serial.Mode{BaudRate: *rate})
		if err != nil {
			failf("failed to open serial port: %v", err)
		}

		fdrw, _ := os.Create("RWDATA.log")
		fdr, _ := os.Create("RDATA.log")
		fdw, _ := os.Create("WDATA.log")
		defer fdrw.Close()
		defer fdr.Close()
		defer fdw.Close()

		c := grbl.NewGrbl(&logger{ReadWriteCloser: p, rw: fdrw, r: fdr, w: fdw})
		u, err := ui.NewJobUI(c, lines)
		if err != nil {
			failf("failed to launch UI: %v", err)
		}

		err = u.Start()
		if err != nil {
			failf("UI crashed: %v", err)
		}
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
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		fd, _ := os.Create("stack.log")
		pprof.Lookup("goroutine").WriteTo(fd, 1)
		fd.Close()
		termbox.Close()
		panic("dumped stack traces to stack.log")
	}()

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

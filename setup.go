package gg

import "flag"

func init() {
	flag.Bool("ui", false, "Launch in UI mode (for running jobs).")
	// flag.Bool("simulate", false, "Run in simulation mode, generating a summary of the job result.")
}

// Setup will parse parameters, ask for input (where required) and make things
// ready to start processing G-Code commands.
func Setup(c Config) {
	flag.Parse()

}

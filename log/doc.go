/*
Package log provides utilities to read and write data in the gg log format.

Logs are intended to be written as work is performed, capturing relevant information for resuming jobs.

Log File Format

The actual file format is line-deliniated ASCII text. Quoted strings are supported for parameter values, and serial data.


	@safety-height="1cm" ; default = "1cm"
	G21 (use mm)
	G90
	G1X1Y2 (first offset) F300

	; Start job
	_ZERO{1,2,-10}
	>"N1G21"
	<"ok"
	>"N2G90"
	<"ok"
	>"N3G1X1Y2F300"
	<"ok"


Parameters

Relevant parameters can be specified with a preceding `@` followed by a flag name, `=` and a quoted string as the value.

	@safety-height="1cm"

GCode

GCode can be specified directly in the log file. This also means a file containing nothing but gcode is a valid log, and enough
to run a job.

	G21
	G90
	G0X1 Y2

Coordinates

Coordinates may be embedded as well. This is useful for storing things like work-zero (relative to homed machine coords) for resuming
after power loss, or a soft reset.

They may be specified by a preceding `_` followed by an identifier, and the coords withing curly-braces.

	_ZERO{1,0,2}

Comments

Comments may be specified the same way as in LinuxCNC and Grbl. That is anything following `;` to the line end, or anything between parentheses.

	@board-thickness="19mm" ; this is a comment

	G0X1 (so is this) Y1

Serial Data

Serial data is preceded by the direction, `>` or `<`, (send or receive, respectively) followed by a quoted string of the data sent.
Only data relevant to the job should be logged (e.g. jog and status commands are not necessary).

	>"N1G21"
	<"ok"

*/
package log

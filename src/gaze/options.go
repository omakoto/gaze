package gaze

import (
	"bytes"
	"github.com/omakoto/zenlog-go/zenlog/shell"
	"os"
	"time"
)

var minInterval = time.Millisecond * 500

/*
   -d, --differences [permanent]
          Highlight the differences between successive updates.  Option will read optional argument that
          changes highlight to be permanent, allowing to see what has changed at least once since  first
          iteration.

   -n, --interval seconds
          Specify  update  interval.   The  command  will not allow quicker than 0.1 second interval, in
          which the smaller values are converted. Both '.' and ',' work for any locales.

   -p, --precise
          Make watch attempt to run command every interval seconds. Try it with ntptime and  notice  how
          the  fractional seconds stays (nearly) the same, as opposed to normal mode where they continu‐
          ously increase.

   -t, --no-title
          Turn off the header showing the interval, command, and current time at the top of the display,
          as well as the following blank line.

   -b, --beep
          Beep if command has a non-zero exit.

   -e, --errexit
          Freeze updates on command error, and exit after a key press.

   -g, --chgexit
          Exit when the output of command changes.

   -c, --color
          Interpret ANSI color and style sequences.

   -x, --exec
          command  is  given  to  sh  -c  which  means that you may need to use extra quoting to get the
          desired effect.  This with the --exec option, which passes the command to exec(2) instead.

   -h, --help
          Display help text and exit.

   -v, --version
          Display version information and exit.
*/

type Options struct {
	Input, Output *os.File

	// CommandLine is the command to execute.
	CommandLine []string

	// Interval between updates.
	Interval time.Duration

	// Precise attempts to run command every interval seconds.
	Precise bool

	// EnableDifferences enables highlighting between updates. (NOT IMPLEMENTED YET)
	EnableDifferences bool

	// EnableBeep controls whether or not to beep when the command returns a non-zero status. (NOT IMPLEMENTED YET)
	EnableBeep bool

	// UseExec controls whether to use exec(2) instead of running the command with "sh -c" (NOT IMPLEMENTED YET)
	UseExec bool

	// NoTitle disables the header.
	NoTitle bool

	ForcedTerminalWidth  int
	ForcedTerminalHeight int
}

func (o *Options) SetInterval(i time.Duration) {
	if i < minInterval {
		o.Interval = minInterval
		return
	}
	o.Interval = i
}

func (o *Options) GetExecCommand() []string {
	if o.UseExec {
		return o.CommandLine
	}
	buf := bytes.Buffer{}
	for i, a := range o.CommandLine {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(a)
	}

	out := make([]string, 0, 3)
	out = append(out, "/bin/sh")
	out = append(out, "-c")
	out = append(out, buf.String())

	return out
}

func (o *Options) GetDisplayCommand() string {
	buf := bytes.Buffer{}
	for i, a := range o.CommandLine {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(shell.Escape(a))
	}
	return buf.String()
}

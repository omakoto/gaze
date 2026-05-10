package main

import (
	"fmt"
	"github.com/omakoto/gaze/src/gaze"
	"github.com/omakoto/go-common/src/common"
	"github.com/pborman/getopt/v2"
	"math"
	"os"
	"time"
)

var (
	help  = getopt.BoolLong("help", 'h', "Show this help.")
	debug = getopt.BoolLong("debug", 'd', "Enable debug output.")

	precise = getopt.BoolLong("precise", 'p', "Attempt to run command at precise intervals (measure from start, not end).")
	noTitle = getopt.BoolLong("no-title", 't', "Hide the header line.")
	exec    = getopt.BoolLong("exec", 'x', "Run command directly via exec(2) instead of \"sh -c\".")
	times   = getopt.IntLong("repeat", 'r', -1, "Run command N times, then exit.")

	width  = getopt.IntLong("width", 0, 0, "Override terminal width. (default: auto-detect)")
	height = getopt.IntLong("height", 0, 0, "Override terminal height. (default: auto-detect)")

	_ = getopt.BoolLong("color", 'c', "Ignored. ANSI colors are always preserved.")

	interval float64 = 2
)

func init() {
	getopt.FlagLong(&interval, "interval", 'n', "Repeat interval in seconds. (default: 2.0, minimum: 0.1)")
	getopt.SetParameters("COMMAND [ARGS...]")
	getopt.SetUsage(printHelp)
}

func printHelp() {
	fmt.Fprintf(os.Stderr, `Gaze - a watch(1) replacement with full ANSI color support.

Repeatedly runs COMMAND and displays its output, refreshing the screen each
cycle. Unlike GNU watch, 8-bit and 24-bit colors are always preserved.

  https://github.com/omakoto/gaze

Usage:
  gaze [OPTIONS] COMMAND [ARGS...]
  gaze [OPTIONS] 'shell pipeline | with | pipes'

Examples:
  gaze df -h
  gaze -n 0.5 cat /proc/loadavg
  gaze -r 10 date
  gaze 'ps aux | grep python'

Options:
`)
	getopt.CommandLine.PrintOptions(os.Stderr)
	fmt.Fprintf(os.Stderr, `
Keyboard shortcuts (while running):
  Enter    Force an immediate refresh
  Space    Toggle pause / resume auto-refresh
  +        Increase interval by 0.5s
  -        Decrease interval by 0.5s
  q        Quit
`)
}

func main() {
	common.RunAndExit(func() int {
		getopt.Parse()
		args := getopt.Args()
		if *help || len(args) == 0 {
			getopt.Usage()
			os.Exit(0)
		}
		interval = math.Max(0.1, interval)
		if *debug {
			common.DebugEnabled = true
		}

		options := gaze.Options{}
		options.Input = os.Stdin
		options.Output = os.Stdout
		options.ForcedTerminalWidth = *width
		options.ForcedTerminalHeight = *height
		options.CommandLine = args
		options.SetInterval(time.Duration(float64(time.Second) * interval))
		options.Precise = *precise
		options.NoTitle = *noTitle
		options.UseExec = *exec

		common.Dump("Options: ", options)
		common.Debugf("Display command: %s\n", options.GetDisplayCommand())
		common.Dump("Exec command: ", options.GetExecCommand())

		gazer := gaze.NewGazer(options)
		defer gazer.Finish()

		gazer.RunLoop(*times)

		return 0
	})
}

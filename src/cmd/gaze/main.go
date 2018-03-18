package main

import (
	"github.com/omakoto/gaze/src/common"
	"github.com/omakoto/gaze/src/gaze"
	"github.com/pborman/getopt/v2"
	"math"
	"os"
	"time"
)

var (
	help  = getopt.BoolLong("help", 'h', "Show this help.")
	debug = getopt.BoolLong("debug", 'd', "Enable debug output.")

	precise = getopt.BoolLong("precise", 'p', "Attempt run command in precise intervals.")
	noTitle = getopt.BoolLong("no-title", 't', "Turn off header.")
	exec    = getopt.BoolLong("exec", 'x', "Pass command to exec instead of \"sh -c\".")
	times   = getopt.IntLong("repeat", 'r', -1, "Repeat command N times and finish.")

	width  = getopt.IntLong("width", 0, 0, "Specify terminal width. (default: auto)")
	height = getopt.IntLong("height", 0, 0, "Specify terminal height. (default: auto)")

	_ = getopt.BoolLong("color", 'c', "Ignored. ANSI colors are always preserved.")

	interval float64 = 2
)

func init() {
	getopt.FlagLong(&interval, "interval", 'n', "Specify interval in seconds.")
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
		options.Term = os.Stdout
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

		//testRun(options)
		gazer := gaze.NewGazer(options)
		defer gazer.Finish()

		gazer.RunLoop(*times)

		return 0
	})
}

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
)

// MustGetBinName returns the filename of the currently running executable.
func MustGetBinName() string {
	me, err := os.Executable()
	Check(err, "Executable failed")
	return filepath.Base(me)
}

func maybePrintStackTrack() {
	if getBinEnv("PRINT_STACK") == "1" {
		debug.PrintStack()
	}
}

func Fatal(message string) {
	Warn(message)
	maybePrintStackTrack()
	ExitFailure()
}

func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args))
}

func Warn(message string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", MustGetBinName(), message)
}

func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args))
}

func Check(err error, message string) {
	if err == nil {
		return
	}
	Fatal(fmt.Sprintf("%s: %s\n", message, err))
}

func Checkf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	Check(err, fmt.Sprintf(format, args...))
}

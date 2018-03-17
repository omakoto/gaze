package common

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"os"
	"strings"
)

var DebugEnabled = false

func init() {
	if getBinEnv("DEBUG") == "1" {
		DebugEnabled = true
	}
}

func getBinEnv(suffix string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", strings.ToUpper(MustGetBinName()), suffix))
}

func Debug(message string) {
	if !DebugEnabled {
		return
	}
	fmt.Fprint(os.Stdout, message)
}

func Debugf(format string, args ...interface{}) {
	if !DebugEnabled {
		return
	}
	Debug(fmt.Sprintf(format, args...))
}

func Dump(prefix string, object interface{}) {
	if !DebugEnabled {
		return
	}
	Debugf("%s%s", prefix, spew.Sdump(object))
}

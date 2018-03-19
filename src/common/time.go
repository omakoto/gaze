package common

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

var (
	timeOverrideFile = getBinEnv("TIME_INJECTION_FILE")
)

// Clock is a mockable clock interface.
type Clock interface {
	Now() time.Time
}

type clock struct {
}

// Return the current time.
func (clock) Now() time.Time {
	if timeOverrideFile == "" {
		return time.Now()
	}
	bytes, err := ioutil.ReadFile(timeOverrideFile)
	Check(err, "ReadFile failed")
	i, err := strconv.ParseInt(strings.TrimRight(string(bytes), "\n"), 10, 64)
	Check(err, "ParseInt failed")

	return time.Unix(i, 0)
}

// Create a new (real) Clock.
func NewClock() Clock {
	return clock{}
}

// InjectedClock is a mock clock.
type InjectedClock struct {
	Time time.Time
}

func (c InjectedClock) Now() time.Time {
	return c.Time
}

// NewInjectedClock creates a new mock clock.
func NewInjectedClock(time time.Time) Clock {
	return InjectedClock{time}
}

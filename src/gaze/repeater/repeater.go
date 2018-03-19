package repeater

import (
	"github.com/omakoto/gaze/src/common"
	"time"
)

type Repeatable interface {
	Run() error
	ShowResumeHelp()
	ShowHelp()
	Interval() time.Duration
	SetInterval(duration time.Duration)
}

type ByteReader interface {
	ReadByteTimeout(timeout time.Duration) (byte, error)
}

type Repeater struct {
	target Repeatable
}

func NewRepeater(target Repeatable) *Repeater {
	return &Repeater{target}
}

// TODO Test
func (r *Repeater) Loop(precise bool, times int, reader ByteReader, clock common.Clock) error {
	if times == 0 {
		return nil
	}
	var pausing bool
	nextExpectedStartTime := clock.Now()
	var lastExpectedStartTime, lastEndTime time.Time

	var baseTime func() time.Time
	if precise {
		baseTime = func() time.Time { return lastExpectedStartTime }
	} else {
		baseTime = func() time.Time { return lastEndTime }
	}
	forceRefresh := func() {
		lastExpectedStartTime = clock.Now()
	}

refresh:
	for i := 0; ; {
		lastExpectedStartTime = nextExpectedStartTime
		err := r.target.Run()
		if err != nil {
			return err
		}
		lastEndTime = clock.Now()

		i++
		if times > 0 && i >= times {
			break
		}

		nextExpectedStartTime = baseTime().Add(r.target.Interval())

	delay:
		for pausing || nextExpectedStartTime.After(clock.Now()) {
			wait := time.Until(nextExpectedStartTime)
			if pausing {
				wait = time.Hour * 24 * 365 * 10 // 10 years.
				r.target.ShowResumeHelp()
			}
			key, err := reader.ReadByteTimeout(wait)
			if err != nil {
				break
			}
			if key == 'q' {
				return nil
			}
			if key == '\n' {
				forceRefresh()
				continue refresh
			}
			if key == '-' {
				r.target.SetInterval(r.target.Interval() - time.Millisecond*500)
				forceRefresh()
				continue refresh
			}
			if key == '+' {
				r.target.SetInterval(r.target.Interval() + time.Millisecond*500)
				forceRefresh()
				continue refresh
			}
			if key == ' ' {
				if pausing {
					pausing = false
					forceRefresh()
					continue refresh
				}
				pausing = true
				continue delay
			}
			r.target.ShowHelp()
		}
	}
	return nil
}

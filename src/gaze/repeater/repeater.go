package repeater

import (
	"github.com/omakoto/gaze/src/common"
	"time"
)

type Repeatable interface {
	Run() error
	ShowResumeHelp()
	ShowHelp()
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

func (r *Repeater) Loop(precise bool, interval time.Duration, minInterval time.Duration, times int, reader ByteReader, clock common.Clock) error {
	if times == 0 {
		return nil
	}
	if interval < minInterval {
		interval = minInterval
	}

	var pausing bool
	lastExpectedStartTime := clock.Now()
	var lastEndTime time.Time

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
		err := r.target.Run()
		if err != nil {
			return err
		}
		lastEndTime = clock.Now()

		i++
		if times > 0 && i >= times {
			break
		}

		lastExpectedStartTime := baseTime().Add(interval)

	delay:
		for pausing || lastExpectedStartTime.After(clock.Now()) {
			wait := time.Until(lastExpectedStartTime)
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
				interval = interval - time.Millisecond*500
				if interval < minInterval {
					interval = minInterval
				}
				forceRefresh()
				continue refresh
			}
			if key == '+' {
				interval = interval + time.Millisecond*500
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

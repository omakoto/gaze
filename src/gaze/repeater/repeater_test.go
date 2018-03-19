package repeater

import (
	"github.com/omakoto/gaze/src/termio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

type MockRepeatable struct {
	mock.Mock
}

func (o *MockRepeatable) Run() error {
	args := o.Called()
	return args.Error(0)
}

func (o *MockRepeatable) ShowResumeHelp() {
	o.Called()
}

func (o *MockRepeatable) ShowHelp() {
	o.Called()
}

func (o *MockRepeatable) Interval() time.Duration {
	args := o.Called()
	return args.Get(0).(time.Duration)
}

func (o *MockRepeatable) SetInterval(duration time.Duration) {
	o.Called(duration)
}

type Event struct {
	delay time.Duration
	key   byte
	err   error
}

type EventsGenerator struct {
	now           time.Time
	lastEventTime time.Time
	events        []Event
	nextIndex     int
}

func NewEvents(events []Event) *EventsGenerator {
	ret := EventsGenerator{
		now:           startTime,
		lastEventTime: startTime,
		events:        events,
		nextIndex:     0,
	}
	return &ret
}

func (e *EventsGenerator) Now() time.Time {
	return e.now
}

func (e *EventsGenerator) nextEventTime() time.Time {
	if e.nextIndex < len(e.events) {
		return e.lastEventTime.Add(e.events[e.nextIndex].delay)
	}
	return e.now
}

func (e *EventsGenerator) ReadByteTimeout(timeout time.Duration) (byte, error) {
	if e.nextIndex < len(e.events) {
		nextEventTime := e.nextEventTime()
		if !nextEventTime.After(e.now.Add(timeout)) {
			nextEvent := e.events[e.nextIndex]
			e.nextIndex++
			e.now = e.now.Add(nextEvent.delay)
			return nextEvent.key, nextEvent.err
		}
		e.now = e.now.Add(timeout)
		return 0, termio.ErrReadTimedOut
	}
	return 0, termio.ErrReadClosing
}

func TestRepeater_Loop_none(t *testing.T) {
	r := MockRepeatable{}
	rp := NewRepeater(&r)

	events := []Event{
		{time.Hour, 'q', nil},
	}

	eg := NewEvents(events)

	rp.Loop(true, 0, eg, eg)

	assert.Equal(t, startTime, eg.now)
}

func TestRepeater_Loop_once(t *testing.T) {
	r := MockRepeatable{}
	rp := NewRepeater(&r)

	events := []Event{
		{time.Hour, 'q', nil},
	}

	r.On("Run").Return(nil)

	eg := NewEvents(events)

	rp.Loop(true, 1, eg, eg)

	assert.Equal(t, startTime, eg.now)
}

// TODO Fix this
func TestRepeater_Loop_twice(t *testing.T) {
	r := MockRepeatable{}
	rp := NewRepeater(&r)

	events := []Event{
		{time.Hour, 'q', nil},
	}

	r.On("Run").Return(nil)
	r.On("Interval").Return(time.Second)

	eg := NewEvents(events)

	rp.Loop(true, 2, eg, eg)

	// assert.Equal(t, startTime.Add(time.Second), eg.now)
}

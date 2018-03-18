package gaze

import (
	"bufio"
	"bytes"
	"github.com/omakoto/gaze/src/common"
	"github.com/omakoto/gaze/src/termio"
	"io"
	"time"
)

type Gazer struct {
	// options encapsulates gaze/watch options.
	options Options

	// execCommand is the command to pass to exec(2).
	execCommand []string

	// title is a human readable version of the command line.
	title string

	// lastStartTime is the start time of the last run.
	lastStartTime time.Time

	// nextExpectedTime is the time when the next run should start if the precise option is on.
	nextExpectedTime time.Time

	// headerBuffer is a cached buffer to build the header line.
	headerBuffer *bytes.Buffer

	// csiBuffer is usded to temporarily keep CSI sequences.
	csiBuffer *bytes.Buffer

	// fetcher abstracts how Gazer should execute a command and read its output.
	fetcher func() (io.ReadCloser, error)

	clock common.Clock
}

func NewGazer(options Options) *Gazer {
	execCommand := options.GetExecCommand()

	fetcher := func() (io.ReadCloser, error) {
		return StartCommand(execCommand)
	}
	clock := common.NewClock()
	err := termio.Init(options.Term, options.ForcedTerminalWidth, options.ForcedTerminalHeight)
	if err != nil {
		common.Fatalf("Unable to initialize terminal.")
	}
	return &Gazer{
		options: options,
		clock:   clock,

		title:       options.GetDisplayCommand(),
		execCommand: execCommand,

		headerBuffer: &bytes.Buffer{},
		csiBuffer:    &bytes.Buffer{},

		nextExpectedTime: clock.Now(),

		fetcher: fetcher,
	}
}

func (g *Gazer) Finish() {
	termio.Finish()
}

func (g *Gazer) Reinit() error {
	g.nextExpectedTime = g.nextExpectedTime.Add(g.options.Interval)
	g.lastStartTime = g.clock.Now()

	termio.Clear()

	return nil
}

func getHeader(hbuf *bytes.Buffer, width int, interval time.Duration, now time.Time, title string) string {
	timestamp := now.Format("2006/01/02 15:04:05.000 -0700")
	timestampWidth := len(timestamp)

	hbuf.Truncate(0)
	hbuf.WriteString("Every ")
	hbuf.WriteString(interval.String())
	hbuf.WriteString(": ")
	hbuf.WriteString(title)

	padLen := width - timestampWidth - termio.StringWidth(hbuf.String())
	if padLen > 0 {
		for i := 0; i < padLen; i++ {
			hbuf.WriteByte(' ')
		}
		hbuf.WriteString(timestamp)
	} else {
		// Not enough space; let's just print the timestamp.
		hbuf.Truncate(0)
		if timestampWidth > width {
			hbuf.WriteString(timestamp[0:width])
		} else {
			padLen := width - timestampWidth
			for i := 0; i < padLen; i++ {
				hbuf.WriteByte(' ')
			}
			hbuf.WriteString(timestamp)
		}
	}
	return hbuf.String()
}

func (g *Gazer) showHeader() {
	header := getHeader(g.headerBuffer, termio.Width(), g.options.Interval, g.lastStartTime, g.title)

	termio.WriteString(header)
	termio.NewLine()
}

func (g *Gazer) RunLoop(times int) error {
	if times == 0 {
		return nil
	}
	var pausing bool
	var forceRefresh bool
refresh:
	for i := 0; ; {
		if forceRefresh {
			g.nextExpectedTime = g.clock.Now()
			forceRefresh = false
		}
		err := g.RunOnce()
		if err != nil {
			return err
		}
		i++
		if times > 0 && i >= times {
			break
		}
		lastEndTime := g.clock.Now()

		var nextTime time.Time
		if g.options.Precise {
			nextTime = g.nextExpectedTime
		} else {
			nextTime = lastEndTime.Add(g.options.Interval)
		}

		// TODO Extract the control logic, clean it up and write tests.

	delay:
		for pausing || nextTime.After(g.clock.Now()) {
			wait := time.Until(nextTime)
			if pausing {
				wait = time.Hour * 24 * 365 * 10 // 10 years.
				g.showResumeHelp()
			}
			key, err := termio.ReadByte(wait)
			if err != nil {
				break
			}
			if key == 'q' {
				return nil
			}
			if key == '\n' {
				forceRefresh = true
				continue refresh
			}
			if key == '-' {
				g.options.SetInterval(g.options.Interval - time.Millisecond*500)
				forceRefresh = true
				continue refresh
			}
			if key == '+' {
				g.options.SetInterval(g.options.Interval + time.Millisecond*500)
				forceRefresh = true
				continue refresh
			}
			if key == ' ' {
				if pausing {
					pausing = false
					forceRefresh = true
					continue refresh
				}
				pausing = true
				continue delay
			}
			g.showHelp()
		}
	}
	return nil
}

func (g *Gazer) showHelp() {
	termio.MoveTo(0, termio.Height()-1)
	termio.WriteZeroWidthString("\x1b[K")
	termio.WriteString("[Enter] Refresh [-] Decrease interval [+] Increase interval [Space] Pause [q] Quit")
	termio.Flush()
}

func (g *Gazer) showResumeHelp() {
	termio.MoveTo(0, termio.Height()-1)
	termio.WriteZeroWidthString("\x1b[K")
	termio.WriteString("<Pausing> [Enter] Refresh [Space] Resume auto refresh [q] Quit")
	termio.Flush()
}

func (g *Gazer) RunOnce() error {
	g.Reinit()

	rd, err := g.fetcher()
	common.Check(err, "StartCommand() failed")
	defer rd.Close()

	if !g.options.NoTitle && termio.Height() >= 2 {
		g.showHeader()
	}

	g.Render(rd)

	return termio.Flush()
}

func (g *Gazer) Render(baseReader io.ReadCloser) {
	rd := bufio.NewReader(baseReader)

	for {
		if !termio.CanWrite() {
			break
		}
		ch, _, err := rd.ReadRune()
		if err != nil {
			return
		}
		if ch == '\n' || ch == '\r' { // Just tread CR as NL.
			if termio.NewLine() {
				continue
			}
			break
		}
		if ch != '\x1b' { // Not ESC?
			if ch < 0x20 { // Control char?
				if ch == '\t' {
					termio.Tab()
					continue
				}
				if termio.WriteRune('^') && termio.WriteRune(rune('@'+ch)) {
					continue
				}
				break
			}
			if ch == '\x7f' {
				if termio.WriteString("\\x7f") {
					continue
				}
				break
			}
			if termio.WriteRune(ch) {
				continue
			}
			break
		}

		// ESC
		peek, err := peekByte(rd)
		if peek == '[' {
			rd.ReadRune() // Throw away [

			b := g.csiBuffer
			b.Truncate(0)

			// CSI.
			// The ESC [ is followed by any number (including none) of
			// "parameter bytes" in the range 0x30–0x3F (ASCII 0–9:;<=>?),
			// then by any number of "intermediate bytes" in the range 0x20–0x2F (ASCII space and !"#$%&'()*+,-./),
			// then finally by a single "final byte" in the range 0x40–0x7E (ASCII @A–Z[\]^_`a–z{|}~).[

			// "parameter bytes" in the range 0x30–0x3F (ASCII 0–9:;<=>?),
			for {
				peek, err := peekByte(rd)
				if err != nil {
					break
				}
				if 0x30 <= peek && peek <= 0x3f {
					ch, _, _ := rd.ReadRune()
					b.WriteRune(ch)
					continue
				}
				break
			}

			// "intermediate bytes" in the range 0x20–0x2F (ASCII space and !"#$%&'()*+,-./),
			for {
				peek, err := peekByte(rd)
				if err != nil {
					break
				}
				if 0x20 <= peek && peek <= 0x2f {
					ch, _, _ := rd.ReadRune()
					b.WriteRune(ch)
					continue
				}
				break
			}

			// a single "final byte" in the range 0x40–0x7E (ASCII @A–Z[\]^_`a–z{|}~).[
			ch, _, _ := rd.ReadRune()
			b.WriteRune(ch)
			if err == nil && ch == 'm' {
				termio.WriteZeroWidthString("\x1b[")
				termio.WriteZeroWidthBytes(b.Bytes())
			} else {
				termio.WriteZeroWidthString("\\x1b[")
				termio.WriteZeroWidthBytes(b.Bytes())
			}
			continue
		}

		// Just print an escape sequence as is.
		if termio.WriteString("^[") {
			continue
		}
		break
	}
}

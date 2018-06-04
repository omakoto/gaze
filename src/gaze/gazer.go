package gaze

import (
	"bufio"
	"bytes"
	"github.com/omakoto/gaze/src/gaze/repeater"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/termio"
	"github.com/omakoto/go-common/src/utils"
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

	lastStartTime time.Time

	term termio.Term

	// headerBuffer is a cached buffer to build the header line.
	headerBuffer *bytes.Buffer

	// csiBuffer is usded to temporarily keep CSI sequences.
	csiBuffer *bytes.Buffer

	// fetcher abstracts how Gazer should execute a command and read its output.
	fetcher func() (io.ReadCloser, error)

	clock utils.Clock
}

type gazerAsRepeatable struct {
	g *Gazer
}

var _ repeater.Repeatable = (*gazerAsRepeatable)(nil)

func (gr *gazerAsRepeatable) Run() error {
	return gr.g.RunOnce()
}

func (gr *gazerAsRepeatable) ShowResumeHelp() {
	gr.g.showResumeHelp()
}

func (gr *gazerAsRepeatable) ShowHelp() {
	gr.g.showHelp()
}

func (gr *gazerAsRepeatable) Interval() time.Duration {
	return gr.g.options.Interval
}

func (gr *gazerAsRepeatable) SetInterval(interval time.Duration) {
	gr.g.options.SetInterval(interval)
}

func NewGazer(options Options) *Gazer {
	execCommand := options.GetExecCommand()

	fetcher := func() (io.ReadCloser, error) {
		return StartCommand(execCommand)
	}
	clock := utils.NewClock()
	term, err := termio.NewTerm(options.Input, options.Output, options.ForcedTerminalWidth, options.ForcedTerminalHeight)
	if err != nil {
		common.Fatalf("Unable to initialize terminal.")
	}
	return &Gazer{
		options: options,
		clock:   clock,

		title:       options.GetDisplayCommand(),
		execCommand: execCommand,

		term: term,

		headerBuffer: &bytes.Buffer{},
		csiBuffer:    &bytes.Buffer{},

		fetcher: fetcher,
	}
}

// Finish does all clean-ups. Must call it once done.
func (g *Gazer) Finish() {
	g.term.Finish()
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
	header := getHeader(g.headerBuffer, g.term.Width(), g.options.Interval, g.lastStartTime, g.title)

	g.term.WriteString(header)
	g.term.NewLine()
}

func (g *Gazer) RunLoop(times int) error {
	r := repeater.NewRepeater(&gazerAsRepeatable{g})
	return r.Loop(g.options.Precise, times, g.term, g.clock)
}

func (g *Gazer) showHelp() {
	g.term.MoveTo(0, g.term.Height()-1)
	g.term.WriteZeroWidthString("\x1b[K")
	g.term.WriteString("[Enter] Refresh [-] Decrease interval [+] Increase interval [Space] Pause [q] Quit")
	g.term.Flush()
}

func (g *Gazer) showResumeHelp() {
	g.term.MoveTo(0, g.term.Height()-1)
	g.term.WriteZeroWidthString("\x1b[K")
	g.term.WriteString("<Pausing> [Enter] Refresh [Space] Resume auto refresh [q] Quit")
	g.term.Flush()
}

func (g *Gazer) RunOnce() error {
	g.lastStartTime = g.clock.Now()
	g.term.Clear()

	rd, err := g.fetcher()
	common.Check(err, "StartCommand() failed")
	defer rd.Close()

	if !g.options.NoTitle && g.term.Height() >= 2 {
		g.showHeader()
	}

	g.render(rd)

	return g.term.Flush()
}

func (g *Gazer) render(baseReader io.ReadCloser) {
	rd := bufio.NewReader(baseReader)

	t := g.term
	for {
		if !t.CanWrite() {
			break
		}
		ch, _, err := rd.ReadRune()
		if err != nil {
			return
		}
		if ch == '\n' || ch == '\r' { // Just tread CR as NL.
			if t.NewLine() {
				continue
			}
			break
		}
		if ch != '\x1b' { // Not ESC?
			if ch < 0x20 { // Control char?
				if ch == '\t' {
					t.Tab()
					continue
				}
				if t.WriteRune('^') && t.WriteRune(rune('@'+ch)) {
					continue
				}
				break
			}
			if ch == '\x7f' {
				if t.WriteString("\\x7f") {
					continue
				}
				break
			}
			if t.WriteRune(ch) {
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
				t.WriteZeroWidthString("\x1b[")
				t.WriteZeroWidthBytes(b.Bytes())
			} else {
				t.WriteZeroWidthString("\\x1b[")
				t.WriteZeroWidthBytes(b.Bytes())
			}
			continue
		}

		// Just print an escape sequence as is.
		if t.WriteString("^[") {
			continue
		}
		break
	}
}

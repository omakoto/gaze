package gaze

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/omakoto/gaze/src/common"
	"github.com/omakoto/gaze/src/termio"
	"golang.org/x/crypto/ssh/terminal"
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

	buffer *termio.Buffer

	// headerBuffer is a cached buffer to build the header line.
	headerBuffer *bytes.Buffer

	// csiBuffer is usded to temporarily keep CSI sequences.
	csiBuffer *bytes.Buffer

	// fetcher abstracts how Gazer should execute a command and read its output.
	fetcher func() (io.ReadCloser, error)
}

func NewGazer(options Options) *Gazer {
	execCommand := options.GetExecCommand()

	fetcher := func() (io.ReadCloser, error) {
		return StartCommand(execCommand)
	}
	return &Gazer{
		options: options,

		title:       options.GetDisplayCommand(),
		execCommand: execCommand,

		buffer:       termio.NewBuffer(),
		headerBuffer: &bytes.Buffer{},
		csiBuffer:    &bytes.Buffer{},

		nextExpectedTime: time.Now(),

		fetcher: fetcher,
	}
}

func (g *Gazer) Reinit() error {
	g.nextExpectedTime = g.nextExpectedTime.Add(g.options.Interval)
	g.lastStartTime = time.Now()

	w, h, err := terminal.GetSize(1)
	if err != nil {
		return err
	}
	g.buffer.Reset(w, h)

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
	header := getHeader(g.headerBuffer, g.buffer.Width(), g.options.Interval, g.lastStartTime, g.title)

	g.buffer.WriteString(header)
	g.buffer.MoveTo(0, 1)
}

func (g *Gazer) RunLoop(times int) error {
	if times == 0 {
		return nil
	}
	for i := 0; ; {
		err := g.RunOnce()
		if err != nil {
			return err
		}
		i++
		if times > 0 && i >= times {
			break
		}
		var realInterval time.Duration
		if g.options.Precise {
			realInterval = time.Until(g.nextExpectedTime)
		} else {
			realInterval = g.options.Interval
		}
		if realInterval > 0 {
			time.Sleep(realInterval)
		}
	}
	return nil
}

func (g *Gazer) flush() error {
	return g.buffer.Flush(g.options.Writer)
}

func (g *Gazer) Finish() {
	fmt.Fprint(g.options.Writer, "\x1b[?25h\n") // Show cursor
}

func (g *Gazer) RunOnce() error {
	g.Reinit()

	rd, err := g.fetcher()
	common.Check(err, "StartCommand() failed")
	defer rd.Close()

	if !g.options.NoTitle {
		g.showHeader()
	}

	g.Render(rd)

	return g.flush()
}

func (g *Gazer) Render(baseReader io.ReadCloser) {
	rd := bufio.NewReader(baseReader)

	out := g.buffer

	for {
		if !out.CanWrite() {
			break
		}
		ch, _, err := rd.ReadRune()
		if err != nil {
			return
		}
		if ch == '\n' || ch == '\r' { // Just tread CR as NL.
			if out.NewLine() {
				continue
			}
			break
		}
		if ch != '\x1b' { // Not ESC?
			if ch < 0x20 { // Control char?
				if ch == '\t' {
					out.Tab()
					continue
				}
				if out.WriteRune('^') && out.WriteRune(rune('@'+ch)) {
					continue
				}
				break
			}
			if ch == '\x7f' {
				if out.WriteString("\\x7f") {
					continue
				}
				break
			}
			if out.WriteRune(ch) {
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
				out.WriteZeroWidthString("\x1b[")
				out.WriteZeroWidthBytes(b.Bytes())
			} else {
				out.WriteZeroWidthString("\\x1b[")
				out.WriteZeroWidthBytes(b.Bytes())
			}
			continue
		}

		// Just print an escape sequence as is.
		if out.WriteString("^[") {
			continue
		}
		break
	}
}

package termio

import (
	"bytes"
	"fmt"
	"github.com/omakoto/gaze/src/common"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"syscall"
	"time"
)

type Term interface {
	Clear()
	Finish()
	Width() int
	Height() int
	WriteZeroWidthString(s string)
	WriteZeroWidthBytes(bytes []byte)
	MoveTo(newX, newY int)
	Tab()
	UpdateCursor()
	CanWrite() bool
	CanWriteChars(charWidth int) bool
	NewLine() bool
	WriteString(s string) bool
	WriteRune(ch rune) bool
	Flush() error
	ReadByteTimeout(timeout time.Duration) (byte, error)
}

type termImpl struct {
	in, out *os.File

	// width is the terminal width.
	width int

	// height is the terminal width.
	height int

	forceSize bool

	// x is the cursor x position.
	x int

	// x is the cursor y position.
	y int

	// buffer is where Gazer stores output. Gazer flushes its content to options.Writer at once.
	buffer *bytes.Buffer

	running bool

	// Used by reader
	readBuffer []byte
	readBytes  chan ByteAndError
	quitChan   chan bool

	origTermiosIn  syscall.Termios
	origTermiosOut syscall.Termios
}

var _ Term = (*termImpl)(nil)

func NewTerm(in, out *os.File, forcedWidth, forcedHeight int) (Term, error) {
	t := &termImpl{}

	t.running = true
	t.buffer = &bytes.Buffer{}

	t.in = in
	t.out = out
	if forcedWidth > 0 && forcedHeight > 0 {
		t.forceSize = true
		t.width = forcedWidth
		t.height = forcedHeight
	}

	err := initTerm(t)
	if err != nil {
		return nil, err
	}

	t.Clear()

	return t, nil
}

func (t *termImpl) Clear() {
	if !t.forceSize {
		w, h, err := terminal.GetSize(1)
		common.Check(err, "Unable to get terminal size.")
		t.width = w
		t.height = h
	}

	t.buffer.Truncate(0)

	t.WriteZeroWidthString("\x1b[2J\x1b[?25l") // Erase entire screen, hide cursor.
	t.MoveTo(0, 0)
}

func (t *termImpl) Finish() {
	if !t.running {
		return
	}
	// TODO Make sure it'll clean up partially initialized state too.
	fmt.Fprint(t.out, "\x1b[?25h\n") // Show cursor
	deinitTerm(t)

	// TODO Don't close them so the process can restart termio. But closing in will finish the reader goroutine.
	t.in.Close()
	t.out.Close()
}

func (t *termImpl) Width() int {
	return t.width
}

func (t *termImpl) Height() int {
	return t.height
}

func (t *termImpl) WriteZeroWidthString(s string) {
	t.buffer.WriteString(s)
}

func (t *termImpl) WriteZeroWidthBytes(bytes []byte) {
	t.buffer.Write(bytes)
}

func (t *termImpl) MoveTo(newX, newY int) {
	t.x = newX
	t.y = newY
	t.UpdateCursor()
}

func (t *termImpl) Tab() {
	t.x += 8 - (t.x % 8)
	if t.x >= t.width {
		t.NewLine()
		return
	}
	t.UpdateCursor()
}

func (t *termImpl) UpdateCursor() {
	t.WriteZeroWidthString(fmt.Sprintf("\x1b[%d;%dH", t.y+1, t.x+1))
}

func (t *termImpl) CanWrite() bool {
	return t.CanWriteChars(1)
}

func (t *termImpl) CanWriteChars(charWidth int) bool {
	if t.y < t.height-1 {
		return true
	}
	return t.x+charWidth <= t.width
}

func (t *termImpl) NewLine() bool {
	t.y++
	t.x = 0
	if t.y < t.height {
		// We don't simply use \n here, because if the last character is a wide char,
		// then we're not confident where the last character will be put.
		t.buffer.WriteByte('\n')
		t.UpdateCursor()
		return true
	}
	return false
}

func (t *termImpl) WriteString(s string) bool {
	for _, ch := range s {
		if t.WriteRune(ch) {
			continue
		}
		return false
	}
	return true
}

func (t *termImpl) WriteRune(ch rune) bool {
	runeWidth := RuneWidth(ch)
	if t.x+runeWidth > t.width {
		if !t.NewLine() {
			return false
		}
	}
	if t.CanWriteChars(runeWidth) {
		t.buffer.WriteRune(ch)
		t.x += runeWidth
		return true
	}
	return false
}

func (t *termImpl) Flush() error {
	_, err := t.out.Write(t.buffer.Bytes())
	return err
}

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
	// Width returns the terminal width.
	Width() int

	// Height returns the terminal height.
	Height() int

	// Clear clears the internal buffer, but it won't flush. Clear also refreshes the terminal size.
	Clear()

	// Flush flushes the internal buffer to the output terminal.
	Flush() error

	// Finish cleans up the terminal and restores the original state.
	Finish()

	// WriteZeroWidthString writes a string to the internal buffer without moving the cursor.
	WriteZeroWidthString(s string)

	// WriteZeroWidthBytes writes a byte array to the internal buffer without moving the cursor.
	WriteZeroWidthBytes(bytes []byte)

	// MoveTo moves the cursor to a given position.
	MoveTo(newX, newY int)

	// Tab moves the cursor to the next tab position.
	Tab()

	// Can write returns whether to be able to write anything at the cursor position without overflowing the terminal.
	CanWrite() bool

	// Can write returns whether to be able to write a char of a given width at the cursor position without overflowing the terminal.
	CanWriteChars(charWidth int) bool

	// NewLine moves the cursor to the beginning of the next line.
	NewLine() bool

	// WriteString writes a string to the internal buffer and moves the cursor.
	WriteString(s string) bool

	// WriteString writes a rune to the internal buffer and moves the cursor.
	WriteRune(ch rune) bool

	// ReadByteTimeout reads a byte from terminal with a timeout.
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
	t.updateCursor()
}

func (t *termImpl) Tab() {
	t.x += 8 - (t.x % 8)
	if t.x >= t.width {
		t.NewLine()
		return
	}
	t.updateCursor()
}

func (t *termImpl) updateCursor() {
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
		t.updateCursor()
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

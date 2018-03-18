package termio

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type Term struct {
	reader io.Reader
	writer io.Writer

	// width is the terminal width.
	width int

	// height is the terminal width.
	height int

	// x is the cursor x position.
	x int

	// x is the cursor y position.
	y int

	// buffer is where Gazer stores output. Gazer flushes its content to options.Writer at once.
	buffer *bytes.Buffer
}

func NewBuffer(reader io.Reader, writer io.Writer, width, height int) *Term {
	if width < 1 || height < 1 {
		log.Panicf("Invalid terminal size: %d x %d", width, height)
	}
	t := &Term{
		reader: reader,
		writer: writer,
		width:  width,
		height: height,
		buffer: &bytes.Buffer{},
	}
	t.Clear(t.width, t.height)
	return t
}

func (t *Term) Clear(width, height int) {
	t.width = width
	t.height = height

	t.buffer.Truncate(0)

	t.WriteZeroWidthString("\x1b[2J\x1b[?25l") // Erase entire screen, hide cursor.
	t.MoveTo(0, 0)
}

func (t *Term) Finish() {
	fmt.Fprint(t.writer, "\x1b[?25h\n") // Show cursor
}

func (t *Term) Width() int {
	return t.width
}

func (t *Term) Height() int {
	return t.height
}

func (t *Term) WriteZeroWidthString(s string) {
	t.buffer.WriteString(s)
}

func (t *Term) WriteZeroWidthBytes(bytes []byte) {
	t.buffer.Write(bytes)
}

func (t *Term) MoveTo(x, y int) {
	t.x = x
	t.y = y
	t.UpdateCursor()
}

func (t *Term) Tab() {
	t.x += 8 - (t.x % 8)
	if t.x >= t.width {
		t.NewLine()
		return
	}
	t.UpdateCursor()
}

func (t *Term) UpdateCursor() {
	t.WriteZeroWidthString(fmt.Sprintf("\x1b[%d;%dH", t.y+1, t.x+1))
}

func (t *Term) CanWrite() bool {
	return t.CanWriteChars(1)
}

func (t *Term) CanWriteChars(charWidth int) bool {
	if t.y < t.height-1 {
		return true
	}
	return t.x+charWidth <= t.width
}

func (t *Term) NewLine() bool {
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

func (t *Term) WriteString(s string) bool {
	for _, ch := range s {
		if t.WriteRune(ch) {
			continue
		}
		return false
	}
	return true
}

func (t *Term) WriteRune(ch rune) bool {
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

func (t *Term) Flush() error {
	_, err := t.writer.Write(t.buffer.Bytes())
	return err
}

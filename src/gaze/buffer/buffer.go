package buffer

import (
	"bytes"
	"fmt"
	"io"
)

type Buffer struct {
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

func NewBuffer() *Buffer {
	return &Buffer{buffer: &bytes.Buffer{}}
}

func (b *Buffer) Width() int {
	return b.width
}

func (b *Buffer) Height() int {
	return b.height
}

func (b *Buffer) Reset(width, height int) {
	b.width = width
	b.height = height

	b.buffer.Truncate(0)

	b.WriteZeroWidthString("\x1b[2J\x1b[?25l") // Erase entire screen, hide cursor.
	b.MoveTo(0, 0)
}

func (b *Buffer) WriteZeroWidthString(s string) {
	b.buffer.WriteString(s)
}

func (b *Buffer) WriteZeroWidthBytes(bytes []byte) {
	b.buffer.Write(bytes)
}

func (b *Buffer) MoveTo(x, y int) {
	b.x = x
	b.y = y
	b.UpdateCursor()
}

func (b *Buffer) Tab() {
	b.x += 8 - (b.x % 8)
	if b.x >= b.width {
		b.NewLine()
		return
	}
	b.UpdateCursor()
}

func (b *Buffer) UpdateCursor() {
	b.WriteZeroWidthString(fmt.Sprintf("\x1b[%d;%dH", b.y+1, b.x+1))
}

func (b *Buffer) CanWrite() bool {
	return b.CanWriteChars(1)
}

func (b *Buffer) CanWriteChars(charWidth int) bool {
	if b.y < b.height-1 {
		return true
	}
	return b.x+charWidth <= b.width
}

func (b *Buffer) NewLine() bool {
	b.y++
	b.x = 0
	if b.y < b.height {
		// We don't simply use \n here, because if the last character is a wide char,
		// then we're not confident about the behavior where the last character will be put.
		b.UpdateCursor()
		return true
	}
	return false
}

func (b *Buffer) WriteString(s string) bool {
	for _, ch := range s {
		if b.WriteRune(ch) {
			continue
		}
		return false
	}
	return true
}

func (b *Buffer) WriteRune(ch rune) bool {
	runeWidth := RuneWidth(ch)
	if b.x+runeWidth > b.width {
		if !b.NewLine() {
			return false
		}
	}
	if b.CanWriteChars(runeWidth) {
		b.buffer.WriteRune(ch)
		b.x += runeWidth
		return true
	}
	return false
}

func (b *Buffer) Flush(wr io.Writer) error {
	_, err := wr.Write(b.buffer.Bytes())
	return err
}

package termio

import (
	"bytes"
	"fmt"
	"github.com/omakoto/gaze/src/common"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

var (
	term *os.File

	// width is the terminal width.
	width int

	// height is the terminal width.
	height int

	forceSize bool

	// x is the cursor x position.
	x = 0

	// x is the cursor y position.
	y = 0

	// buffer is where Gazer stores output. Gazer flushes its content to options.Writer at once.
	buffer = &bytes.Buffer{}
)

func Init(terminal *os.File, forcedWidth, forcedHeight int) error {
	term = terminal
	if forcedWidth > 0 && forcedHeight > 0 {
		forceSize = true
		width = forcedWidth
		height = forcedHeight
	}

	err := initTerm(term)
	if err != nil {
		return err
	}
	Clear()
	return nil
}

func Clear() {
	if !forceSize {
		w, h, err := terminal.GetSize(1)
		common.Check(err, "Unable to get terminal size.")
		width = w
		height = h
	}

	buffer.Truncate(0)

	WriteZeroWidthString("\x1b[2J\x1b[?25l") // Erase entire screen, hide cursor.
	MoveTo(0, 0)
}

func Finish() {
	fmt.Fprint(term, "\x1b[?25h\n") // Show cursor
	deinitTerm(term)
}

func Width() int {
	return width
}

func Height() int {
	return height
}

func WriteZeroWidthString(s string) {
	buffer.WriteString(s)
}

func WriteZeroWidthBytes(bytes []byte) {
	buffer.Write(bytes)
}

func MoveTo(newX, newY int) {
	x = newX
	y = newY
	UpdateCursor()
}

func Tab() {
	x += 8 - (x % 8)
	if x >= width {
		NewLine()
		return
	}
	UpdateCursor()
}

func UpdateCursor() {
	WriteZeroWidthString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
}

func CanWrite() bool {
	return CanWriteChars(1)
}

func CanWriteChars(charWidth int) bool {
	if y < height-1 {
		return true
	}
	return x+charWidth <= width
}

func NewLine() bool {
	y++
	x = 0
	if y < height {
		// We don't simply use \n here, because if the last character is a wide char,
		// then we're not confident where the last character will be put.
		buffer.WriteByte('\n')
		UpdateCursor()
		return true
	}
	return false
}

func WriteString(s string) bool {
	for _, ch := range s {
		if WriteRune(ch) {
			continue
		}
		return false
	}
	return true
}

func WriteRune(ch rune) bool {
	runeWidth := RuneWidth(ch)
	if x+runeWidth > width {
		if !NewLine() {
			return false
		}
	}
	if CanWriteChars(runeWidth) {
		buffer.WriteRune(ch)
		x += runeWidth
		return true
	}
	return false
}

func Flush() error {
	_, err := term.Write(buffer.Bytes())
	return err
}

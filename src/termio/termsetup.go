package termio

import (
	"github.com/mattn/go-isatty"
	"os"
	"syscall"
)

func initTerm(t *Term) error {
	t.quitChan = make(chan bool, 1)
	t.readBuffer = make([]byte, 1)
	t.readBytes = make(chan ByteAndError, 1)
	go reader(t)

	err := saveTermios(t.out, &t.origTermiosOut)
	if err != nil {
		return err
	}
	err = saveTermios(t.in, &t.origTermiosIn)
	if err != nil {
		return err
	}

	err = initTermios(t.out, &t.origTermiosOut)
	if err != nil {
		return err
	}
	err = initTermios(t.in, &t.origTermiosIn)
	if err != nil {
		return err
	}
	return nil
}

// From termbox-go
//tios.Iflag &^= syscall_IGNBRK | syscall_BRKINT | syscall_PARMRK |
//	syscall_ISTRIP | syscall_INLCR | syscall_IGNCR |
//	syscall_ICRNL | syscall_IXON
//tios.Lflag &^= syscall_ECHO | syscall_ECHONL | syscall_ICANON |
//	syscall_ISIG | syscall_IEXTEN
//tios.Cflag &^= syscall_CSIZE | syscall_PARENB
//tios.Cflag |= syscall_CS8
//tios.Cc[syscall_VMIN] = 1
//tios.Cc[syscall_VTIME] = 0

func saveTermios(file *os.File, to *syscall.Termios) error {
	if !isatty.IsTerminal(file.Fd()) {
		return nil
	}
	orig := syscall.Termios{}
	err := tcgetattr(file.Fd(), &orig)
	if err != nil {
		return err
	}
	*to = orig
	return nil
}

func initTermios(file *os.File, orig *syscall.Termios) error {
	if !isatty.IsTerminal(file.Fd()) {
		return nil
	}

	new := *orig
	new.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON

	return tcsetattr(file.Fd(), &new)
}

func restoreTermios(file *os.File, orig *syscall.Termios) error {
	if !isatty.IsTerminal(file.Fd()) {
		return nil
	}
	return tcsetattr(file.Fd(), orig)
}

func deinitTerm(t *Term) error {
	if t.quitChan != nil {
		t.quitChan <- true // Stop the reader
		close(t.quitChan)
		close(t.readBytes)
	}

	fd := t.out.Fd()
	if !isatty.IsTerminal(fd) {
		return nil
	}
	err := restoreTermios(t.out, &t.origTermiosOut)
	if err != nil {
		return err
	}
	return restoreTermios(t.in, &t.origTermiosIn)
}

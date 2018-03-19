package termio

import (
	"github.com/mattn/go-isatty"
	"syscall"
)

func initTerm(t *Term) error {
	fd := t.term.Fd()
	if !isatty.IsTerminal(fd) {
		// If output is not terminal, let's just still work.
		return nil
	}

	_, err := fcntl(fd, syscall.F_SETFL, syscall.O_ASYNC|syscall.O_NONBLOCK)
	if err != nil {
		return err
	}

	_, err = fcntl(fd, syscall.F_SETOWN, syscall.Getpid())
	if err != nil {
		return err
	}

	err = tcgetattr(fd, &t.origTermios)
	if err != nil {
		return err
	}

	tios := t.origTermios

	tios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON

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

	err = tcsetattr(fd, &tios)
	if err != nil {
		return err
	}

	t.quitChan = make(chan bool, 1)
	t.readBuffer = make([]byte, 1)
	t.readBytes = make(chan ByteAndError, 1)
	go reader(t)

	return nil
}

func deinitTerm(t *Term) error {
	if t.quitChan != nil {
		t.quitChan <- true // Stop the reader
		close(t.quitChan)
		close(t.readBytes)
	}

	fd := t.term.Fd()
	if !isatty.IsTerminal(fd) {
		return nil
	}
	tcsetattr(fd, &t.origTermios)
	return nil
}

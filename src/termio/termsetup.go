package termio

import (
	"github.com/mattn/go-isatty"
	"os"
	"os/signal"
	"syscall"
)

var (
	sigio = make(chan os.Signal, 1)
	quit  chan bool

	origTermios = syscall.Termios{}
)

func initTerm() error {
	fd := term.Fd()
	if !isatty.IsTerminal(fd) {
		// If output is not terminal, let's just still work.
		return nil
	}

	signal.Notify(sigio, syscall.SIGIO)

	_, err := fcntl(fd, syscall.F_SETFL, syscall.O_ASYNC|syscall.O_NONBLOCK)
	if err != nil {
		return err
	}

	_, err = fcntl(fd, syscall.F_SETOWN, syscall.Getpid())
	if err != nil {
		return err
	}

	err = tcgetattr(fd, &origTermios)
	if err != nil {
		return err
	}

	tios := origTermios

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

	quit = make(chan bool, 1)
	go reader(quit)

	return nil
}

func deinitTerm() error {
	if quit != nil {
		quit <- true // Stop the reader
	}
	quit = nil

	fd := term.Fd()
	if !isatty.IsTerminal(fd) {
		return nil
	}
	tcsetattr(fd, &origTermios)
	return nil
}

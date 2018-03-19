package termio

import (
	"errors"
	"time"
)

var (
	ErrReadTimedOut = errors.New("read timed out")
	ErrReadClosing  = errors.New("termio closing")
)

type ByteAndError struct {
	b   byte
	err error
}

func reader(t *Term) {
	// TODO Don't use sigio, use blocked read with select.
	for {
		select {
		//FIXME
		//case <-sigio:
		//	read, _ := t.term.Read(t.readBuffer)
		//	// common.Check(err, "TODO Handle it somehow")
		//	if read > 0 {
		//		t.readBytes <- ByteAndError{t.readBuffer[0], nil}
		//	}
		case <-t.quitChan:
			return
		}
	}
}

func (t *Term) ReadByte(timeout time.Duration) (byte, error) {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- true
	}()

	for {
		select {
		case b := <-t.readBytes:
			return b.b, b.err
		case <-timeoutChan:
			return 0, ErrReadTimedOut
		}
	}
}

package termio

import (
	"errors"
	"io"
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

func startReader(t *termImpl) {
	go func() {
		for {
			read, err := t.out.Read(t.readBuffer)
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			if read > 0 {
				t.readBytes <- ByteAndError{t.readBuffer[0], nil}
			}
		}
	}()
}

func (t *termImpl) ReadByteTimeout(timeout time.Duration) (byte, error) {
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
		case <-t.quitChan:
			return 0, ErrReadClosing
		}
	}
}

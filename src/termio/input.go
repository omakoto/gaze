package termio

import (
	"errors"
	"time"
)

var (
	ErrReadTimedOut = errors.New("read timed out")
	ErrReadClosing  = errors.New("termio closing")

	readBuffer = make([]byte, 1)
	readBytes  = make(chan struct {
		byte
		error
	}, 1)
)

func reader(quitChan chan bool) {
	for {
		select {
		case <-sigio:
			read, _ := term.Read(readBuffer)
			// common.Check(err, "TODO Handle it somehow")
			if read > 0 {
				readBytes <- struct {
					byte
					error
				}{readBuffer[0], nil}
			}
		case <-quitChan:
			return
		}
	}
}

func ReadByte(timeout time.Duration) (byte, error) {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutChan <- true
	}()

	for {
		select {
		case b := <-readBytes:
			return b.byte, b.error
		case <-timeoutChan:
			return 0, ErrReadTimedOut
		}
	}
}

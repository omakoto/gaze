package gaze

import (
	"github.com/omakoto/gaze/src/common"
	"io"
	"os"
	"os/exec"
)

type startCommandReader struct {
	rd     io.ReadCloser
	cmd    *exec.Cmd
	closed bool
}

func (r *startCommandReader) Read(p []byte) (n int, err error) {
	return r.rd.Read(p)
}

func (r *startCommandReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true

	err := r.rd.Close()
	if err != nil {
		return err
	}

	r.cmd.Process.Signal(os.Interrupt)

	return r.cmd.Wait()
}

var _ = io.ReadCloser((*startCommandReader)(nil))

func StartCommand(commandLine []string) (io.ReadCloser, error) {
	common.Debugf("Executing: %v...\n", commandLine)
	cmd := exec.Command(commandLine[0], commandLine[1:]...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	err = cmd.Start()
	if err != nil {
		out.Close()
		out = nil
		return nil, err
	}

	return &startCommandReader{rd: out, cmd: cmd}, nil
}

package gaze

import "bufio"

func peekByte(r *bufio.Reader) (byte, error) {
	p, err := r.Peek(1)
	if len(p) > 0 {
		return p[0], nil
	}
	return 0, err
}

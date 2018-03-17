package gaze

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGazer_getHeader(t *testing.T) {
	hbuf := &bytes.Buffer{}
	now, err := time.Parse("2006/01/02 15:04:05.000 -0700", "2006/01/02 15:04:05.000 -0700")
	if err != nil {
		panic(fmt.Sprintf("Unable to parse date: %s", err))
	}
	inputs := []struct {
		width    int
		interval float32
		title    string
		expected string
	}{
		{0, 1, "title", ""},
		{1, 1, "title", "2"},
		{2, 1, "title", "20"},
		{28, 1, "title", "2006/01/02 15:04:05.000 -070"},
		{29, 1, "title", "2006/01/02 15:04:05.000 -0700"},
		{30, 1, "title", " 2006/01/02 15:04:05.000 -0700"},
		{31, 1, "title", "  2006/01/02 15:04:05.000 -0700"},
		{60, 1, "title", "Every 1s: title                2006/01/02 15:04:05.000 -0700"},
		{50, 1, "title", "Every 1s: title      2006/01/02 15:04:05.000 -0700"},
		{50, 1.5, "title", "Every 1.5s: title    2006/01/02 15:04:05.000 -0700"},
		{50, .5, "title", "Every 500ms: title   2006/01/02 15:04:05.000 -0700"},
		{50, .5, "title12", "Every 500ms: title12 2006/01/02 15:04:05.000 -0700"},
		{50, .5, "title123", "                     2006/01/02 15:04:05.000 -0700"},
	}
	for _, v := range inputs {

		assert.Equal(t, v.expected, getHeader(hbuf, v.width, time.Duration(float32(time.Second)*v.interval), now, v.title), "W=%d T=%s", v.width, v.title)
	}
}

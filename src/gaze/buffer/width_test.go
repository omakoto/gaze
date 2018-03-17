package buffer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringWidth(t *testing.T) {
	inputs := []struct {
		text  string
		width int
	}{
		{"", 0},
		{"a", 1},
		{"abc", 3},
		{"あいうえお", 10},
		{"aあいうえお", 11},
	}
	for _, v := range inputs {
		assert.Equal(t, v.width, StringWidth(v.text), "Width of %s", v.text)
	}
}

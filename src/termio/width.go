package termio

import "github.com/mattn/go-runewidth"

// RuneWidth returns the width in the number of characters of a given rune.
func RuneWidth(ch rune) int {
	w := runewidth.RuneWidth(ch)
	if w == 0 || w == 2 && runewidth.IsAmbiguousWidth(ch) {
		return 1
	}
	return w
}

// RuneWidth returns the width in the number of characters of a given string.
func StringWidth(s string) int {
	ret := 0
	for _, ch := range s {
		ret += RuneWidth(ch)
	}
	return ret
}

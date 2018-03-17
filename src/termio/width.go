package termio

import "github.com/mattn/go-runewidth"

func RuneWidth(ch rune) int {
	w := runewidth.RuneWidth(ch)
	if w == 0 || w == 2 && runewidth.IsAmbiguousWidth(ch) {
		return 1
	}
	return w
}

func StringWidth(s string) int {
	ret := 0
	for _, ch := range s {
		ret += RuneWidth(ch)
	}
	return ret
}

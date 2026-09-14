package editor

import (
	"github.com/rivo/uniseg"
)

func previousGraphemeStart(line []rune, column int) int {
	if column == 0 {
		return 0
	}

	text := string(line)

	byteColumn := len(string(line[:column]))

	gr := uniseg.NewGraphemes(text)
	for gr.Next() {
		start, end := gr.Positions()

		if end == byteColumn {
			return len([]rune(text[:start]))
		}
	}

	return 0
}

func nextGraphemeEnd(line []rune, column int) int {
	text := string(line)
	byteColumn := len(string(line[:column]))

	gr := uniseg.NewGraphemes(text)
	for gr.Next() {
		start, end := gr.Positions()

		if start == byteColumn {
			return len([]rune(text[:end]))
		}
	}

	return len(line)
}

func normalizeGraphemeColumn(line []rune, column int) int {
	if column >= len(line) {
		return len(line)
	}

	text := string(line)
	byteColumn := len(string(line[:column]))

	gr := uniseg.NewGraphemes(text)

	for gr.Next() {
		start, end := gr.Positions()

		if byteColumn > start && byteColumn < end {
			return len([]rune(text[:start]))
		}
	}

	return column
}

func displayWidth(line []rune) int {
	return uniseg.StringWidth(string(line))
}

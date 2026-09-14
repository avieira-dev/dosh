package editor

import (
	"fmt"
	"unicode"

	"github.com/avieira-dev/dosh/internal/input"
	"github.com/avieira-dev/dosh/internal/terminal"
)

type Editor struct {
	Lines         []Line
	Row           int
	Column        int
	DesiredColumn int
	ScrollRow     int
	StatusMessage string
}

var wordDelimiters = []rune{
	'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	':', ';', '<', '=', '>', '?', '@',
	'[', '\\', ']', '^', '_', '`',
	'{', '|', '}', '~',
}

func isWordDelimiter(value rune) bool {
	for _, sb := range wordDelimiters {
		if sb == value {
			return true
		}
	}

	return false
}

func isSpace(value rune) bool {
	return unicode.IsSpace(value)
}

func moveToWordStart(line []rune, column int) int {
	for column > 0 && isSpace(line[column-1]) {
		column--
	}

	if column > 0 && isWordDelimiter(line[column-1]) {
		return column - 1
	}

	for column > 0 && !isSpace(line[column-1]) && !isWordDelimiter(line[column-1]) {
		column--
	}

	return column
}

func moveToWordEnd(line []rune, column int) int {
	for column < len(line) && isSpace(line[column]) {
		column++
	}

	if column < len(line) && isWordDelimiter(line[column]) {
		return column + 1
	}

	for column < len(line) && !isSpace(line[column]) && !isWordDelimiter(line[column]) {
		column++
	}

	return column
}

func (ed *Editor) Insert(value input.SimpleKey) {
	ed.Lines[ed.Row].Content = append(ed.Lines[ed.Row].Content, 0)
	copy(ed.Lines[ed.Row].Content[ed.Column+1:], ed.Lines[ed.Row].Content[ed.Column:])
	ed.Lines[ed.Row].Content[ed.Column] = rune(value)
	ed.Column++
	ed.DesiredColumn = ed.Column
}

func (ed *Editor) Backspace() {
	if ed.Column > 0 {
		current := ed.Lines[ed.Row].Content
		start := previousGraphemeStart(current, ed.Column)

		ed.Lines[ed.Row].Content = append(current[:start], current[ed.Column:]...)
		ed.Column = start
		ed.DesiredColumn = ed.Column
		return
	}

	if ed.Row > 0 {
		previous := ed.Lines[ed.Row-1].Content
		current := ed.Lines[ed.Row].Content

		ed.Column = len(previous)
		ed.DesiredColumn = ed.Column
		ed.Lines[ed.Row-1].Content = append(previous, current...)
		ed.Lines = append(ed.Lines[:ed.Row], ed.Lines[ed.Row+1:]...)
		ed.Row--
	}
}

func (ed *Editor) Delete() {
	currentLine := ed.Lines[ed.Row].Content

	if ed.Column < len(currentLine) {
		end := nextGraphemeEnd(currentLine, ed.Column)

		ed.Lines[ed.Row].Content = append(
			currentLine[:ed.Column],
			currentLine[end:]...,
		)

		return
	}

	if ed.Column == len(currentLine) && ed.Row < len(ed.Lines)-1 {
		nextLine := ed.Lines[ed.Row+1].Content
		ed.Lines[ed.Row].Content = append(currentLine, nextLine...)
		ed.Lines = append(ed.Lines[:ed.Row+1], ed.Lines[ed.Row+2:]...)
	}
}

func (ed *Editor) DeleteLineContent() {
	ed.Lines[ed.Row].Content = nil
	ed.Column = 0
	ed.DesiredColumn = ed.Column
}

func (ed *Editor) Tab() {
	for i := 0; i < 4; i++ {
		ed.Insert(input.SimpleKey(' '))
	}
}

func (ed *Editor) Enter() {
	line := ed.Lines[ed.Row].Content
	before := line[:ed.Column]
	after := line[ed.Column:]

	ed.Lines[ed.Row].Content = before

	newLine := Line{
		Content: after,
	}

	ed.Lines = append(ed.Lines, Line{})
	copy(ed.Lines[ed.Row+2:], ed.Lines[ed.Row+1:])
	ed.Lines[ed.Row+1] = newLine

	ed.Row++
	ed.Column = 0
	ed.DesiredColumn = ed.Column
}

func (ed *Editor) MoveUp() {
	if ed.Row > 0 {
		ed.Row--

		currentLineLength := len(ed.Lines[ed.Row].Content)

		if currentLineLength < ed.DesiredColumn {
			ed.Column = currentLineLength
		} else {
			ed.Column = ed.DesiredColumn
		}

		ed.Column = normalizeGraphemeColumn(
			ed.Lines[ed.Row].Content,
			ed.Column,
		)
	}
}

func (ed *Editor) MoveDown() {
	if ed.Row < len(ed.Lines)-1 {
		ed.Row++

		currentLineLength := len(ed.Lines[ed.Row].Content)

		if currentLineLength < ed.DesiredColumn {
			ed.Column = currentLineLength
		} else {
			ed.Column = ed.DesiredColumn
		}

		ed.Column = normalizeGraphemeColumn(
			ed.Lines[ed.Row].Content,
			ed.Column,
		)
	}
}

func (ed *Editor) Scroll(size terminal.Size) {
	editorHeight := size.Height - 1

	if ed.Row < ed.ScrollRow {
		ed.ScrollRow = ed.Row
	}

	if ed.Row >= ed.ScrollRow+editorHeight {
		ed.ScrollRow = ed.Row - editorHeight + 1
	}

	maxScrollRow := len(ed.Lines) - editorHeight

	if maxScrollRow < 0 {
		maxScrollRow = 0
	}

	if ed.ScrollRow > maxScrollRow {
		ed.ScrollRow = maxScrollRow
	}
}

func (ed *Editor) MoveLeft() {
	if ed.Column == 0 {
		if ed.Row > 0 {
			ed.Row--
			ed.Column = len(ed.Lines[ed.Row].Content)
			ed.DesiredColumn = ed.Column
		}
	} else {
		ed.Column = previousGraphemeStart(ed.Lines[ed.Row].Content, ed.Column)
		ed.DesiredColumn = ed.Column
	}
}

func (ed *Editor) MoveRight() {
	if ed.Column < len(ed.Lines[ed.Row].Content) {
		ed.Column = nextGraphemeEnd(ed.Lines[ed.Row].Content, ed.Column)
		ed.DesiredColumn = ed.Column
	} else if ed.Row < len(ed.Lines)-1 {
		ed.Row++
		ed.Column = 0
		ed.DesiredColumn = ed.Column
	}
}

func (ed *Editor) MoveWordLeft() {
	if ed.Column == 0 {
		if ed.Row == 0 {
			return
		}

		ed.Row--
		ed.Column = len(ed.Lines[ed.Row].Content)
	}

	ed.Column = moveToWordStart(ed.Lines[ed.Row].Content, ed.Column)

	ed.Column = normalizeGraphemeColumn(
		ed.Lines[ed.Row].Content,
		ed.Column,
	)

	ed.DesiredColumn = ed.Column
}

func (ed *Editor) MoveWordRight() {
	if ed.Column >= len(ed.Lines[ed.Row].Content) {
		if ed.Row >= len(ed.Lines)-1 {
			return
		}

		ed.Row++
		ed.Column = 0
	}

	ed.Column = moveToWordEnd(ed.Lines[ed.Row].Content, ed.Column)

	ed.Column = normalizeGraphemeColumn(
		ed.Lines[ed.Row].Content,
		ed.Column,
	)

	ed.DesiredColumn = ed.Column
}

func (ed *Editor) Home() {
	ed.Column = 0
	ed.DesiredColumn = ed.Column
}

func (ed *Editor) End() {
	ed.Column = len(ed.Lines[ed.Row].Content)
	ed.DesiredColumn = ed.Column
}

func Render(ed *Editor, size terminal.Size) {
	terminal.ClearScreen()

	editorHeight := size.Height - 1

	linesToRender := len(ed.Lines) - ed.ScrollRow
	if linesToRender > editorHeight {
		linesToRender = editorHeight
	}

	for i := 0; i < linesToRender; i++ {
		lineIndex := ed.ScrollRow + i
		line := ed.Lines[lineIndex]

		if lineIndex == ed.Row {
			fmt.Print(string(line.Content[:ed.Column]))
			fmt.Print("\0337")
			fmt.Print(string(line.Content[ed.Column:]))
		} else {
			fmt.Print(string(line.Content))
		}

		if i < linesToRender-1 {
			fmt.Print("\r\n")
		}
	}

	fmt.Printf("\033[%d;1H", size.Height)
	fmt.Printf("Line %d, Column %d	", ed.Row+1, ed.Column+1)

	if ed.StatusMessage != "" {
		fmt.Print(ed.StatusMessage)
	} else {
		fmt.Print(terminal.BgBlue + terminal.White + " ^S " + terminal.Reset + " Save")
	}

	fmt.Print("\0338")
}

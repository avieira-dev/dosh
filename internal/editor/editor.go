package editor

import (
	"fmt"
	"os"
	"strings"
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
	Dirty         bool
	SearchQuery   string
}

var wordDelimiters = []rune{
	'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	':', ';', '<', '=', '>', '?', '@',
	'[', '\\', ']', '^', '_', '`',
	'{', '|', '}', '~',
}

func isWordDelimiter(value rune) bool {
	for _, delimiter := range wordDelimiters {
		if delimiter == value {
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

func findAllMatchStarts(line []rune, query []rune) []int {
	var starts []int

	if len(query) == 0 || len(query) > len(line) {
		return starts
	}

	for i := 0; i <= len(line)-len(query); i++ {
		match := true

		for j := range query {
			if line[i+j] != query[j] {
				match = false
				break
			}
		}

		if match {
			starts = append(starts, i)
		}
	}

	return starts
}

func (ed *Editor) findNextMatch() (int, int) {
	query := []rune(ed.SearchQuery)

	if len(query) == 0 {
		return -1, -1
	}

	startRow := ed.Row
	startColumn := ed.Column

	for row := startRow; row < len(ed.Lines); row++ {
		line := ed.Lines[row].Content
		minStart := 0

		if row == startRow {
			minStart = startColumn + 1
		}

		for _, start := range findAllMatchStarts(line, query) {
			if start >= minStart {
				return row, start
			}
		}
	}

	for row := 0; row <= startRow; row++ {
		line := ed.Lines[row].Content
		maxStart := len(line)

		if row == startRow {
			maxStart = startColumn
		}

		for _, start := range findAllMatchStarts(line, query) {
			if start <= maxStart {
				return row, start
			}
		}
	}

	return -1, -1
}

func (ed *Editor) FindNextMatch() (int, int) {
	return ed.findNextMatch()
}

func (ed *Editor) Insert(value input.SimpleKey) {
	ed.Lines[ed.Row].Content = append(ed.Lines[ed.Row].Content, 0)
	copy(ed.Lines[ed.Row].Content[ed.Column+1:], ed.Lines[ed.Row].Content[ed.Column:])
	ed.Lines[ed.Row].Content[ed.Column] = rune(value)
	ed.Column++
	ed.DesiredColumn = ed.Column
	ed.Dirty = true
}

func (ed *Editor) Backspace() {
	if ed.Column > 0 {
		current := ed.Lines[ed.Row].Content
		start := previousGraphemeStart(current, ed.Column)

		ed.Lines[ed.Row].Content = append(current[:start], current[ed.Column:]...)
		ed.Column = start
		ed.DesiredColumn = ed.Column
		ed.Dirty = true
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
		ed.Dirty = true
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

		ed.Dirty = true

		return
	}

	if ed.Column == len(currentLine) && ed.Row < len(ed.Lines)-1 {
		nextLine := ed.Lines[ed.Row+1].Content
		ed.Lines[ed.Row].Content = append(currentLine, nextLine...)
		ed.Lines = append(ed.Lines[:ed.Row+1], ed.Lines[ed.Row+2:]...)
		ed.Dirty = true
	}
}

func (ed *Editor) DeleteLineContent() {
	if len(ed.Lines[ed.Row].Content) > 0 {
		ed.Lines[ed.Row].Content = nil
		ed.Dirty = true
	}

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
	ed.Dirty = true
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
	editorHeight := size.Height - 3

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

func gutterWidth(lineCount int) int {
	digits := len(fmt.Sprintf("%d", lineCount))

	if digits < 3 {
		digits = 3
	}

	return digits + 1
}

func writeHighlightedLine(buf *strings.Builder, content []rune, query []rune, isCurrentLine bool) {
	if len(query) == 0 {
		buf.WriteString(string(content))
		return
	}

	baseBg := ""
	if isCurrentLine {
		baseBg = terminal.CurrentLineBg
	}

	matches := findAllMatchStarts(content, query)
	matchLen := len(query)
	matchIndex := 0
	i := 0

	for i < len(content) {
		for matchIndex < len(matches) && matches[matchIndex] < i {
			matchIndex++
		}

		if matchIndex < len(matches) && matches[matchIndex] == i {
			buf.WriteString(terminal.MatchBg)
			buf.WriteString(terminal.MatchFg)
			buf.WriteString(string(content[i : i+matchLen]))
			buf.WriteString(terminal.Reset)
			buf.WriteString(baseBg)
			i += matchLen
			matchIndex++
			continue
		}

		buf.WriteString(string(content[i]))
		i++
	}
}

func Render(ed *Editor, size terminal.Size, fileName string) {
	var buf strings.Builder

	buf.WriteString(terminal.HideCursor)
	buf.WriteString(terminal.MoveCursorSeq(1, 1))

	title := " DOSH v0.1.0 "
	titleLen := len(title)
	leftRule := 2

	buf.WriteString(terminal.RuleFg)
	buf.WriteString(strings.Repeat("─", leftRule))
	buf.WriteString(terminal.Reset)
	buf.WriteString(terminal.TitleFg)
	buf.WriteString(title)
	buf.WriteString(terminal.Reset)
	buf.WriteString(terminal.RuleFg)

	rightRule := size.Width - leftRule - titleLen
	if rightRule < 0 {
		rightRule = 0
	}

	buf.WriteString(strings.Repeat("─", rightRule))
	buf.WriteString(terminal.Reset)

	editorHeight := size.Height - 3
	gw := gutterWidth(len(ed.Lines))
	query := []rune(ed.SearchQuery)

	for i := 0; i < editorHeight; i++ {
		buf.WriteString(terminal.MoveCursorSeq(i+2, 1))
		buf.WriteString(terminal.ClearLineSeq())

		lineIndex := ed.ScrollRow + i
		if lineIndex >= len(ed.Lines) {
			continue
		}

		isCurrent := lineIndex == ed.Row
		lineNumberField := fmt.Sprintf("%*d ", gw-1, lineIndex+1)

		if isCurrent {
			buf.WriteString(terminal.CurrentLineBg)
			buf.WriteString(terminal.GutterActiveFg)
			buf.WriteString(lineNumberField)
		} else {
			buf.WriteString(terminal.GutterFg)
			buf.WriteString(lineNumberField)
		}

		buf.WriteString(terminal.Reset)

		if isCurrent {
			buf.WriteString(terminal.CurrentLineBg)
		}

		line := ed.Lines[lineIndex]
		writeHighlightedLine(&buf, line.Content, query, isCurrent)

		if isCurrent {
			padding := size.Width - gw - displayWidth(line.Content)

			if padding > 0 {
				buf.WriteString(strings.Repeat(" ", padding))
			}
		}

		buf.WriteString(terminal.Reset)
	}

	buf.WriteString(terminal.MoveCursorSeq(size.Height-1, 1))
	buf.WriteString(terminal.RuleFg)
	buf.WriteString(strings.Repeat("─", size.Width))
	buf.WriteString(terminal.Reset)

	buf.WriteString(terminal.MoveCursorSeq(size.Height, 1))
	buf.WriteString(terminal.ClearLineSeq())

	if fileName != "" {
		buf.WriteString(terminal.TitleFg)
		buf.WriteString(fileName)
		buf.WriteString(terminal.Reset)
		buf.WriteString(terminal.StatusFg)
		buf.WriteString("  ·  ")
	} else {
		buf.WriteString(terminal.StatusFg)
	}

	fmt.Fprintf(&buf, "Ln %d, Col %d", ed.Row+1, ed.Column+1)
	buf.WriteString(terminal.Reset)

	if ed.StatusMessage != "" {
		buf.WriteString("   ")
		buf.WriteString(terminal.StatusFg)
		buf.WriteString(ed.StatusMessage)
		buf.WriteString(terminal.Reset)
	} else {
		shortcuts := []struct {
			key, label string
		}{
			{"^S", "Save"},
			{"^F", "Find"},
			{"^C", "Quit"},
			{"^K", "Del Line"},
			{"^←/^→", "Word"},
		}

		buf.WriteString("        ")

		for i, s := range shortcuts {
			buf.WriteString(terminal.StatusKeyFg)
			buf.WriteString(s.key)
			buf.WriteString(terminal.Reset)
			buf.WriteString(terminal.StatusFg)
			buf.WriteString(" " + s.label)
			buf.WriteString(terminal.Reset)

			if i < len(shortcuts)-1 {
				buf.WriteString("  ")
			}
		}
	}

	cursorColumn := displayWidth(ed.Lines[ed.Row].Content[:ed.Column])

	buf.WriteString(
		terminal.MoveCursorSeq(
			ed.Row-ed.ScrollRow+2,
			cursorColumn+gw+1,
		),
	)

	buf.WriteString(terminal.ShowCursor)

	os.Stdout.WriteString(buf.String())
}

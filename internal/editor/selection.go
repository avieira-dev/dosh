package editor

import "strings"

func (ed *Editor) ClearSelection() {
	ed.selectionActive = false
	ed.selecting = false
}

func (ed *Editor) HasSelection() bool {
	return ed.selectionActive && (ed.selectionAnchorRow != ed.Row || ed.selectionAnchorColumn != ed.Column)
}

func (ed *Editor) startSelection() {
	if !ed.selectionActive {
		ed.selectionAnchorRow = ed.Row
		ed.selectionAnchorColumn = ed.Column
		ed.selectionActive = true
	}
}

func (ed *Editor) SelectLeft() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveLeft()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectRight() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveRight()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectUp() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveUp()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectDown() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveDown()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectWordLeft() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveWordLeft()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectWordRight() {
	ed.startSelection()
	ed.selecting = true
	ed.MoveWordRight()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectHome() {
	ed.startSelection()
	ed.selecting = true
	ed.Home()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) SelectEnd() {
	ed.startSelection()
	ed.selecting = true
	ed.End()
	ed.selecting = false

	if !ed.HasSelection() {
		ed.selectionActive = false
	}
}

func (ed *Editor) selectionRange() (int, int, int, int) {
	if ed.selectionAnchorRow < ed.Row || (ed.selectionAnchorRow == ed.Row && ed.selectionAnchorColumn <= ed.Column) {
		return ed.selectionAnchorRow, ed.selectionAnchorColumn, ed.Row, ed.Column
	}

	return ed.Row, ed.Column, ed.selectionAnchorRow, ed.selectionAnchorColumn
}

func (ed *Editor) selectionBounds(row int) (int, int, bool) {
	if !ed.HasSelection() {
		return 0, 0, false
	}

	startRow, startColumn, endRow, endColumn := ed.selectionRange()

	if row < startRow || row > endRow {
		return 0, 0, false
	}

	if startRow == endRow {
		return startColumn, endColumn, true
	}

	if row == startRow {
		return startColumn, len(ed.Lines[row].Content), true
	}

	if row == endRow {
		return 0, endColumn, true
	}

	return 0, len(ed.Lines[row].Content), true
}

func (ed *Editor) SelectedText() string {
	if !ed.HasSelection() {
		return ""
	}

	startRow, startColumn, endRow, endColumn := ed.selectionRange()
	var text strings.Builder

	for row := startRow; row <= endRow; row++ {
		start := 0
		end := len(ed.Lines[row].Content)

		if row == startRow {
			start = startColumn
		}

		if row == endRow {
			end = endColumn
		}

		text.WriteString(string(ed.Lines[row].Content[start:end]))

		if row < endRow {
			text.WriteByte('\n')
		}
	}

	return text.String()
}

func (ed *Editor) deleteSelection() bool {
	if !ed.HasSelection() {
		return false
	}

	startRow, startColumn, endRow, endColumn := ed.selectionRange()

	if startRow == endRow {
		line := ed.Lines[startRow].Content
		result := make([]rune, 0, len(line)-(endColumn-startColumn))
		result = append(result, line[:startColumn]...)
		result = append(result, line[endColumn:]...)
		ed.Lines[startRow].Content = result
	} else {
		first := ed.Lines[startRow].Content[:startColumn]
		last := ed.Lines[endRow].Content[endColumn:]

		merged := make([]rune, 0, len(first)+len(last))
		merged = append(merged, first...)
		merged = append(merged, last...)

		ed.Lines[startRow].Content = merged
		ed.Lines = append(ed.Lines[:startRow+1], ed.Lines[endRow+1:]...)
	}

	ed.Row = startRow
	ed.Column = startColumn
	ed.DesiredColumn = startColumn
	ed.selectionActive = false
	ed.selecting = false

	return true
}

func (ed *Editor) DeleteSelection() bool {
	if !ed.HasSelection() {
		return false
	}

	ed.beginChange("selection")

	if !ed.deleteSelection() {
		return false
	}

	ed.Dirty = true

	return true
}

func (ed *Editor) InsertText(text string) {
	if text == "" {
		return
	}

	ed.beginChange("paste")

	if ed.HasSelection() {
		ed.deleteSelection()
	}

	parts := strings.Split(text, "\n")
	line := ed.Lines[ed.Row].Content
	before := line[:ed.Column]
	after := line[ed.Column:]

	if len(parts) == 1 {
		insert := []rune(parts[0])
		content := make([]rune, 0, len(before)+len(insert)+len(after))
		content = append(content, before...)
		content = append(content, insert...)
		content = append(content, after...)

		ed.Lines[ed.Row].Content = content
		ed.Column += len(insert)
	} else {
		newLines := make([]Line, 0, len(parts))

		first := make([]rune, 0, len(before)+len([]rune(parts[0])))
		first = append(first, before...)
		first = append(first, []rune(parts[0])...)
		newLines = append(newLines, Line{Content: first})

		for _, part := range parts[1 : len(parts)-1] {
			newLines = append(newLines, Line{Content: []rune(part)})
		}

		last := []rune(parts[len(parts)-1])
		lastContent := make([]rune, 0, len(last)+len(after))
		lastContent = append(lastContent, last...)
		lastContent = append(lastContent, after...)
		newLines = append(newLines, Line{Content: lastContent})

		lines := make([]Line, 0, len(ed.Lines)+len(parts)-1)
		lines = append(lines, ed.Lines[:ed.Row]...)
		lines = append(lines, newLines...)
		lines = append(lines, ed.Lines[ed.Row+1:]...)

		ed.Lines = lines
		ed.Row += len(parts) - 1
		ed.Column = len(last)
	}

	ed.DesiredColumn = ed.Column
	ed.Dirty = true
}

package editor

import (
	"github.com/avieira-dev/dosh/internal/terminal"
	"github.com/rivo/uniseg"
)

type visualRow struct {
	Line  int
	Start int
	End   int
}

func (ed *Editor) textWidth(size terminal.Size) int {
	return max(size.Width-gutterWidth(len(ed.Lines)), 1)
}

func (ed *Editor) wrapWidth() int {
	if ed.WrapWidth <= 0 {
		return 1 << 30
	}

	return ed.WrapWidth
}

func wrapLine(line []rune, width int) [][2]int {
	if len(line) == 0 {
		return [][2]int{{0, 0}}
	}

	var segments [][2]int
	start, pos, current := 0, 0, 0

	gr := uniseg.NewGraphemes(string(line))
	for gr.Next() {
		w := gr.Width()

		if current+w > width && pos > start {
			segments = append(segments, [2]int{start, pos})
			start = pos
			current = 0
		}

		current += w
		pos += len(gr.Runes())
	}

	segments = append(segments, [2]int{start, pos})

	if current == width {
		segments = append(segments, [2]int{pos, pos})
	}

	return segments
}

func (ed *Editor) visualRows(width int) []visualRow {
	rows := make([]visualRow, 0, len(ed.Lines))

	for i, line := range ed.Lines {
		for _, seg := range wrapLine(line.Content, width) {
			rows = append(rows, visualRow{Line: i, Start: seg[0], End: seg[1]})
		}
	}

	return rows
}

func isLastOfLine(rows []visualRow, index int) bool {
	return index == len(rows)-1 || rows[index+1].Line != rows[index].Line
}

func cursorVisual(rows []visualRow, lines []Line, row, col int) (int, int) {
	for i, vr := range rows {
		if vr.Line != row || col < vr.Start {
			continue
		}

		if col < vr.End || (col == vr.End && isLastOfLine(rows, i)) {
			return i, displayWidth(lines[row].Content[vr.Start:col])
		}
	}

	return 0, 0
}

func columnAtVisualWidth(line []rune, vr visualRow, target int, lastOfLine bool) int {
	col := vr.Start
	width := 0

	for col < vr.End {
		next := min(nextGraphemeEnd(line, col), vr.End)
		w := displayWidth(line[col:next])

		if width+w > target {
			break
		}

		width += w
		col = next
	}

	if !lastOfLine && col == vr.End && col > vr.Start {
		col = previousGraphemeStart(line, col)
	}

	return col
}

func visualRowEnd(rows []visualRow, index int, lines []Line) int {
	vr := rows[index]

	if isLastOfLine(rows, index) || vr.End == vr.Start {
		return vr.End
	}

	return previousGraphemeStart(lines[vr.Line].Content, vr.End)
}

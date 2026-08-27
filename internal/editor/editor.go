package editor

import (
	"fmt"

	"github.com/avieira-dev/dosh/internal/input"
	"github.com/avieira-dev/dosh/internal/terminal"
)

type Editor struct {
	Lines  []Line
	Row    int
	Column int
}

func (ed *Editor) Insert(value input.SimpleKey) {
	ed.Lines[ed.Row].Content = append(ed.Lines[ed.Row].Content, 0)
	copy(ed.Lines[ed.Row].Content[ed.Column+1:], ed.Lines[ed.Row].Content[ed.Column:])
	ed.Lines[ed.Row].Content[ed.Column] = byte(value)
	ed.Column++
}

func (ed *Editor) Backspace() {
	if ed.Column > 0 {
		if len(ed.Lines[ed.Row].Content) > 0 {
			ed.Lines[ed.Row].Content = append(ed.Lines[ed.Row].Content[:ed.Column-1], ed.Lines[ed.Row].Content[ed.Column:]...)
			ed.Column--
		}
	} else if ed.Row > 0 {
		previous := ed.Lines[ed.Row-1].Content
		current := ed.Lines[ed.Row].Content

		ed.Column = len(previous)

		previous = append(previous, current...)
		ed.Lines[ed.Row-1].Content = previous

		ed.Lines = append(ed.Lines[:ed.Row], ed.Lines[ed.Row+1:]...)

		ed.Row--
	}
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
}

func (ed *Editor) MoveUp() {
	if ed.Row > 0 {
		if len(ed.Lines[ed.Row-1].Content) < len(ed.Lines[ed.Row].Content) {
			ed.Row--
			ed.Column = len(ed.Lines[ed.Row].Content)
		} else {
			ed.Row--
		}

	}
}

func (ed *Editor) MoveDown() {
	if ed.Row < len(ed.Lines)-1 {
		if len(ed.Lines[ed.Row+1].Content) < len(ed.Lines[ed.Row].Content) {
			ed.Row++
			ed.Column = len(ed.Lines[ed.Row].Content)
		} else {
			ed.Row++
		}
	}
}

func (ed *Editor) MoveLeft() {
	if ed.Column == 0 {
		if ed.Row > 0 {
			ed.Row--
			ed.Column = len(ed.Lines[ed.Row].Content)
		}
	} else {
		ed.Column--
	}
}

func (ed *Editor) MoveRight() {
	if ed.Column < len(ed.Lines[ed.Row].Content) {
		ed.Column++
	} else if ed.Row < len(ed.Lines)-1 {
		ed.Row++
		ed.Column = 0
	}
}

func Render(ed *Editor) {
	terminal.ClearScreen()

	for _, line := range ed.Lines {
		fmt.Print(string(line.Content))
		fmt.Print("\r\n")
	}

	fmt.Printf("\033[%d;%dH", ed.Row+1, ed.Column+1)
}

package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/avieira-dev/dosh/internal/editor"
	"github.com/avieira-dev/dosh/internal/input"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ed := editor.Editor{
		Lines: []editor.Line{
			{Content: []byte{}},
		},
	}

	running := true

	for running {
		key := input.ReadKey()

		if value, ok := key.Value.(input.SimpleKey); ok {
			ed.Insert(value)
		}

		if value, ok := key.Value.(input.SpecialKey); ok {
			switch value {
			case input.KeyArrowUp:
				ed.MoveUp()
			case input.KeyArrowDown:
				ed.MoveDown()
			case input.KeyArrowLeft:
				ed.MoveLeft()
			case input.KeyArrowRight:
				ed.MoveRight()
			case input.KeyCtrlC:
				running = false
			case input.KeyCtrlLeft:
				ed.MoveWordLeft()
			case input.KeyCtrlRight:
				ed.MoveWordRight()
			case input.KeyBackspace:
				ed.Backspace()
			case input.KeyCtrlK:
				ed.DeleteLineContent()
			case input.KeyTab:
				ed.Tab()
			case input.KeyEnter:
				ed.Enter()
			case input.KeyDelete:
				ed.Delete()
			case input.KeyHome:
				ed.Home()
			case input.KeyEnd:
				ed.End()
			}
		}

		if running {
			editor.Render(&ed)
		}
	}

	fmt.Printf("\033[%d;1H", len(ed.Lines)+1)
	fmt.Print("Dosh successfully closed!")
	fmt.Print("\r\n")
}

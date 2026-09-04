package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"github.com/avieira-dev/dosh/internal/editor"
	"github.com/avieira-dev/dosh/internal/input"
	"github.com/avieira-dev/dosh/internal/terminal"
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

	size, err := terminal.GetSize()
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}

	keys := make(chan input.Key)

	go func() {
		for {
			keys <- input.ReadKey()
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)

	running := true
	for running {
		select {
		case key := <-keys:
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
				ed.Scroll(size)
				editor.Render(&ed, size)
			}
		case <-signals:
			size, err = terminal.GetSize()
			if err != nil {
				fmt.Println("[ERROR]", err)
				running = false
				continue
			}

			ed.Scroll(size)
			editor.Render(&ed, size)
		}

	}

	fmt.Printf("\033[%d;1H", len(ed.Lines)+1)
	fmt.Print("Dosh successfully closed!")
	fmt.Print("\r\n")
}

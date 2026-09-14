package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/avieira-dev/dosh/internal/editor"
	"github.com/avieira-dev/dosh/internal/file"
	"github.com/avieira-dev/dosh/internal/input"
	"github.com/avieira-dev/dosh/internal/terminal"
)

func fileDisplayName(openFile *file.File) string {
	if openFile == nil {
		return ""
	}

	return filepath.Base(openFile.Path)
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	if len(os.Args) > 2 {
		fmt.Println("[ERROR] Invalid number of arguments!")
		return
	}

	var ed editor.Editor
	var openFile *file.File

	if len(os.Args) == 1 {
		ed = editor.Editor{
			Lines: []editor.Line{
				{Content: []rune{}},
			},
		}
	} else {
		path := os.Args[1]

		openedFile, err := file.Open(path)
		if err != nil {
			fmt.Println("[ERROR]", err)
			return
		}

		openFile = &openedFile

		ed = editor.Editor{
			Lines: openedFile.Lines,
		}
	}

	size, err := terminal.GetSize()
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}

	terminal.ClearScreen()
	ed.Scroll(size)
	editor.Render(&ed, size, fileDisplayName(openFile))

	keys := make(chan input.Key)

	go func() {
		for {
			keys <- input.ReadKey()
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)

	statusTimer := time.NewTimer(time.Hour)
	statusTimer.Stop()

	inputFileName := false
	fileName := ""
	confirmExit := false
	confirmOverwrite := false

	running := true
	for running {
		select {
		case key := <-keys:
			if confirmExit {
				if value, ok := key.Value.(input.SimpleKey); ok {
					switch value {
					case 'y', 'Y':
						running = false
					case 'n', 'N':
						confirmExit = false
						ed.StatusMessage = ""
					}
				}

				if running {
					ed.Scroll(size)
					editor.Render(&ed, size, fileDisplayName(openFile))
				}

				continue
			}

			if confirmOverwrite {
				if value, ok := key.Value.(input.SimpleKey); ok {
					switch value {
					case 'y', 'Y':
						newFile := file.File{
							Path: fileName,
						}

						if err := newFile.Save(ed.Lines); err != nil {
							ed.StatusMessage = "Error saving!"
						} else {
							openFile = &newFile
							ed.Dirty = false
							ed.StatusMessage = "Saved successfully!"
							statusTimer.Reset(2 * time.Second)
						}

						confirmOverwrite = false

					case 'n', 'N':
						confirmOverwrite = false
						ed.StatusMessage = ""
					}
				}

				editor.Render(&ed, size, fileDisplayName(openFile))
				continue
			}

			if inputFileName {
				if value, ok := key.Value.(input.SimpleKey); ok {
					fileName += string(value)
					ed.StatusMessage = "File name: " + fileName
				}

				if value, ok := key.Value.(input.SpecialKey); ok {
					switch value {
					case input.KeyBackspace:
						if len(fileName) > 0 {
							fileName = fileName[:len(fileName)-1]
							ed.StatusMessage = "File name: " + fileName
						}

					case input.KeyEnter:
						if fileName == "" {
							ed.StatusMessage = "File name cannot be empty!"
							break
						}

						if file.Exists(fileName) {
							inputFileName = false
							confirmOverwrite = true
							ed.StatusMessage = "File already exists. Overwrite? (y/n)"
							break
						}

						newFile := file.File{
							Path: fileName,
						}

						if err := newFile.Save(ed.Lines); err != nil {
							ed.StatusMessage = "Error saving!"
							break
						}

						openFile = &newFile
						inputFileName = false
						ed.Dirty = false
						ed.StatusMessage = "Saved successfully!"
						statusTimer.Reset(2 * time.Second)
					}
				}

				editor.Render(&ed, size, fileDisplayName(openFile))

				continue
			}

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
					if !ed.Dirty {
						running = false
					} else {
						confirmExit = true
						ed.StatusMessage = "Unsaved changes. Exit anyway? (y/n)"
					}
				case input.KeyCtrlLeft:
					ed.MoveWordLeft()
				case input.KeyCtrlRight:
					ed.MoveWordRight()
				case input.KeyBackspace:
					ed.Backspace()
				case input.KeyCtrlK:
					ed.DeleteLineContent()
				case input.KeyCtrlS:
					if openFile == nil {
						inputFileName = true
						fileName = ""
						ed.StatusMessage = "File name: "
					} else {
						if err := openFile.Save(ed.Lines); err != nil {
							ed.StatusMessage = "Error saving!"
						} else {
							ed.Dirty = false
							ed.StatusMessage = "Saved successfully!"
							statusTimer.Reset(2 * time.Second)
						}
					}
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
				editor.Render(&ed, size, fileDisplayName(openFile))
			}

		case <-signals:
			size, err = terminal.GetSize()
			if err != nil {
				fmt.Println("[ERROR]", err)
				running = false
				continue
			}

			terminal.ClearScreen()
			ed.Scroll(size)
			editor.Render(&ed, size, fileDisplayName(openFile))

		case <-statusTimer.C:
			ed.StatusMessage = ""
			ed.Scroll(size)
			editor.Render(&ed, size, fileDisplayName(openFile))
		}
	}

	terminal.ClearScreen()
}

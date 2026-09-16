package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/avieira-dev/dosh/internal/clipboard"
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

func saveFile(ed *editor.Editor, target *file.File, statusTimer *time.Timer) bool {
	if err := target.Save(ed.Lines); err != nil {
		ed.StatusMessage = "Error saving!"
		return false
	}

	ed.Dirty = false
	ed.StatusMessage = "Saved successfully!"
	statusTimer.Reset(2 * time.Second)

	return true
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
		ed = editor.NewEditor([]editor.Line{
			{Content: []rune{}},
		})
	} else {
		path := os.Args[1]

		openedFile, err := file.Open(path)
		if err != nil {
			fmt.Println("[ERROR]", err)
			return
		}

		openFile = &openedFile

		ed = editor.NewEditor(openedFile.Lines)
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
	inputWord := false
	inputReplace := false
	confirmExit := false
	confirmOverwrite := false
	fileName := ""

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

						if saveFile(&ed, &newFile, statusTimer) {
							openFile = &newFile
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

					case input.KeyEscape, input.KeyCtrlC:
						inputFileName = false
						fileName = ""
						ed.StatusMessage = ""

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

						if saveFile(&ed, &newFile, statusTimer) {
							openFile = &newFile
							inputFileName = false
						}
					}
				}

				editor.Render(&ed, size, fileDisplayName(openFile))
				continue
			}

			if inputReplace {
				if value, ok := key.Value.(input.SimpleKey); ok {
					ed.ReplaceQuery += string(value)
					ed.StatusMessage = "Replace with: " + ed.ReplaceQuery
				}

				if value, ok := key.Value.(input.SpecialKey); ok {
					switch value {
					case input.KeyBackspace:
						if len(ed.ReplaceQuery) > 0 {
							ed.ReplaceQuery = ed.ReplaceQuery[:len(ed.ReplaceQuery)-1]
							ed.StatusMessage = "Replace with: " + ed.ReplaceQuery
						}

					case input.KeyEscape, input.KeyCtrlC:
						inputReplace = false
						ed.ReplaceQuery = ""
						ed.StatusMessage = ""

					case input.KeyEnter:
						if ed.ReplaceCurrentMatch() {
							ed.StatusMessage = "Replaced!"
						} else {
							ed.StatusMessage = "No match found!"
							inputReplace = false
						}

						statusTimer.Reset(2 * time.Second)

					case input.KeyCtrlA:
						count := ed.ReplaceAll()
						inputReplace = false
						ed.ReplaceQuery = ""
						ed.StatusMessage = fmt.Sprintf("Replaced %d occurrence(s)!", count)
						statusTimer.Reset(2 * time.Second)
					}
				}

				editor.Render(&ed, size, fileDisplayName(openFile))
				continue
			}

			if inputWord {
				if value, ok := key.Value.(input.SimpleKey); ok {
					ed.SearchQuery += string(value)
					ed.StatusMessage = "Search: " + ed.SearchQuery
				}

				if value, ok := key.Value.(input.SpecialKey); ok {
					switch value {
					case input.KeyBackspace:
						if len(ed.SearchQuery) > 0 {
							ed.SearchQuery = ed.SearchQuery[:len(ed.SearchQuery)-1]
							ed.StatusMessage = "Search: " + ed.SearchQuery
						}

					case input.KeyEscape, input.KeyCtrlC:
						inputWord = false
						ed.SearchQuery = ""
						ed.StatusMessage = ""

					case input.KeyEnter:
						if ed.SearchQuery == "" {
							ed.StatusMessage = "The search cannot be empty!"
							break
						}

						row, column := ed.FindNextMatch()

						if row == -1 {
							ed.StatusMessage = "No match found!"
							break
						}

						ed.Row = row
						ed.Column = column
						ed.DesiredColumn = column
						inputWord = false
						ed.StatusMessage = ""
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

				case input.KeyShiftArrowUp:
					ed.SelectUp()

				case input.KeyShiftArrowDown:
					ed.SelectDown()

				case input.KeyShiftArrowLeft:
					ed.SelectLeft()

				case input.KeyShiftArrowRight:
					ed.SelectRight()

				case input.KeyShiftCtrlLeft:
					ed.SelectWordLeft()

				case input.KeyShiftCtrlRight:
					ed.SelectWordRight()

				case input.KeyShiftHome:
					ed.SelectHome()

				case input.KeyShiftEnd:
					ed.SelectEnd()

				case input.KeyCtrlC:
					if ed.HasSelection() {
						if err := clipboard.Copy(ed.SelectedText()); err != nil {
							ed.StatusMessage = "Error copying selection!"
						} else {
							ed.StatusMessage = "Copied!"
						}

						statusTimer.Reset(2 * time.Second)
					}

				case input.KeyCtrlX:
					if ed.HasSelection() {
						if err := clipboard.Copy(ed.SelectedText()); err != nil {
							ed.StatusMessage = "Error cutting selection!"
						} else if ed.DeleteSelection() {
							ed.StatusMessage = "Cut!"
						}

						statusTimer.Reset(2 * time.Second)
					}

				case input.KeyCtrlV:
					text, err := clipboard.Paste()

					if err != nil {
						ed.StatusMessage = "Clipboard is empty!"
					} else {
						ed.InsertText(text)
						ed.StatusMessage = "Pasted!"
					}

					statusTimer.Reset(2 * time.Second)

				case input.KeyCtrlQ:
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
						saveFile(&ed, openFile, statusTimer)
					}

				case input.KeyCtrlF:
					ed.ClearSelection()
					inputWord = true
					ed.SearchQuery = ""
					ed.StatusMessage = "Search: "

				case input.KeyCtrlR:
					ed.ClearSelection()

					if ed.SearchQuery == "" {
						ed.StatusMessage = "Search for a term first (Ctrl+F)!"
						statusTimer.Reset(2 * time.Second)
					} else {
						inputReplace = true
						ed.ReplaceQuery = ""
						ed.StatusMessage = "Replace with: "
					}

				case input.KeyCtrlZ:
					ed.Undo()
					statusTimer.Reset(2 * time.Second)

				case input.KeyCtrlY:
					ed.Redo()
					statusTimer.Reset(2 * time.Second)

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

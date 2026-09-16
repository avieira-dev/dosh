package clipboard

import (
	"errors"
	"os/exec"
	"strings"
)

var internalClipboard string

func Copy(text string) error {
	internalClipboard = text

	command, args, err := copyCommand()

	if err != nil {
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(text)

	_ = cmd.Run()

	return nil
}

func Paste() (string, error) {
	command, args, err := pasteCommand()

	if err == nil {
		output, err := exec.Command(command, args...).Output()

		if err == nil {
			return string(output), nil
		}
	}

	if internalClipboard != "" {
		return internalClipboard, nil
	}

	return "", errors.New("clipboard is empty")
}

func copyCommand() (string, []string, error) {
	commands := []struct {
		name string
		args []string
	}{
		{"wl-copy", nil},
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
	}

	for _, command := range commands {
		if _, err := exec.LookPath(command.name); err == nil {
			return command.name, command.args, nil
		}
	}

	return "", nil, errors.New("clipboard command not found")
}

func pasteCommand() (string, []string, error) {
	commands := []struct {
		name string
		args []string
	}{
		{"wl-paste", nil},
		{"xclip", []string{"-selection", "clipboard", "-o"}},
		{"xsel", []string{"--clipboard", "--output"}},
	}

	for _, command := range commands {
		if _, err := exec.LookPath(command.name); err == nil {
			return command.name, command.args, nil
		}
	}

	return "", nil, errors.New("clipboard command not found")
}

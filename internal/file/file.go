package file

import (
	"os"
	"strings"

	"github.com/avieira-dev/dosh/internal/editor"
)

type File struct {
	Lines []editor.Line
	Path  string
}

func Open(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}

	lines := strings.Split(string(data), "\n")

	file := File{
		Lines: make([]editor.Line, 0, len(lines)),
		Path:  path,
	}

	for _, line := range lines {
		file.Lines = append(file.Lines, editor.Line{
			Content: []rune(line),
		})
	}

	return file, nil
}

func (f *File) Save(lines []editor.Line) error {
	var content strings.Builder

	for i, line := range lines {
		content.WriteString(string(line.Content))

		if i < len(lines)-1 {
			content.WriteString("\n")
		}
	}

	return os.WriteFile(f.Path, []byte(content.String()), 0644)
}

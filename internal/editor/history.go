package editor

type Snapshot struct {
	Lines  []Line
	Row    int
	Column int
}

type History struct {
	undoStack []Snapshot
	redoStack []Snapshot
}

func NewHistory() *History {
	return &History{}
}

func cloneLines(lines []Line) []Line {
	cloned := make([]Line, len(lines))

	for i, line := range lines {
		content := make([]rune, len(line.Content))
		copy(content, line.Content)
		cloned[i] = Line{Content: content}
	}

	return cloned
}

func (h *History) Push(snapshot Snapshot) {
	h.undoStack = append(h.undoStack, snapshot)
	h.redoStack = nil
}

func (h *History) Undo(current Snapshot) (Snapshot, bool) {
	if len(h.undoStack) == 0 {
		return Snapshot{}, false
	}

	previous := h.undoStack[len(h.undoStack)-1]
	h.undoStack = h.undoStack[:len(h.undoStack)-1]
	h.redoStack = append(h.redoStack, current)

	return previous, true
}

func (h *History) Redo(current Snapshot) (Snapshot, bool) {
	if len(h.redoStack) == 0 {
		return Snapshot{}, false
	}

	next := h.redoStack[len(h.redoStack)-1]
	h.redoStack = h.redoStack[:len(h.redoStack)-1]
	h.undoStack = append(h.undoStack, current)

	return next, true
}

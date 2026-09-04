package input

type SimpleKey rune
type SpecialKey int
type KeyType int

const (
	KeyUnknown SpecialKey = iota

	// Navigation
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyHome
	KeyEnd
	KeyCtrlLeft
	KeyCtrlRight

	// Editing
	KeyBackspace
	KeyDelete
	KeyEnter
	KeyTab
	KeyCtrlK

	// Control
	KeyEscape
	KeyCtrlC
)

const (
	Simple KeyType = iota
	Special
)

type Key struct {
	Type  KeyType
	Value any
}

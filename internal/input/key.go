package input

type SimpleKey rune
type SpecialKey int
type KeyType int

const (
	KeyUnknown SpecialKey = iota

	KeyArrowUp
	KeyArrowDown
	KeyArrowRight
	KeyArrowLeft

	KeyEnter
	KeyBackspace
	KeyTab
	KeyEscape
	KeyCtrlC
	KeyDelete
)

const (
	Simple KeyType = iota
	Special
)

type Key struct {
	Type  KeyType
	Value any
}

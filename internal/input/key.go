package input

type SimpleKey rune
type SpecialKey int
type KeyType int

const (
	KeyUnknown SpecialKey = iota

	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyHome
	KeyEnd
	KeyCtrlLeft
	KeyCtrlRight

	KeyBackspace
	KeyDelete
	KeyEnter
	KeyTab
	KeyCtrlK
	KeyCtrlS

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

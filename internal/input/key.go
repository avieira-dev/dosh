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
	KeyCtrlF

	KeyBackspace
	KeyDelete
	KeyEnter
	KeyTab
	KeyCtrlK
	KeyCtrlS
	KeyCtrlZ
	KeyCtrlY
	KeyCtrlR
	KeyCtrlA

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

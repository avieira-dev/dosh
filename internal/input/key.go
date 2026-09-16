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

	KeyShiftArrowUp
	KeyShiftArrowDown
	KeyShiftArrowLeft
	KeyShiftArrowRight
	KeyShiftHome
	KeyShiftEnd
	KeyShiftCtrlLeft
	KeyShiftCtrlRight

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
	KeyCtrlV
	KeyCtrlX
	KeyCtrlQ

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

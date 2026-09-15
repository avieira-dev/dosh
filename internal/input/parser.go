package input

import (
	"unicode/utf8"
)

type ParserState int

const (
	StateNormal ParserState = iota
	StateEscape
	StateCSI
)

type Parser struct {
	State     ParserState
	Buffer    []byte
	UTFBuffer []byte
}

var (
	arrowRightSequence = []byte{27, '[', 'C'}
	arrowLeftSequence  = []byte{27, '[', 'D'}
	deleteSequence     = []byte{27, '[', '3', '~'}
	ctrlRightSequence  = []byte{27, '[', '1', ';', '5', 'C'}
	ctrlLeftSequence   = []byte{27, '[', '1', ';', '5', 'D'}
)

func matchesSequence(buffer []byte, sequence []byte) bool {
	if len(buffer) != len(sequence) {
		return false
	}

	for i, b := range buffer {
		if b != sequence[i] {
			return false
		}
	}

	return true
}

func (parser *Parser) parseUTF8(value byte) (Key, bool) {
	parser.UTFBuffer = append(parser.UTFBuffer, value)

	if !utf8.FullRune(parser.UTFBuffer) {
		return Key{}, false
	}

	r, size := utf8.DecodeRune(parser.UTFBuffer)

	if r == utf8.RuneError && size == 1 {
		parser.UTFBuffer = parser.UTFBuffer[:0]
		return Key{}, false
	}

	parser.UTFBuffer = parser.UTFBuffer[:0]

	return Key{Type: Simple, Value: SimpleKey(r)}, true
}

func (parser *Parser) Parse(value byte) (Key, bool) {
	switch parser.State {
	case StateNormal:
		switch value {
		case 1:
			return Key{Type: Special, Value: KeyCtrlA}, true
		case 3:
			return Key{Type: Special, Value: KeyCtrlC}, true
		case 6:
			return Key{Type: Special, Value: KeyCtrlF}, true
		case 9:
			return Key{Type: Special, Value: KeyTab}, true
		case 11:
			return Key{Type: Special, Value: KeyCtrlK}, true
		case 13:
			return Key{Type: Special, Value: KeyEnter}, true
		case 18:
			return Key{Type: Special, Value: KeyCtrlR}, true
		case 19:
			return Key{Type: Special, Value: KeyCtrlS}, true
		case 25:
			return Key{Type: Special, Value: KeyCtrlY}, true
		case 26:
			return Key{Type: Special, Value: KeyCtrlZ}, true
		case 127:
			return Key{Type: Special, Value: KeyBackspace}, true
		case 27:
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateEscape
			return Key{}, false
		default:
			return parser.parseUTF8(value)
		}
	case StateEscape:
		if value == '[' {
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateCSI
			return Key{}, false
		}
	case StateCSI:
		parser.Buffer = append(parser.Buffer, value)

		var specialKey SpecialKey

		switch value {
		case 'A':
			specialKey = KeyArrowUp
		case 'B':
			specialKey = KeyArrowDown
		case 'F':
			specialKey = KeyEnd
		case 'H':
			specialKey = KeyHome
		case '~':
			if matchesSequence(parser.Buffer, deleteSequence) {
				specialKey = KeyDelete
			}
		case 'C':
			if matchesSequence(parser.Buffer, arrowRightSequence) {
				specialKey = KeyArrowRight
			}

			if matchesSequence(parser.Buffer, ctrlRightSequence) {
				specialKey = KeyCtrlRight
			}
		case 'D':
			if matchesSequence(parser.Buffer, arrowLeftSequence) {
				specialKey = KeyArrowLeft
			}

			if matchesSequence(parser.Buffer, ctrlLeftSequence) {
				specialKey = KeyCtrlLeft
			}
		}

		if specialKey > 0 {
			parser.State = StateNormal
			parser.Buffer = parser.Buffer[:0]

			return Key{Type: Special, Value: specialKey}, true
		}

	}

	return Key{}, false
}

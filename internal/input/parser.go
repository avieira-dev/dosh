package input

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type ParserState int

const (
	StateNormal ParserState = iota
	StateEscape
	StateCSI
	StateSS3
)

type Parser struct {
	State     ParserState
	Buffer    []byte
	UTFBuffer []byte
}

func (parser *Parser) reset() {
	parser.State = StateNormal
	parser.Buffer = parser.Buffer[:0]
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

func parseCSIParams(sequence string) (rune, []int, bool) {
	if sequence == "" {
		return 0, nil, false
	}

	final := rune(sequence[len(sequence)-1])
	parameters := sequence[:len(sequence)-1]

	if parameters == "" {
		return final, nil, true
	}

	parts := strings.Split(parameters, ";")
	params := make([]int, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			params = append(params, 0)
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return 0, nil, false
		}

		params = append(params, value)
	}

	return final, params, true
}

func csiModifier(params []int) int {
	if len(params) < 2 {
		return 1
	}

	return params[len(params)-1]
}

func csiPrimary(params []int) int {
	if len(params) == 0 {
		return 0
	}

	return params[0]
}

func (parser *Parser) parseCSI() (Key, bool) {
	if len(parser.Buffer) < 3 {
		parser.reset()
		return Key{}, false
	}

	sequence := string(parser.Buffer[2:])
	final, params, ok := parseCSIParams(sequence)

	if !ok {
		parser.reset()
		return Key{}, false
	}

	primary := csiPrimary(params)
	modifier := csiModifier(params)

	switch final {
	case 'A':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyArrowUp}, true
		case 2:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftArrowUp}, true
		}

	case 'B':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyArrowDown}, true
		case 2:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftArrowDown}, true
		}

	case 'C':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyArrowRight}, true
		case 2:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftArrowRight}, true
		case 5:
			parser.reset()
			return Key{Type: Special, Value: KeyCtrlRight}, true
		case 6:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftCtrlRight}, true
		}

	case 'D':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyArrowLeft}, true
		case 2:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftArrowLeft}, true
		case 5:
			parser.reset()
			return Key{Type: Special, Value: KeyCtrlLeft}, true
		case 6:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftCtrlLeft}, true
		}

	case 'H':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyHome}, true
		case 2, 5, 6:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftHome}, true
		}

	case 'F':
		switch modifier {
		case 1:
			parser.reset()
			return Key{Type: Special, Value: KeyEnd}, true
		case 2, 5, 6:
			parser.reset()
			return Key{Type: Special, Value: KeyShiftEnd}, true
		}

	case '~':
		switch primary {
		case 1, 7:
			switch modifier {
			case 1:
				parser.reset()
				return Key{Type: Special, Value: KeyHome}, true
			case 2, 5, 6:
				parser.reset()
				return Key{Type: Special, Value: KeyShiftHome}, true
			}

		case 3:
			parser.reset()
			return Key{Type: Special, Value: KeyDelete}, true

		case 4, 8:
			switch modifier {
			case 1:
				parser.reset()
				return Key{Type: Special, Value: KeyEnd}, true
			case 2, 5, 6:
				parser.reset()
				return Key{Type: Special, Value: KeyShiftEnd}, true
			}
		}
	}

	parser.reset()
	return Key{}, false
}

func (parser *Parser) parseSS3() (Key, bool) {
	if len(parser.Buffer) < 3 {
		parser.reset()
		return Key{}, false
	}

	switch parser.Buffer[2] {
	case 'A':
		parser.reset()
		return Key{Type: Special, Value: KeyArrowUp}, true

	case 'B':
		parser.reset()
		return Key{Type: Special, Value: KeyArrowDown}, true

	case 'C':
		parser.reset()
		return Key{Type: Special, Value: KeyArrowRight}, true

	case 'D':
		parser.reset()
		return Key{Type: Special, Value: KeyArrowLeft}, true

	case 'H':
		parser.reset()
		return Key{Type: Special, Value: KeyHome}, true

	case 'F':
		parser.reset()
		return Key{Type: Special, Value: KeyEnd}, true
	}

	parser.reset()
	return Key{}, false
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
		case 17:
			return Key{Type: Special, Value: KeyCtrlQ}, true
		case 18:
			return Key{Type: Special, Value: KeyCtrlR}, true
		case 19:
			return Key{Type: Special, Value: KeyCtrlS}, true
		case 22:
			return Key{Type: Special, Value: KeyCtrlV}, true
		case 24:
			return Key{Type: Special, Value: KeyCtrlX}, true
		case 25:
			return Key{Type: Special, Value: KeyCtrlY}, true
		case 26:
			return Key{Type: Special, Value: KeyCtrlZ}, true
		case 127:
			return Key{Type: Special, Value: KeyBackspace}, true
		case 27:
			parser.Buffer = append(parser.Buffer[:0], value)
			parser.State = StateEscape
			return Key{}, false
		default:
			return parser.parseUTF8(value)
		}

	case StateEscape:
		switch value {
		case '[':
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateCSI
			return Key{}, false

		case 'O':
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateSS3
			return Key{}, false
		}

		parser.reset()
		return Key{Type: Special, Value: KeyEscape}, true

	case StateCSI:
		parser.Buffer = append(parser.Buffer, value)

		if value >= 0x40 && value <= 0x7e {
			return parser.parseCSI()
		}

		return Key{}, false

	case StateSS3:
		parser.Buffer = append(parser.Buffer, value)
		return parser.parseSS3()
	}

	return Key{}, false
}

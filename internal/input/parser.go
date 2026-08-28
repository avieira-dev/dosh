package input

type ParserState int

const (
	StateNormal ParserState = iota
	StateEscape
	StateCSI
)

type Parser struct {
	State ParserState
	Buffer []byte
}

func (parser *Parser) Parse(value byte) (Key, bool) {
	switch parser.State {
	case StateNormal:
		switch value {
		case 3:
			return Key{Type: Special, Value: KeyCtrlC}, true
		case 9:
			return Key{Type: Special, Value: KeyTab}, true
		case 13:
			return Key{Type: Special, Value: KeyEnter}, true
		case 127:
			return Key{Type: Special, Value: KeyBackspace}, true
		case 27:
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateEscape
			return Key{}, false
		default:
			return Key{Type: Simple, Value: SimpleKey(value)}, true
		}
	case StateEscape:
		if value == 91 {
			parser.Buffer = append(parser.Buffer, value)
			parser.State = StateCSI
			return Key{}, false
		}
	case StateCSI:
		parser.Buffer = append(parser.Buffer, value)

		var specialKey SpecialKey

		switch value {
		case 65:
			specialKey = KeyArrowUp
		case 66:
			specialKey = KeyArrowDown
		case 67:
			specialKey = KeyArrowRight
		case 68:
			specialKey = KeyArrowLeft
		case 70:
			specialKey = KeyEnd
		case 72:
			specialKey = KeyHome
		case 126:
			if len(parser.Buffer) == 4 {
				if parser.Buffer[0] == 27 && parser.Buffer[1] == '[' && parser.Buffer[2] == '3' && parser.Buffer[3] == '~' {
					specialKey = KeyDelete
				}
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

package input

import (
	"os"
)

func ReadKey() Key {
	buffer := make([]byte, 3)

	n, err := os.Stdin.Read(buffer)
	if err != nil {
		return Key{}
	}

	if n == 1 {
		switch buffer[0] {
		case 3:
			return Key{
				Type: Special,
				Value: KeyCtrlC,
			}
		case 9:
			return Key{
				Type: Special,
				Value: KeyTab,
			}
		case 13:
			return Key{
				Type: Special,
				Value: KeyEnter,
			}
		case 27:
			return Key{
				Type: Special,
				Value: KeyEscape,
			}
		case 127:
			return Key{
				Type: Special,
				Value: KeyBackspace,
			}
		default:
			return Key{
				Type:  Simple,
				Value: SimpleKey(buffer[0]),
			}
		}
	}

	if n == 3 {
		var key SpecialKey

		switch buffer[2] {
		case 65:
			key = KeyArrowUp
		case 66:
			key = KeyArrowDown
		case 67:
			key = KeyArrowRight
		case 68:
			key = KeyArrowLeft
		default:
			key = KeyUnknown
		}

		return Key{
			Type: Special,
			Value: key,
		}
	}

	return Key{}
}
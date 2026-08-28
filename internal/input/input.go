package input

import (
	"os"
)

func ReadKey() Key {
	parser := Parser{}

	buffer := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			return Key{}
		}

		key, complete := parser.Parse(buffer[0])

		if complete {
			return key
		}
	}
}

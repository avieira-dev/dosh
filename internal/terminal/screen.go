package terminal

import (
	"fmt"
)

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
	fmt.Print("\033[3J")
}

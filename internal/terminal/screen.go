package terminal

import (
	"fmt"
	"syscall"
	"unsafe"
)

type Size struct {
	Width  int
	Height int
}

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
	fmt.Print("\033[3J")
}

func GetSize() (Size, error) {
	var windowSize struct {
		Rows    uint16
		Columns uint16
		XPixels uint16
		YPixels uint16
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&windowSize)))

	if errno != 0 {
		return Size{}, errno
	}

	return Size{Width: int(windowSize.Columns), Height: int(windowSize.Rows)}, nil
}

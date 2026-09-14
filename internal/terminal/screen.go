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

const (
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"
)

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
	fmt.Print("\033[3J")
}

func MoveCursorSeq(row, column int) string {
	return fmt.Sprintf("\033[%d;%dH", row, column)
}

func ClearLineSeq() string {
	return "\033[2K"
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

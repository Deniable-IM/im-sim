package logger

import "fmt"

type Function string

const (
	SaveCursorPos    Function = "\0337"
	RestoreCursorPos Function = "\0338"
	ClearEntierLine  Function = "\033[2K"
	MoveCursorUp     Function = "\033[A"
	MoveCursorDown   Function = "\033[B"
	HideCursor       Function = "\033[?25l"
	ShowCursor       Function = "\033[?25h"
)

type Color string

const (
	PlumForeground   Color = "\033[38;5;96m"
	BlueForeground Color = "\033[38;5;75m"
	GreyForeground   Color = "\033[38;5;243m"
)

func (color Color) Set(string string) string {
	return fmt.Sprintf("%s%s\033[0m", color, string)
}

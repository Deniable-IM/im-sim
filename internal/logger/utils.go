package logger

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 20
	}
	return width, height
}

func fitTerminal(str string) string {
	width, _ := getTerminalSize()
	if len(str) > width {
		str = str[:width-2]
	}
	return str
}

func StatusBar(progress int, finished int, barWidth int) string {
	status := (progress * barWidth) / finished
	return fmt.Sprintf("[%s>%s] %d/%d", strings.Repeat("=", status), strings.Repeat(" ", barWidth-status), progress, finished)
}

// Restore cursor on forced exit
func handleForcedExit() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Print(ShowCursor)
		os.Exit(1)
	}()
}

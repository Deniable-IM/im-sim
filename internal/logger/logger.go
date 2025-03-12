package logger

import (
	"encoding/json"
	"fmt"
	"io"
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

type imageBuildStream struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

type imagePullStream struct {
	Status         string `json:"status"`
	ProgressDetail string `json:"progressDetail"`
	Progress       string `json:"progress"`
	Id             string `json:"id"`
}

func LogImageBuild(reader io.Reader) {
	decoder := json.NewDecoder(reader)

	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)

	// Restore cursor on forced exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Print(ShowCursor)
		os.Exit(1)
	}()

	for {
		var msg imageBuildStream
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
		}

		stream := msg.Stream
		if stream != "" {
			if strings.HasPrefix(stream, "Step") {
				fmt.Print(ClearEntierLine)
				fmt.Print(PlumForeground.Set(fitTerminal(stream)))
			} else if strings.HasPrefix(stream, "Successfully built") {
			} else if strings.HasPrefix(stream, "Successfully tagged") {
				fmt.Print(ClearEntierLine)
				fmt.Print(BlueForeground.Set(fitTerminal(stream)))
				fmt.Print("\n\n")
			} else if strings.Contains(stream, "--->") {
				fmt.Print(MoveCursorDown)
				fmt.Print(ClearEntierLine)
				fmt.Printf(GreyForeground.Set(fitTerminal(stream)))
				fmt.Print(MoveCursorUp)
				fmt.Print(MoveCursorUp)
			} else {
				fmt.Print(GreyForeground.Set(fitTerminal(stream)))
			}
		}
	}
}

func LogImagePull(reader io.Reader) {
	decoder := json.NewDecoder(reader)

	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)

	// Restore cursor on forced exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Print(ShowCursor)
		os.Exit(1)
	}()

	var imageName string
	progressMap := make(map[string]string)

	for {
		var msg imagePullStream
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
		}

		if msg.Progress != "" {
			for range progressMap {
				fmt.Print(MoveCursorUp)
				fmt.Print(ClearEntierLine)
			}

			progressMap[msg.Id] = msg.Progress
			for key, value := range progressMap {
				fmt.Print(GreyForeground.Set(fmt.Sprintf("[%s] %s\n", key, value)))
			}
		}

		if strings.HasPrefix(msg.Status, "Pulling from") {
			imageName = strings.Split(msg.Status, " ")[2]
			fmt.Print(PlumForeground.Set(fmt.Sprintf("Pulling: %s\n", imageName)))
		}
	}

	fmt.Print(BlueForeground.Set(fmt.Sprintf("Pulled: %s\n", imageName)))
}

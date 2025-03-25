package logger

import (
	"bufio"
	"deniable-im/im-sim/internal/types"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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

type containerSliceStream struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
	Image  string `json:"image"`
	Name   string `json:"name"`
}

func LogImageBuild(reader io.Reader) {
	decoder := json.NewDecoder(reader)

	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)
	handleForcedExit()

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
				fmt.Print(ClearEntireLine)
				fmt.Print(PlumForeground.Set(fitTerminal(stream)))
			} else if strings.HasPrefix(stream, "Successfully built") {
			} else if strings.HasPrefix(stream, "Successfully tagged") {
				fmt.Print(ClearEntireLine)
				fmt.Print(BlueForeground.Set(fitTerminal(stream)))
				fmt.Print("\n\n")
			} else if strings.Contains(stream, "--->") {
				fmt.Print(MoveCursorDown)
				fmt.Print(ClearEntireLine)
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
	handleForcedExit()

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
				fmt.Print(ClearEntireLine)
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

func LogContainerSlice(reader io.Reader) {
	decoder := json.NewDecoder(reader)

	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)
	handleForcedExit()

	progress := 0
	for {
		var stream containerSliceStream
		if err := decoder.Decode(&stream); err != nil {
			if err == io.EOF {
				break
			}
		}

		if stream.Status == "created" {
			finished := stream.Total
			progress = progress + 1

			title := fmt.Sprintf("Creating %s containers ", stream.Image)
			statusBar := fmt.Sprintf("%s %s\r", title, StatusBar(progress, finished, 50))

			if progress < finished {
				fmt.Print(ClearEntireLine)
				fmt.Printf("%s\r", PlumForeground.Set(statusBar))
			} else {
				fmt.Print(ClearEntireLine)
				fmt.Printf("%s\n", BlueForeground.Set(statusBar))
			}
		}
	}
}

func LogStartContainers(reader io.Reader) {
	decoder := json.NewDecoder(reader)

	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)
	handleForcedExit()

	progress := 0
	for {
		var stream containerSliceStream
		if err := decoder.Decode(&stream); err != nil {
			if err == io.EOF {
				break
			}
		}

		if stream.Status == "started" {
			finished := stream.Total
			progress = progress + 1

			title := fmt.Sprintf("Starting %s containers ", stream.Image)
			statusBar := fmt.Sprintf("%s %s\r", title, StatusBar(progress, finished, 50))

			if progress < finished {
				fmt.Print(ClearEntireLine)
				fmt.Printf("%s\r", PlumForeground.Set(statusBar))
			} else {
				fmt.Print(ClearEntireLine)
				fmt.Printf("%s\n", BlueForeground.Set(statusBar))
			}
		}
	}
}

func LogContainerStarted(string string) {
	fmt.Printf("%s\n", BlueForeground.Set(string))
}

func LogContainerOptions(string string) {
	fmt.Printf("%s\n", GreyForeground.Set(string))
}

func LogNetworkNew(string string) {
	fmt.Printf("%s\n", GreyForeground.Set(string))
}

func LogNetworkConnect(string string) {
	fmt.Print("\n")
	fmt.Print(ClearEntireLine)
	fmt.Printf("%s\r", GreyForeground.Set(string))
	fmt.Print(MoveCursorUp)
}

func LogContainerExec(reader io.Reader, commands []string, containerName string) {
	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)
	handleForcedExit()

	cmd := strings.Join(commands, " ")
	scanner := bufio.NewScanner(reader)

	fmt.Print(ClearEntireLine)
	log := fmt.Sprintf("[*] %s:$ %s\n", containerName, cmd)
	fmt.Print(GreyForeground.Set(log))

	logSet := make(types.Set[string])
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			if err := logSet.Add(text); err != nil {
				break
			}
		}
		fmt.Print(ClearEntireLine)
		log := fmt.Sprintf("\t%s\n\r", text)
		fmt.Print(GreyForeground.Set(log))
	}
}

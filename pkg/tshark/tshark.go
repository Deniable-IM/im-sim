package tshark

import (
	"fmt"
	"os"
	"os/exec"
)

func RunTshark(networkInterface, dir string, duration int64) (*exec.Cmd, error) {
	filename := fmt.Sprintf("%v/capture.pcapng", dir)
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	file.Close() //intentional close as the file simply has to exist for tshark to use it.

	chmodErr := os.Chmod(filename, 0666)
	if chmodErr != nil {
		fmt.Println("Error setting permissions:", chmodErr)
		return nil, chmodErr
	}

	cmd := exec.Command("tshark", "-i", networkInterface, "-a", fmt.Sprintf("duration:%v", duration), "-F", "pcapng", "-w", filename)

	cmdErr := cmd.Start()
	if cmdErr != nil {
		fmt.Printf("Encountered error while starting tshark: %v", cmdErr)
		return nil, cmdErr
	}

	return cmd, nil
}

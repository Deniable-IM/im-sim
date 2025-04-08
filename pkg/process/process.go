package process

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type Process struct {
	conn net.Conn
	// Reader   *bufio.Reader
	Buffer   *bytes.Buffer
	kill_sig chan int
}

func NewProcess(conn net.Conn, reader *bytes.Buffer) *Process {
	return &Process{conn, reader, make(chan int)}
}

func (process *Process) Cmd(cmd []byte) error {
	_, err := process.conn.Write(cmd)
	if err != nil {
		return fmt.Errorf("Process failed to write cmd: %w.", err)
	}
	return nil
}

func (process *Process) Close() error {
	if process.conn != nil {
		process.kill_sig <- 1
		return process.conn.Close()
	}
	return nil
}

func (process *Process) ProcessReader() {

	go func() {
		for len(process.kill_sig) == 0 {
			time.Sleep(1 * time.Second)
			process.Cmd([]byte("read\n"))
		}
	}()
}

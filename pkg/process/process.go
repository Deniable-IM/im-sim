package process

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type Process struct {
	conn   net.Conn
	Buffer *bytes.Buffer
	Mu     sync.Mutex
}

func NewProcess(conn net.Conn, reader *bytes.Buffer) *Process {
	return &Process{conn, reader, sync.Mutex{}}
}

func (process *Process) Cmd(cmd []byte) error {
	process.Mu.Lock()
	defer process.Mu.Unlock()

	_, err := process.conn.Write(cmd)
	if err != nil {
		return fmt.Errorf("Process failed to write cmd: %w.", err)
	}
	return nil
}

func (process *Process) Read(delim byte) []string {
	process.Mu.Lock()
	defer process.Mu.Unlock()

	lines := []string{}
	for process.Buffer.Len() != 0 {
		line, err := process.Buffer.ReadString(delim)
		if len(line) > 1 {
			lines = append(lines, line)
		}

		// Read until EOF
		if err != nil {
			break
		}
	}
	return lines
}

func (process *Process) Close() error {
	if process.conn != nil {
		return process.conn.Close()
	}
	return nil
}

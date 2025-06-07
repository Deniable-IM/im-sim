package process

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
)

var (
	processSem chan struct{} = make(chan struct{}, runtime.NumCPU())
)

type Process struct {
	conn     net.Conn
	Buffer   *bytes.Buffer
	commands []string
	execFunc func([]string, bool) (*Process, error)
	mu       sync.Mutex
}

func NewProcess(conn net.Conn, reader *bytes.Buffer, commands []string, execFunc func([]string, bool) (*Process, error)) *Process {
	return &Process{conn, reader, commands, execFunc, sync.Mutex{}}
}

func (process *Process) Cmd(cmd []byte) error {
	processSem <- struct{}{}
	defer func() { <-processSem }()

	process.mu.Lock()
	defer process.mu.Unlock()

	_, err := process.conn.Write(cmd)
	if err != nil {
		err := process.retry(cmd)
		if err != nil {
			return fmt.Errorf("Failed to retry process: %w.", err)
		}
	}
	return nil
}

func (process *Process) Read(delim byte) []string {
	processSem <- struct{}{}
	defer func() { <-processSem }()

	process.mu.Lock()
	defer process.mu.Unlock()

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

func (process *Process) retry(cmd []byte) error {
	process.Close()

	newProcess, err := process.execFunc(process.commands, true)
	if err != nil {
		return fmt.Errorf("Failed to create new process: %w.", err)
	}
	log.Printf("New process created: %v.\n", process.commands)

	process.conn = newProcess.conn
	process.Buffer = newProcess.Buffer
	process.commands = newProcess.commands
	process.execFunc = newProcess.execFunc

	_, err = process.conn.Write(cmd)
	if err != nil {
		return fmt.Errorf("New process failed to write: %w.", err)
	}
	log.Printf("New process write.")

	return nil
}

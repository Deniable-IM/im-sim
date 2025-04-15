package process

import (
	"bytes"
	"fmt"
	"net"
)

type Process struct {
	conn   net.Conn
	Buffer *bytes.Buffer
}

func NewProcess(conn net.Conn, reader *bytes.Buffer) *Process {
	return &Process{conn, reader}
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
		return process.conn.Close()
	}
	return nil
}

// radare - LGPL - Copyright 2015 - nibble

/*
Package r2pipe allows to call r2 commands from Go. A simple hello world would
look like the following snippet:

	package main

	import (
		"fmt"

		"github.com/kr4phy/r2pipe-go"
	)

	func main() {
		r2p, err := r2pipe.NewPipe("malloc://256")
		if err != nil {
			panic(err)
		}
		defer r2p.Close()

		_, err = r2p.Cmd("w Hello World")
		if err != nil {
			panic(err)
		}
		buf, err := r2p.Cmd("ps")
		if err != nil {
			panic(err)
		}
		fmt.Println(buf)
	}
*/
package r2pipe

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

// Pipe represents a communication interface with r2 that will be used to
// execute commands and obtain their results.
type Pipe struct {
	File   string
	r2cmd  *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	Core   unsafe.Pointer
	cmd    CmdDelegate
	close  CloseDelegate
}

// CmdDelegate is a function type for executing commands.
type CmdDelegate func(*Pipe, string) (string, error)

// CloseDelegate is a function type for closing the pipe.
type CloseDelegate func(*Pipe) error

// EventDelegate is a function type for handling events.
type EventDelegate func(*Pipe, string, any, string) bool

// NewPipe returns a new r2 pipe and initializes an r2 core that will try to
// load the provided file or URI. If file is an empty string, the env vars
// R2PIPE_{IN,OUT} will be used as file descriptors for input and output, this
// is the case when r2pipe is called within r2.
func NewPipe(file string) (*Pipe, error) {
	if file == "" {
		return newPipeFd()
	}
	return newPipeCmd(file)
}

func newPipeFd() (*Pipe, error) {
	r2pipeIn := os.Getenv("R2PIPE_IN")
	r2pipeOut := os.Getenv("R2PIPE_OUT")

	if r2pipeIn == "" || r2pipeOut == "" {
		return nil, fmt.Errorf("missing R2PIPE_{IN,OUT} vars")
	}

	r2pipeInFd, err := strconv.Atoi(r2pipeIn)
	if err != nil {
		return nil, fmt.Errorf("failed to convert IN into file descriptor: %w", err)
	}

	r2pipeOutFd, err := strconv.Atoi(r2pipeOut)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OUT into file descriptor: %w", err)
	}

	stdout := os.NewFile(uintptr(r2pipeInFd), "R2PIPE_IN")
	stdin := os.NewFile(uintptr(r2pipeOutFd), "R2PIPE_OUT")

	r2p := &Pipe{
		stdin:  stdin,
		stdout: stdout,
	}

	return r2p, nil
}

func newPipeCmd(file string) (*Pipe, error) {
	r2p := &Pipe{
		File:  file,
		r2cmd: exec.Command("radare2", "-q0", file),
	}

	var err error
	r2p.stdin, err = r2p.r2cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	r2p.stdout, err = r2p.r2cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	r2p.stderr, err = r2p.r2cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err = r2p.r2cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start radare2: %w", err)
	}

	// Read the initial data
	if _, err = bufio.NewReader(r2p.stdout).ReadString('\x00'); err != nil {
		return nil, fmt.Errorf("failed to read initial data: %w", err)
	}

	return r2p, nil
}

// Write implements the standard Write interface: it writes data to the r2
// pipe, blocking until r2 have consumed all the data.
func (r2p *Pipe) Write(p []byte) (n int, err error) {
	return r2p.stdin.Write(p)
}

// Read implements the standard Read interface: it reads data from the r2
// pipe's stdin, blocking until the previously issued commands have finished.
func (r2p *Pipe) Read(p []byte) (n int, err error) {
	return r2p.stdout.Read(p)
}

// ReadErr reads from the stderr pipe.
func (r2p *Pipe) ReadErr(p []byte) (n int, err error) {
	return r2p.stderr.Read(p)
}

// On registers an event callback for stderr messages.
func (r2p *Pipe) On(evname string, p any, cb EventDelegate) error {
	path, err := r2p.Cmd("===stderr")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	go func() {
		defer f.Close()
		var buf bytes.Buffer
		for {
			if _, err := io.Copy(&buf, f); err != nil {
				break
			}
			if buf.Len() > 0 {
				if !cb(r2p, evname, p, buf.String()) {
					break
				}
			}
		}
	}()
	return nil
}

// Cmd is a helper that allows to run r2 commands and receive their output.
func (r2p *Pipe) Cmd(cmd string) (string, error) {
	if r2p.Core != nil {
		if r2p.cmd != nil {
			return r2p.cmd(r2p, cmd)
		}
		return "", nil
	}

	if _, err := fmt.Fprintln(r2p, cmd); err != nil {
		return "", err
	}

	buf, err := bufio.NewReader(r2p).ReadString('\x00')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(buf, "\n\x00"), nil
}

// Cmdf is like Cmd but formats the command.
func (r2p *Pipe) Cmdf(f string, args ...any) (string, error) {
	return r2p.Cmd(fmt.Sprintf(f, args...))
}

// Cmdj acts like Cmd but interprets the output of the command as json. It
// returns the parsed json keys and values.
func (r2p *Pipe) Cmdj(cmd string) (out any, err error) {
	rstr, err := r2p.Cmd(cmd)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(rstr), &out)
	return out, err
}

// CmdjStruct acts like Cmdj but it will fill the provided struct with the wanted values.
// It returns the command execution error.
func (r2p *Pipe) CmdjStruct(cmd string, out any) error {
	rstr, err := r2p.Cmd(cmd)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(rstr), out)
}

// Cmdjf is like Cmdj but formats the command.
func (r2p *Pipe) Cmdjf(f string, args ...any) (any, error) {
	return r2p.Cmdj(fmt.Sprintf(f, args...))
}

// CmdjfStruct is like CmdjStruct, but also formats the command.
func (r2p *Pipe) CmdjfStruct(f string, out any, args ...any) error {
	return r2p.CmdjStruct(fmt.Sprintf(f, args...), out)
}

// Close shuts down r2, closing the created pipe.
func (r2p *Pipe) Close() error {
	if r2p.close != nil {
		return r2p.close(r2p)
	}

	if r2p.File == "" {
		return nil
	}

	if _, err := r2p.Cmd("q"); err != nil {
		return err
	}

	return r2p.r2cmd.Wait()
}

// ForceClose forces shutdown of r2, closing the created pipe.
func (r2p *Pipe) ForceClose() error {
	if r2p.close != nil {
		return r2p.close(r2p)
	}

	if r2p.File == "" {
		return nil
	}

	if _, err := r2p.Cmd("q!"); err != nil {
		return err
	}

	return r2p.r2cmd.Wait()
}

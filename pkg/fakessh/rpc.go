// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"context"
	"errors"
	"sync"

	"github.com/hashicorp/packer/packer"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/ctxio"
)

// RPC command pipe handles
type RpcState struct {
	Stdin  deadlineReaderCloser
	Stdout deadlineWriterCloser
	Stderr deadlineWriterCloser
}

// Exported RPC object
//
// If Comm is nil, do nothing
type RpcSsh struct {
	M    map[RpcCmd]RpcState
	Comm packer.Communicator
	L sync.RWMutex
}

// RPC argument
type RpcCmd struct {
	Cmd        string
	StdinPipe  string
	StdoutPipe string
	StderrPipe string
}

// Open pipes in preparation for command
func (ssh *RpcSsh) OpenPipes(ctx context.Context, c *RpcCmd, exitCode *int) error {
	inpipe, err := openReadPipe(ctx, c.StdinPipe)
	if err != nil {
		return err
	}
	outpipe, err := openWritePipe(ctx, c.StdoutPipe)
	if err != nil {
		return err
	}
	errpipe, err := openWritePipe(ctx, c.StderrPipe)
	if err != nil {
		return err
	}
	ssh.L.Lock()
	defer ssh.L.Unlock()
	ssh.M[*c] = RpcState{
		Stdin:  inpipe,
		Stdout: outpipe,
		Stderr: errpipe,
	}
	return nil
}

// Run command c on communicator and return exitcode
func (ssh *RpcSsh) Run(ctx context.Context, c *RpcCmd, exitCode *int) error {
	var err error = nil

	ssh.L.RLock()
	pipes, ok := ssh.M[*c]
	ssh.L.RUnlock()

	if !ok {
		return errors.New("RpcSsh.Run: call RpcSsh.OpenPipes before calling Run")
	}
	defer func() {
		ssh.L.Lock()
		defer ssh.L.Unlock()
		delete(ssh.M, *c)
	}()

	cmd := &packer.RemoteCmd{
		Command: c.Cmd,
		Stdin:   ctxio.ReaderAdapter(ctx, pipes.Stdin),
		Stdout:  ctxio.WriterAdapter(ctx, pipes.Stdout),
		Stderr:  ctxio.WriterAdapter(ctx, pipes.Stderr),
	}
	defer pipes.Stdin.Close()
	defer pipes.Stdout.Close()
	defer pipes.Stderr.Close()

	err = ssh.Comm.Start(ctx, cmd)
	if err != nil {
		return err
	}

	*exitCode = cmd.Wait()
	return nil
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build windows

package localcommunicator

import (
	"context"
	"os/exec"
	"strconv"

	"github.com/hashicorp/packer/packer"
)

// Run command by passing it directly as an argument to powershell
func (c *comm) Start(ctx context.Context, cmd *packer.RemoteCmd) (err error) {
	lcmd := exec.Command("cmd.exe", "/c", cmd.Command)
	lcmd.Stdin = cmd.Stdin
	lcmd.Stdout = cmd.Stdout
	lcmd.Stderr = cmd.Stderr
	exitChan := make(chan struct{})
	go func() {
		exitCode := RunExitCode(lcmd)
		if ctx.Err() != nil {
			exitCode = EXIT_FAILURE
		}
		cmd.SetExited(exitCode)
		close(exitChan)
	}()
	// exec.CommandContext did not seem to kill the command correctly
	go func() {
		select {
		case <-exitChan:
		case <-ctx.Done():
			exec.Command("TASKKILL", "/T", "/F",
				"/PID", strconv.Itoa(lcmd.Process.Pid),
			).Run()
		}
	}()
	return nil
}

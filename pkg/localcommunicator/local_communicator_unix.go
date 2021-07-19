// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build darwin linux

package localcommunicator

import (
	"context"
	"os/exec"

	"github.com/hashicorp/packer/packer"
)

// Run command by passing it directly as an argument to /bin/sh
func (c *comm) Start(ctx context.Context, cmd *packer.RemoteCmd) (err error) {
	lcmd := exec.CommandContext(ctx, "/bin/sh", "-c", cmd.Command)
	lcmd.Stdin = cmd.Stdin
	lcmd.Stdout = cmd.Stdout
	lcmd.Stderr = cmd.Stderr
	go func() {
		exitCode := RunExitCode(lcmd)
		cmd.SetExited(exitCode)
	}()
	return nil
}

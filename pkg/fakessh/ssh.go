// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"context"
	"os"
	"os/signal"
)

// The fake ssh command
func Ssh() int {
	ctx := context.Background()
	dctx, cancel := context.WithCancel(ctx)
	defer func() {
		select {
		case <-dctx.Done():
		default:
			cancel()
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	go func() {
		select {
		case <-signalChan:
			cancel()
		case <-dctx.Done():
		}
		return
	}()

	rpcDir, envSet := os.LookupEnv(RPCDirEnvVarName)
	if !envSet {
		return EXIT_FAILURE
	}

	sshCmd := ArgvToSh(ParseCmd(os.Args))

	if len(sshCmd) == 0 {
		return EXIT_FAILURE
	}

	cmd := &Cmd{
		Command: sshCmd,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	exitCode, err := RunCmd(dctx, rpcDir, cmd)
	if err != nil {
		return EXIT_FAILURE
	} else {
		return exitCode
	}
}

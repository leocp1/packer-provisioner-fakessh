// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package localcommunicator_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/localcommunicator"
)

const (
	MAXTESTTIME = time.Duration(10) * time.Second
)

func TestLocalComm(t *testing.T) {
	lc, _ := localcommunicator.New()
	ctx := context.Background()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			stdin := bytes.NewBufferString(tt.stdin)
			cmd := &packer.RemoteCmd{
				Command: tt.cmd,
				Stdin:   stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			}
			dctx, cancel := context.WithTimeout(ctx, MAXTESTTIME)
			defer cancel()
			lc.Start(dctx, cmd)
			exitCode := cmd.Wait()
			sstdout := stdout.String()
			sstderr := stderr.String()
			var ok bool = true
			ok = ok && stdout.String() == tt.stdout
			ok = ok && sstderr == tt.stderr
			exitCode = (exitCode%256 + 256) % 256
			ok = ok && exitCode == (tt.exitCode%256+256)%256
			if !ok {
				actual := struct {
					stdout   string
					stderr   string
					exitCode int
				}{
					stdout:   sstdout,
					stderr:   sstderr,
					exitCode: exitCode,
				}
				t.Errorf("failed for %#v ... (actual: %#v)", tt, actual)
			}
		})
	}
}

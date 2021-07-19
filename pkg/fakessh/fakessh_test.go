// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/yookoala/realpath"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/fakessh"
	"github.com/leocp1/packer-provisioner-fakessh/pkg/localcommunicator"
)

const (
	MAXTESTTIME = time.Duration(10) * time.Second
)

// bytes.Buffer but with a dummy Close function
// so it can become a deadlineReader/Writer Closer
type drwcBuffer struct {
	B *bytes.Buffer
}

func (b *drwcBuffer) Close() error {
	return nil
}
func (b *drwcBuffer) Write(p []byte) (n int, err error) {
	return b.B.Write(p)
}
func (b *drwcBuffer) SetWriteDeadline(t time.Time) error {
	return nil
}
func (b *drwcBuffer) Read(p []byte) (n int, err error) {
	return b.B.Read(p)
}
func (b *drwcBuffer) SetReadDeadline(t time.Time) error {
	return nil
}

// testing.T logging as a io.Writer
type tWriter struct {
	t *testing.T
}

func (tw *tWriter) Write(p []byte) (n int, err error) {
	tw.t.Logf("%s", p)
	return len(p), nil
}

func TestServer(t *testing.T) {
	var err error = nil

	ctx := context.Background()

	comm, err := localcommunicator.New()
	if err != nil {
		t.Fatal(err)
	}

	srv, err := fakessh.NewServer(comm, "")
	if err != nil {
		t.Fatal(err)
	}

	srvChan := make(chan error)
	go func() {
		srvChan <- srv.Serve()
	}()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			stdout := &drwcBuffer{&bytes.Buffer{}}
			stderr := &drwcBuffer{&bytes.Buffer{}}
			stdin := &drwcBuffer{bytes.NewBufferString(tt.stdin)}
			cmd := &fakessh.Cmd{
				Command: tt.cmd,
				Stdin:   stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			dctx, cancel := context.WithTimeout(ctx, MAXTESTTIME)
			defer cancel()

			exitCode, err := fakessh.RunCmd(dctx, srv.Dir, cmd)
			if err != nil && !os.IsTimeout(err) {
				t.Error(err)
			}

			var ok bool = true
			ok = ok && stdout.B.String() == tt.stdout
			ok = ok && stderr.B.String() == tt.stderr
			exitCode = (256 + exitCode%256) % 256
			ok = ok && exitCode == (tt.exitCode%256+256)%256
			if !ok {
				actual := struct {
					stdout   string
					stderr   string
					exitCode int
				}{
					stdout:   stdout.B.String(),
					stderr:   stderr.B.String(),
					exitCode: exitCode,
				}
				t.Errorf("failed for %#v ... (actual: %#v)", tt, actual)
			}
		})
	}

	srv.Shutdown(ctx)
	err = <-srvChan
	if err != http.ErrServerClosed {
		t.Error(err)
	}
}

func TestFakessh(t *testing.T) {
	var err error = nil
	var ok bool = false

	ctx := context.Background()

	comm, err := localcommunicator.New()
	if err != nil {
		t.Fatal(err)
	}

	srv, err := fakessh.NewServer(comm, "")
	if err != nil {
		t.Fatal(err)
	}

	srvChan := make(chan error)
	go func() {
		srvChan <- srv.Serve()
	}()

	sshExeDir, ok := fakessh.FakeSshPath()
	if !ok {
		t.Log("ssh not found, building...")
		sshExeDir, err = fakessh.GoBuildFakeSsh(ctx)
		defer os.RemoveAll(sshExeDir)
		if err != nil {
			t.Skip("ssh executable not found or buildable")
		}
	}
	t.Logf("Using ssh path:%s", sshExeDir)

	sshExe := filepath.Join(sshExeDir, fakessh.SSHEXENAME)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			stdin := bytes.NewBufferString(tt.stdin)

			dctx, cancel := context.WithTimeout(ctx, MAXTESTTIME)
			defer cancel()

			cmd := exec.CommandContext(dctx, sshExe, "user@host", tt.cmd)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			cmd.Stdin = stdin
			cmd.Env, err = fakessh.AddFakeSshPath(cmd.Env, sshExeDir, srv.Dir)
			if err != nil {
				t.Fatal(err)
			}
			exitCode := localcommunicator.RunExitCode(cmd)
			if dctx.Err() != nil {
				exitCode = localcommunicator.EXIT_FAILURE
			}

			ok = true
			ok = ok && stdout.String() == tt.stdout
			ok = ok && stderr.String() == tt.stderr
			exitCode = (256 + exitCode%256) % 256
			ok = ok && exitCode == (tt.exitCode%256+256)%256
			if !ok {
				actual := struct {
					stdout   string
					stderr   string
					exitCode int
				}{
					stdout:   stdout.String(),
					stderr:   stderr.String(),
					exitCode: exitCode,
				}
				t.Errorf("failed for %#v ... (actual: %#v)", tt, actual)
			}
		})
	}

	srv.Shutdown(ctx)
	err = <-srvChan
	if err != http.ErrServerClosed {
		t.Error(err)
	}
}

// Copy Nix to itself on the localhost
func TestNixCopyClosure(t *testing.T) {
	ctx := context.Background()

	comm, err := localcommunicator.New()
	if err != nil {
		t.Fatal(err)
	}
	srv, err := fakessh.NewServer(comm, "")
	if err != nil {
		t.Fatal(err)
	}
	srvChan := make(chan error)
	go func() {
		srvChan <- srv.Serve()
	}()

	sshExeDir, ok := fakessh.FakeSshPath()
	if !ok {
		sshExeDir, err = fakessh.GoBuildFakeSsh(ctx)
		defer os.RemoveAll(sshExeDir)
		if err != nil {
			t.Skip("ssh executable not found or buildable")
		}
	}
	t.Logf("Using ssh path:%s", sshExeDir)

	nix, err := exec.LookPath("nix")
	if err != nil {
		t.Skip("nix not found")
	}
	// NOTE: nix-copy-closure and friends are all symlinks to the nix binary
	// nix determines which command to run using argv[0]
	nix, err = realpath.Realpath(nix)
	if err != nil {
		t.Fatal("finding absolute path of nix failed")
	}
	t.Logf("Using nix path: %s", nix)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx,
		nix, "copy", "--to", "ssh://user@localhost", nix)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env, err = fakessh.AddFakeSshPath(cmd.Env, sshExeDir, srv.Dir)
	if err != nil {
		t.Fatal("AddFakeSshPath failed")
	}
	t.Logf("cmd.Env: %#v", cmd.Env)
	exitCode := localcommunicator.RunExitCode(cmd)

	ok = true
	ok = ok && stdout.String() == ""
	ok = ok && stderr.String() == ""
	exitCode = (256 + exitCode%256) % 256
	ok = ok && exitCode == 0
	if !ok {
		actual := struct {
			stdout   string
			stderr   string
			exitCode int
		}{
			stdout:   stdout.String(),
			stderr:   stderr.String(),
			exitCode: exitCode,
		}
		t.Logf("Unexpected output %#v)", actual)
	}
	if exitCode != 0 {
		t.Errorf("nix copy failed")
	}

	srv.Shutdown(ctx)
	err = <-srvChan
	if err != http.ErrServerClosed {
		t.Error(err)
	}
}

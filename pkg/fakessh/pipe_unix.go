// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Requires POSIX Mkfifo
// +build darwin linux

package fakessh

import (
	"context"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	// The permissions for the named pipes carrying stdin/stdout/stderr
	PIPEPERM = 0600
)

func makePipe(name string) (pipe, error) {
	pname := strings.Join([]string{
		filepath.Join(os.TempDir(), "packer-provisioner-fakessh"),
		strconv.Itoa(os.Getpid()),
		name,
	}, "-")
	err := unix.Mkfifo(pname, PIPEPERM)
	return pipe{Dir: pname}, err
}

func (p *pipe) Close() error {
	return os.RemoveAll(p.Dir)
}

func openReadPipe(ctx context.Context, dir string,
) (deadlineReaderCloser, error) {
	return os.OpenFile(dir, os.O_RDONLY|syscall.O_NONBLOCK, PIPEPERM)
}

func openWritePipe(ctx context.Context, dir string,
) (deadlineWriterCloser, error) {
	return os.OpenFile(dir, os.O_WRONLY|syscall.O_NONBLOCK, PIPEPERM)
}

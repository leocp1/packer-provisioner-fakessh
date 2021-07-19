// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build windows

package fakessh

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	winio "github.com/Microsoft/go-winio"
)

// Fake a nonblocking write side open by forwarding the arguments to the
// accepted connection.
// If SetReadDeadline is called before the connection is accepted,
// all future Read and SetReadDeadlines will immediately fail
type drc struct {
	l           net.Listener
	cond        *sync.Cond
	dlBeforeAcc bool
	accepted    bool
	c           net.Conn
}

func newDrc(l net.Listener) (*drc, error) {
	lck := sync.Mutex{}
	d := &drc{
		cond:        sync.NewCond(&lck),
		accepted:    false,
		dlBeforeAcc: false,
		l:           l,
	}
	go func() {
		d.c, _ = l.Accept()
		d.cond.L.Lock()
		d.accepted = true
		d.cond.L.Unlock()
		d.cond.Broadcast()
	}()
	return d, nil
}

var errDLBeforeAcc error = errors.New(
	"SetReadDeadline before connection accepted",
)

func (d *drc) SetReadDeadline(t time.Time) error {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()
	if d.accepted {
		return d.c.SetReadDeadline(t)
	}
	d.dlBeforeAcc = true
	d.cond.Broadcast()
	return errDLBeforeAcc
}
func (d *drc) Read(b []byte) (int, error) {
	d.cond.L.Lock()
	for !d.accepted && !d.dlBeforeAcc {
		d.cond.Wait()
	}
	if d.dlBeforeAcc {
		d.cond.L.Unlock()
		return 0, errDLBeforeAcc
	} else {
		d.cond.L.Unlock()
		return d.c.Read(b)
	}
}
func (d *drc) Close() error {
	lerr := d.l.Close()
	derr := d.c.Close()
	if lerr != nil {
		return lerr
	}
	return derr
}

func makePipe(name string) (pipe, error) {
	pname := strings.Join([]string{
		`\\.\pipe\packer-provisioner-fakessh`,
		strconv.Itoa(os.Getpid()),
		name,
	}, "-")
	return pipe{Dir: pname}, nil
}

func (p *pipe) Close() error {
	return nil
}

func openReadPipe(ctx context.Context, dir string,
) (deadlineReaderCloser, error) {
	l, err := winio.ListenPipe(dir, nil)
	if err != nil {
		return nil, err
	}
	return newDrc(l)
}

func openWritePipe(ctx context.Context, dir string,
) (deadlineWriterCloser, error) {
	return winio.DialPipeContext(ctx, dir)
}

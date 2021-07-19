// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ctxio_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/ctxio"
)

const (
	MAXCANCELTESTTIME = time.Duration(10) * time.Second
)

// Replacement for os.ErrDeadlineExceeded
var (
	ERRDE = errors.New("Deadline exceeded")
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

type copyResult struct {
	dst string
	src string
	n   int64
	err error
}

// Test that without deadlines, ReadCtx and WriteCtx produce the same results as
// normal Read and Write
func TestCopy(t *testing.T) {
	ctx := context.Background()

	tstr := "copied successfully"

	dst := &bytes.Buffer{}
	src := bytes.NewBufferString(tstr)
	n, err := io.Copy(dst, src)
	expected := copyResult{
		dst: dst.String(),
		src: src.String(),
		n:   n,
		err: err,
	}

	cdst := &drwcBuffer{&bytes.Buffer{}}
	csrc := &drwcBuffer{bytes.NewBufferString(tstr)}
	cn, cerr := io.Copy(
		WriterAdapter(ctx, cdst),
		ReaderAdapter(ctx, csrc))
	got := copyResult{
		dst: cdst.B.String(),
		src: csrc.B.String(),
		n:   cn,
		err: cerr,
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("copy = %#v ; context copy = %#v", expected, got)
	}
}

// A reader/writer that sleeps until the deadline
// For simplicity, read/write are considered the same operation
// WARNING: be careful with time.Timer:
// https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers/
type slowReaderWriter struct {
	// broadcast to all readers/writers that the timer expired
	cond *sync.Cond
	// new read/write deadline
	dLchan chan time.Time
	// timer expired
	timerExpired bool
	// timer channel drained
	timerDrained bool
	// closed when Close() called
	closechan chan struct{}
	// has this been closed
	closed bool
}

var errClosedSRW = errors.New(
	"slowReaderWriter: read/write on closed slowReaderWriter",
)

func NewSlowReaderWriter() *slowReaderWriter {
	l := sync.Mutex{}
	srw := &slowReaderWriter{
		cond:         sync.NewCond(&l),
		dLchan:       make(chan time.Time),
		timerExpired: false,
		timerDrained: true,
		closechan:    make(chan struct{}),
		closed:       false,
	}

	timer := time.NewTimer(time.Duration(0))
	if !timer.Stop() {
		<-timer.C
	}
	go func() {
		var t time.Time
		for {
			select {
			case t = <-srw.dLchan:
				srw.cond.L.Lock()
				srw.timerExpired = false
				if !srw.timerDrained {
					if !timer.Stop() {
						<-timer.C
					}
					srw.timerDrained = true
				}
				if !t.Equal(time.Time{}) {
					timer.Reset(time.Until(t))
					srw.timerDrained = false
				}
				srw.cond.L.Unlock()
			case <-timer.C:
				srw.cond.L.Lock()
				srw.timerDrained = true
				srw.timerExpired = true
				srw.cond.L.Unlock()
				srw.cond.Broadcast()
			case <-srw.closechan:
				srw.cond.L.Lock()
				srw.closed = true
				if !srw.timerDrained {
					if !timer.Stop() {
						<-timer.C
					}
				}
				srw.cond.L.Unlock()
				return
			}
		}
	}()
	return srw
}
func (srw *slowReaderWriter) Close() error {
	close(srw.closechan)
	return nil
}
func (srw *slowReaderWriter) Read(p []byte) (n int, err error) {
	srw.cond.L.Lock()
	defer srw.cond.L.Unlock()
	if srw.closed {
		return 0, errClosedSRW
	}
	for !srw.timerExpired {
		srw.cond.Wait()
	}
	return 0, ERRDE
}
func (srw *slowReaderWriter) SetReadDeadline(t time.Time) error {
	if srw.closed {
		return errClosedSRW
	}
	srw.dLchan <- t
	return nil
}
func (srw *slowReaderWriter) Write(p []byte) (n int, err error) {
	return srw.Read(p)
}
func (srw *slowReaderWriter) SetWriteDeadline(t time.Time) error {
	return srw.SetReadDeadline(t)
}

type ctxTest struct {
	name string
	ctx  context.Context
}

func makeCtxTests(ctx context.Context) ([]ctxTest, func()) {
	mstctx, mstcancel := context.WithTimeout(ctx, time.Millisecond)

	cctx, ccancel := context.WithCancel(ctx)
	ccancel()

	csctx, cscancel := context.WithCancel(ctx)
	go func() {
		timer := time.NewTimer(time.Duration(time.Second))
		<-timer.C
		cscancel()
	}()

	return []ctxTest{
			ctxTest{
				name: "millisecond timeout",
				ctx:  mstctx,
			},
			ctxTest{
				name: "cancelled",
				ctx:  cctx,
			},
			ctxTest{
				name: "cancelled in a second",
				ctx:  csctx,
			},
		},
		func() {
			mstcancel()
		}
}

func testCancelRead(t *testing.T, ctx context.Context) {
	timer := time.NewTimer(MAXCANCELTESTTIME)
	srw := NewSlowReaderWriter()
	asrw := ReaderAdapter(ctx, srw)

	donec := make(chan error)
	go func() {
		asrw.Read(nil)
		close(donec)
	}()
	select {
	case <-donec:
	case <-timer.C:
		t.Errorf("Reader not cancelled in time")
	}
	srw.Close()
}

// Test reads cancellable
func TestCancelRead(t *testing.T) {
	ctxs, cancel := makeCtxTests(context.Background())
	defer cancel()
	for i, tt := range ctxs {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			testCancelRead(t, tt.ctx)
		})
	}
}

func testCancelWrite(t *testing.T, ctx context.Context) {
	timer := time.NewTimer(MAXCANCELTESTTIME)
	srw := NewSlowReaderWriter()
	asrw := WriterAdapter(ctx, srw)

	donec := make(chan error)
	go func() {
		asrw.Write(nil)
		close(donec)
	}()
	select {
	case <-donec:
	case <-timer.C:
		t.Errorf("Reader not cancelled in time")
	}
	srw.Close()
}

// Test writes cancellable
func TestCancelWrite(t *testing.T) {
	ctxs, cancel := makeCtxTests(context.Background())
	defer cancel()
	for i, tt := range ctxs {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			testCancelWrite(t, tt.ctx)
		})
	}
}

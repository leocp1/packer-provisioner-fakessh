// Context cancellable io.
//
// Essentially a copy of "Canceling I/O in Go Cap’n Proto" by Ross Light [1]
// (licensed under CC BY 4.0 [2]) with an interface similar to "Context-aware
// io.Reader for Go" by David Hernandez and Mat Ryer [3].
//	[1] https://medium.com/@zombiezen/canceling-i-o-in-go-capn-proto-5ae8c09c5b29
//	[2] https://creativecommons.org/licenses/by/4.0/
//	[3] https://pace.dev/blog/2020/02/03/context-aware-ioreader-for-golang-by-mat-ryer.html
package ctxio

import (
	"context"
	"io"
	"time"
)

// A Writer with SetWriteDeadline
type DeadlineWriter interface {
	Write(b []byte) (n int, err error)
	SetWriteDeadline(t time.Time) error
}

// Writes bytes to a writer while attempting to respect ctx
func WriteCtx(ctx context.Context, w DeadlineWriter, b []byte) (int, error) {
	// Check for early cancel.
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	// Start separate goroutine to wait on Context.Done.
	if d, ok := ctx.Deadline(); ok {
		w.SetWriteDeadline(d)
	} else {
		w.SetWriteDeadline(time.Time{})
	}
	writeDone := make(chan struct{})
	listenDone := make(chan struct{})
	go func() {
		defer close(listenDone)
		select {
		case <-ctx.Done():
			w.SetWriteDeadline(time.Now()) // interrupt write
		case <-writeDone:
		}
	}()
	n, err := w.Write(b)
	close(writeDone)
	<-listenDone
	return n, err
}

type ctxwriter struct {
	w DeadlineWriter
	c context.Context
}

func (cw *ctxwriter) Write(b []byte) (int, error) {
	return WriteCtx(cw.c, cw.w, b)
}

// Create a Writer that attempts to respect ctx
func WriterAdapter(ctx context.Context, w DeadlineWriter) io.Writer {
	return &ctxwriter{
		w: w,
		c: ctx,
	}
}

// A Reader with SetReadDeadline
type DeadlineReader interface {
	Read(p []byte) (n int, err error)
	SetReadDeadline(t time.Time) error
}

// Reads bytes from a reader while attempting to respect ctx
func ReadCtx(ctx context.Context, r DeadlineReader, p []byte) (int, error) {
	// Check for early cancel.
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	// Start separate goroutine to wait on Context.Done.
	if d, ok := ctx.Deadline(); ok {
		r.SetReadDeadline(d)
	} else {
		r.SetReadDeadline(time.Time{})
	}
	readDone := make(chan struct{})
	listenDone := make(chan struct{})
	go func() {
		defer close(listenDone)
		select {
		case <-ctx.Done():
			r.SetReadDeadline(time.Now()) // interrupt read
		case <-readDone:
		}
	}()
	n, err := r.Read(p)
	close(readDone)
	<-listenDone
	return n, err
}

type ctxreader struct {
	r DeadlineReader
	c context.Context
}

func (cr *ctxreader) Read(b []byte) (int, error) {
	return ReadCtx(cr.c, cr.r, b)
}

// Create a Reader that attempts to respect ctx
func ReaderAdapter(ctx context.Context, r DeadlineReader) io.Reader {
	return &ctxreader{
		r: r,
		c: ctx,
	}
}

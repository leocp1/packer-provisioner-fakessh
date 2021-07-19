// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Requires /dev/zero
// +build linux

package ctxio_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/ctxio"
)

type LimitedDeadlineReader struct {
	R DeadlineReader
	N int64
	LR io.Reader
}
func (l *LimitedDeadlineReader) Limit() {
	l.LR = io.LimitReader(l.R, l.N)
}

func (l *LimitedDeadlineReader) Read(p []byte) (n int, err error) {
	return l.LR.Read(p)
}

func (l *LimitedDeadlineReader) SetReadDeadline(t time.Time) error {
	return l.R.SetReadDeadline(t)
}

// Benchmark standard copy
func BenchmarkCopy(b *testing.B) {
	ctx := context.Background()
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0600)
	if err != nil {
		b.Fatal(err)
	}
	zero, err := os.Open("/dev/zero")
	if err != nil {
		b.Fatal(err)
	}
	lzero := &LimitedDeadlineReader{
		R: zero,
		N: 1e9,
	}
	b.Run("vanilla-1MiB", func(b *testing.B) {
		lzero.Limit()
		_, err = io.Copy(null, lzero)
		if err != nil {
			b.Fatal(err)
		}
	})
	b.Run("context-1MiB", func(b *testing.B) {
		lzero.Limit()
		_, err = io.Copy(WriterAdapter(ctx, null), ReaderAdapter(ctx, lzero))
		if err != nil {
			b.Fatal(err)
		}
	})
	lzero = &LimitedDeadlineReader{
		R: zero,
		N: 1e12,
	}
	b.Run("vanilla-1GiB", func(b *testing.B) {
		lzero.Limit()
		_, err = io.Copy(null, lzero)
		if err != nil {
			b.Fatal(err)
		}
	})
	b.Run("context-1GiB", func(b *testing.B) {
		lzero.Limit()
		_, err = io.Copy(WriterAdapter(ctx, null), ReaderAdapter(ctx, lzero))
		if err != nil {
			b.Fatal(err)
		}
	})
}

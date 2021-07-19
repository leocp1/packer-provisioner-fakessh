// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"io"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/ctxio"
)

// Compare to github.com/hashicorp/nomad/client/lib/fifo
type pipe struct {
	Dir string
}

type deadlineReaderCloser interface {
	ctxio.DeadlineReader
	io.Closer
}

type deadlineWriterCloser interface {
	ctxio.DeadlineWriter
	io.Closer
}

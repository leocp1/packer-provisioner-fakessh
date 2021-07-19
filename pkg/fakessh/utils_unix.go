// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build darwin linux

package fakessh

import (
	"os"
	"path/filepath"
)

const (
	// The basename of the ssh executable
	SSHEXENAME = "ssh"
)

// Check if the passed directory has a ssh executable
func dirHasFakeSsh(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, SSHEXENAME))
	if err != nil {
		return false
	}
	return fi.Mode()&0111 != 0000
}

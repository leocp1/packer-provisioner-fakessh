// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Requires POSIX shell commands
// +build darwin linux

package localcommunicator_test

import "github.com/leocp1/packer-provisioner-fakessh/pkg/localcommunicator"

var tests = []struct {
	name     string
	cmd      string
	stdin    string
	stdout   string
	stderr   string
	exitCode int
}{
	{
		name:     "stdout",
		cmd:      "printf test",
		stdin:    "",
		stdout:   "test",
		stderr:   "",
		exitCode: 0,
	},
	{
		name:     "stderr",
		cmd:      "printf test 1>&2",
		stdin:    "",
		stdout:   "",
		stderr:   "test",
		exitCode: 0,
	},
	{
		name:     "stdin",
		cmd:      "sort",
		stdin:    "3\n1\n2\n",
		stdout:   "1\n2\n3\n",
		stderr:   "",
		exitCode: 0,
	},
	{
		name:     "exitCode",
		cmd:      "exit 42",
		stdin:    "",
		stdout:   "",
		stderr:   "",
		exitCode: 42,
	},
	{
		name:     "bell escape",
		cmd:      "printf \a",
		stdin:    "",
		stdout:   "\a",
		stderr:   "",
		exitCode: 0,
	},
	{
		name:     "cancelling",
		cmd:      "sleep 1h",
		stdin:    "",
		stdout:   "",
		stderr:   "",
		exitCode: localcommunicator.EXIT_FAILURE,
	},
}

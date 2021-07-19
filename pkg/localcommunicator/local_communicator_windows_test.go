// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build windows

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
		cmd:      "echo test",
		stdin:    "",
		stdout:   "test\r\n",
		stderr:   "",
		exitCode: 0,
	},
	{
		name:     "stderr",
		cmd:      "echo test 1>&2",
		stdin:    "",
		stdout:   "",
		stderr:   "test \r\n",
		exitCode: 0,
	},
	{
		name:     "stdin",
		cmd:      "sort",
		stdin:    "3\r\n1\r\n2\r\n",
		stdout:   "1\r\n2\r\n3\r\n",
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
		cmd:      "echo \a",
		stdin:    "",
		stdout:   "\a\r\n",
		stderr:   "",
		exitCode: 0,
	},
	{
		name:     "cancelling",
		cmd:      "ping -n 1000 127.0.0.1 >NUL & echo oops",
		stdin:    "",
		stdout:   "",
		stderr:   "",
		exitCode: localcommunicator.EXIT_FAILURE,
	},
}

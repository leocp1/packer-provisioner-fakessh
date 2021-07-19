// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// A communicator that runs the command locally.
//
// Created for testing purposes.
// Compared to github.com/hashicorp/packer/common/shell-local.Communicator,
// the command is passed in cmd instead of being stored in the communicator.
package localcommunicator

import (
	"errors"
	"io"
	"os"
	"os/exec"
)

const (
	EXIT_FAILURE = 255
	EXIT_SUCCESS = 0
)

type comm struct{}

// Create local communicator
func New() (result *comm, err error) {
	return &comm{}, nil
}

// Run a command and output an exit code
func RunExitCode(cmd *exec.Cmd) int {
	err := cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return exitError.ExitCode()
		} else {
			return EXIT_FAILURE
		}
	} else {
		return EXIT_SUCCESS
	}
}

func (c *comm) Upload(path string, input io.Reader, fi *os.FileInfo) error {
	return errors.New("Upload is not implemented")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	return errors.New("UploadDir is not implemented")
}

func (c *comm) Download(path string, output io.Writer) error {
	return errors.New("Download is not implemented")
}

func (c *comm) DownloadDir(dst string, src string, excl []string) error {
	return errors.New("DownloadDir is not implemented")
}

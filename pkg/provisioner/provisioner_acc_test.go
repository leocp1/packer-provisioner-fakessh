// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provisioner_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/helper/tests/acc"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner/shell"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/provisioner"
)

func TestFakeSshProvisioner(t *testing.T) {
	acc.TestProvisionersPreCheck("fakessh", t)
	acc.TestProvisionersAgainstBuilders(new(FakeSshAccTest), t)
}

type FakeSshAccTest struct{}

func (s *FakeSshAccTest) GetName() string {
	return "fakessh"
}

func (s *FakeSshAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("./testdata", "fakessh.json")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), err
}

func (s *FakeSshAccTest) GetProvisionerStore() packer.MapOfProvisioner {
	return packer.MapOfProvisioner{
		"fakessh": func() (packer.Provisioner, error) {
			return &Provisioner{}, nil
		},
		"shell": func() (packer.Provisioner, error) {
			return &shell.Provisioner{}, nil
		},
	}
}

func (s *FakeSshAccTest) IsCompatible(builder string, vmOS string) bool {
	return vmOS == "linux"
}

func (s *FakeSshAccTest) RunTest(c *command.BuildCommand, args []string) error {
	if code := c.Run(args); code != 0 {
		ui := c.Meta.Ui.(*packer.BasicUi)
		out := ui.Writer.(*bytes.Buffer)
		err := ui.ErrorWriter.(*bytes.Buffer)
		return fmt.Errorf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			out.String(),
			err.String())
	}
	return nil
}

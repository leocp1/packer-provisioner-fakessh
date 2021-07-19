// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate mapstructure-to-hcl2 -type Provisioner

package provisioner

import (
	"context"
	"errors"
	"net/http"

	"github.com/hashicorp/hcl/v2/hcldec"
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/fakessh"
)

type Provisioner struct {
	config    sl.Config
	sshExeDir string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}

	err = sl.Validate(&p.config)
	if err != nil {
		return err
	}

	sshExeDir, ok := fakessh.FakeSshPath()
	if !ok {
		return errors.New("fake ssh executable not found")
	}
	p.sshExeDir = sshExeDir

	return nil
}

func (p *Provisioner) Provision(
	ctx context.Context,
	ui packer.Ui,
	comm packer.Communicator,
	generatedData map[string]interface{},
) error {
	var err error = nil

	srv, err := fakessh.NewServer(comm, "")
	if err != nil {
		return err
	}
	srvChan := make(chan error)
	go func() {
		srvChan <- srv.Serve()
	}()

	p.config.Vars, err =
		fakessh.AddFakeSshPath(p.config.Vars, p.sshExeDir, srv.Dir)
	if err != nil {
		return err
	}
	/*
		DO NOT rerun sl.Validate here:

		At least at the time of this comment (Packer v1.6.3) sl.Validate also
		modifies the config, leading to errors if the script is passed in as a
		file and Validate is run twice. In the future, it may be necessary to
		run Validate here instead of in Prepare but for now, modifying
		p.config.Vars after running Validate is safe.
	*/

	_, retErr := sl.Run(ctx, ui, &p.config, generatedData)

	srv.Shutdown(ctx)
	err = <-srvChan
	if err != http.ErrServerClosed {
		return err
	}

	return retErr
}

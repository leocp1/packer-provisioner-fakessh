// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"github.com/hashicorp/packer/packer/plugin"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/provisioner"
)

func main() {
	var err error = nil
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	err = server.RegisterProvisioner(new(Provisioner))
	if err != nil {
		panic(err)
	}
	server.Serve()
}

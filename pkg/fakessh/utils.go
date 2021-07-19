// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"context"
	"errors"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/yookoala/realpath"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/envmap"
)

const (
	// Name of the environment variable containing the unix domain socket
	// providing RPC for fake ssh.
	RPCDirEnvVarName = "PACKER_FAKE_SSH_RPC_DIR"

	// Name of the environment variable that can set the path to the fakessh
	// command
	SSHEXEEnvVarName = "PACKER_FAKE_SSH_EXECUTABLE_PATH"

	// The name of the go package containing the fake ssh command
	SSHPKGNAME = "github.com/leocp1/packer-provisioner-fakessh/cmd/ssh"
)

// Return a reasonable guess for the parent directory of the fake ssh binary
func FakeSshPath() (string, bool) {
	var d string
	var err error

	// Check if it has been set in the environment variable
	d = os.Getenv(SSHEXEEnvVarName)
	d, err = realpath.Realpath(d)
	if err == nil && dirHasFakeSsh(d) {
		return d, true
	}

	// Check if it has been patched in with nix
	d = "@out@/share/bin"
	if d != "@"+"out@/share/bin" {
		d, err = realpath.Realpath(d)
		if err == nil && dirHasFakeSsh(d) {
			return d, true
		}
	}

	// Check executable folder
	d, err = osext.ExecutableFolder()
	if err == nil {
		d, err = realpath.Realpath(d)
		if err == nil && dirHasFakeSsh(d) {
			return d, true
		}
	}

	// Check command install directory
	pkg, err := build.Import(SSHPKGNAME, "", 0)
	d = pkg.BinDir
	if d != "" {
		d, err = realpath.Realpath(d)
		if err == nil && dirHasFakeSsh(d) {
			return d, true
		}
	}

	return "", false
}

// Modify environment slice to use fakessh instead.
// If the PATH variable is unset,
// set it to its value from os.Environ
func AddFakeSshPath(es []string, sshDir string, rpcDir string) ([]string, error) {
	var ok bool = false
	var err error = nil
	if sshDir == "" {
		sshDir, ok = FakeSshPath()
		if !ok {
			return es, errors.New("AddFakeSshPath: ssh directory not found")
		}
	}
	if !dirHasFakeSsh(sshDir) {
		return es, errors.New("AddFakeSshPath: ssh directory not found")
	}
	em := envmap.NewEnvMap()
	em.M["PATH"] = &envmap.Path{}
	err = em.AddSlice(es)
	if err != nil {
		return es, err
	}

	path := em.M["PATH"].(*envmap.Path)
	if len(path.S) == 0 {
		osem := envmap.NewEnvMap()
		osem.M["PATH"] = &envmap.Path{}
		err = osem.AddSlice(os.Environ())
		if err != nil {
			return es, err
		}
		path.S = osem.M["PATH"].(*envmap.Path).S
	}
	newpath := make([]string, 0, len(path.S)+1)
	newpath = append(newpath, sshDir)
	newpath = append(newpath, path.S...)
	path.S = newpath

	em.M[RPCDirEnvVarName] = &envmap.String{S: rpcDir}

	return em.EnvSlice(), nil
}

// Create a temporary directory and build a ssh binary inside
func GoBuildFakeSsh(ctx context.Context) (string, error) {
	sshDir, err := ioutil.TempDir("", "packer-provisioner-fakessh")
	if err != nil {
		return sshDir, err
	}
	sshExe := filepath.Join(sshDir, SSHEXENAME)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", sshExe, SSHPKGNAME)
	err = cmd.Run()
	if err != nil {
		return sshDir, err
	}

	return sshDir, nil
}

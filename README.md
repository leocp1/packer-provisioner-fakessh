# `fakessh` Provisioner

Type: `fakessh`

The `fakessh` [Packer](https://www.packer.io/) provisioner is a modification of
the [`shell-local`](https://www.packer.io/docs/provisioners/shell-local.html)
provisioner that provides a fake `ssh` command that forwards its arguments to
the Communicator. It's intended for use in commands like `nix` and `git` that
use `ssh` to access remote servers.

This is a bit of a hack: consider using the `Host`, `Port`, `User`, `Password`,
`SSHPublicKey` and `SSHPrivateKey` functions provided by the
[Template Engine](https://www.packer.io/docs/templates/engine) instead. One main
advantage of using a system `ssh` directly instead of this provisioner is
support for `gzip` compression, which is
[currently unimplemented](https://github.com/golang/go/issues/31369) in the Go
`ssh` library:

## Basic Example

```json
{
  "type": "fakessh",
  "inline": [
    "nix-copy-closure --to user@server /nix/store/hashofmynixossystemconfiguration-nixos-system-hostname-version"
  ]
}
```

`nix-copy-closure` internally calls `ssh user@server "nix-daemon --stdio"`. The
provisioner will change the `PATH` variable so that `ssh` points to a fake `ssh`
command. The fake `ssh` command will ignore the destination and any command line
flags and forward the "nix-daemon --stdio" command to the Packer Communicator.
(`fakessh` doesn't support interactive `ssh` sessions.)

## Configuration Reference

The configuration options are the same as the
[`shell-local`](https://www.packer.io/docs/provisioners/shell-local.html)
provisioner.

If the provisioner is reporting it can not find the `ssh` directory,

- Run `go build github.com/leocp1/packer-provisioner-fakessh/cmd/ssh`
- Set `PACKER_FAKE_SSH_EXECUTABLE_PATH` to the parent directory of the fake
  `ssh` executable produced

## Acceptance test

Running the acceptance test requires

- Copying or symlinking the `build` directory of the
  [Packer repository](https://github.com/hashicorp/packer) to the root directory
  of this repository
- Setting `PACKER_FAKE_SSH_EXECUTABLE_PATH` to the parent directory of the fake
  `ssh` executable
- Setting `ACC_TEST_BUILDERS` to the builders to run the test on
- Setting `ACC_TEST_PROVISIONERS` to include `fakessh,shell`
- Running `go test ./pkg/provisioner`

## See Also

- [packer-provisioner-tunnel](https://github.com/josharian/packer-provisioner-tunnel)

## Licenses

The file `./pkg/ctxio/ctxio.go` in this repository contains an adaptation of
functions published in "Cancelling I/O in Go Cap'n Proto" by Ross Light
([link](https://medium.com/@zombiezen/canceling-i-o-in-go-capn-proto-5ae8c09c5b29))
which was licensed under the
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) License.
[![License: CC BY 4.0](https://licensebuttons.net/l/by/4.0/80x15.png)](https://creativecommons.org/licenses/by/4.0/)

All other Go source files (files with extension `.go`) in this repository are
licensed under the Mozilla Public License 2.0.
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

All Nix files (files with extension `.nix`) in this repository are licensed
under the MIT License.
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Requires POSIX shell commands
// +build darwin linux

package fakessh_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yookoala/realpath"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/fakessh"
	"github.com/leocp1/packer-provisioner-fakessh/pkg/localcommunicator"
)

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
		exitCode: -1,
	},
}

// Clone git repository.
// Not tested on Windows, since having git correctly parse Windows paths is tricky.
func TestGitClone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), MAXTESTTIME)
	defer cancel()

	comm, err := localcommunicator.New()
	if err != nil {
		t.Fatal(err)
	}
	srv, err := fakessh.NewServer(comm, "")
	if err != nil {
		t.Fatal(err)
	}
	srvChan := make(chan error)
	go func() {
		srvChan <- srv.Serve()
	}()

	sshExeDir, ok := fakessh.FakeSshPath()
	if !ok {
		sshExeDir, err = fakessh.GoBuildFakeSsh(ctx)
		defer os.RemoveAll(sshExeDir)
		if err != nil {
			t.Skip("ssh executable not found or buildable")
		}
	}
	t.Logf("Using ssh path:%s", sshExeDir)

	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}
	git, err = realpath.Realpath(git)
	if err != nil {
		t.Fatal("finding absolute path of git failed")
	}
	t.Logf("Using git path: %s", git)

	t.Logf("git init")
	srcGitRepo, _ := ioutil.TempDir("", "fakessh-src-git-repo")
	exec.CommandContext(ctx, git, "-C", srcGitRepo, "init").Run()

	t.Logf("git add")
	message := []byte("My git commit")
	testfile := filepath.Join(srcGitRepo, "testfile")
	ioutil.WriteFile(testfile, message, 0644)
	exec.CommandContext(ctx, git, "-C", srcGitRepo, "add", testfile).Run()

	t.Logf("git commit")
	exec.CommandContext(ctx,
		git, "-C", srcGitRepo, "commit", "-m", "msg",
	).Run()

	env, err := fakessh.AddFakeSshPath(nil, sshExeDir, srv.Dir)
	if err != nil {
		t.Fatal("AddFakeSshPath failed")
	}
	sshExe := filepath.Join(sshExeDir, fakessh.SSHEXENAME)
	env = append(env, fmt.Sprintf("GIT_SSH=%s", sshExe))
	t.Logf("env: %#v", env)

	t.Logf("git clone")
	dstGitRepo, _ := ioutil.TempDir("", "fakessh-dst-git-repo")
	cmd := exec.CommandContext(ctx,
		git, "clone", fmt.Sprintf(`git@githost:%s`, srcGitRepo), dstGitRepo)
	cmd.Env = env
	cmd.Stderr = &tWriter{t: t}
	cmd.Run()

	// Test
	got, _ := ioutil.ReadFile(filepath.Join(dstGitRepo, "testfile"))
	if !bytes.Equal(got, message) {
		t.Error("git clone failed")
	}

	srv.Shutdown(ctx)
	err = <-srvChan
	if err != http.ErrServerClosed {
		t.Error(err)
	}
}

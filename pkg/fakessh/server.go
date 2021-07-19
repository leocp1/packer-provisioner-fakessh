// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/communicator/none"
	"github.com/hashicorp/packer/packer"
	"github.com/keegancsmith/rpc"

	"github.com/leocp1/packer-provisioner-fakessh/pkg/ctxio"
)

const (
	// Default path of unix domain socket used for fakessh RPC
	UDSPath = "/fakessh.sock"
	// Default exit code on failure
	EXIT_FAILURE = 255
)

// A type representing a server that forwards ssh commands to a packer
// Communicator.
type server struct {
	// rpc http server
	Server *http.Server
	// rpc uds listener
	Ln net.Listener
	// rpc uds sock directory
	Dir string
}

// Allocates and initializes a new fakessh server with uds socket in dir.
// If comm is nil, ignore passed commands.
// If dir is the empty string, create a temporary directory.
func NewServer(comm packer.Communicator, dir string) (*server, error) {
	var err error = nil
	if comm == nil {
		comm, err = none.New("")
		if err != nil {
			return nil, err
		}
	}
	rpcssh := &RpcSsh{
		Comm: comm,
		M:    make(map[RpcCmd]RpcState),
	}

	// equivalent to rpc.Register(rpcssh)
	rpcSrv := rpc.NewServer()
	err = rpcSrv.Register(rpcssh)
	if err != nil {
		return nil, err
	}

	// equivalent to rpc.HandleHTTP(), but without debugHTTP
	srvMux := http.NewServeMux()
	srvMux.Handle(rpc.DefaultRPCPath, rpcSrv)

	if dir == "" {
		dir, err = ioutil.TempDir("", "fakessh")
		if err != nil {
			return nil, err
		}
	}
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	udsDir := filepath.Join(dir, UDSPath)
	ln, err := net.Listen("unix", udsDir)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	httpSrv := &http.Server{Handler: srvMux}

	srv := &server{
		Server: httpSrv,
		Ln:     ln,
		Dir:    dir,
	}

	return srv, nil
}

// Run fake ssh server.
//
// Like http.Server, returns http.ErrServerClosed on Shutdown
func (srv *server) Serve() error {
	return srv.Server.Serve(srv.Ln)
}

// Gracefully stop fake ssh server and delete working directory
func (srv *server) Shutdown(ctx context.Context) error {
	serr := srv.Server.Shutdown(ctx)
	lerr := srv.Ln.Close()
	derr := os.RemoveAll(srv.Dir)

	if serr != nil {
		return serr
	}
	if lerr != nil {
		return lerr
	}
	if derr != nil {
		return derr
	}
	return nil
}

// A command to send to the communicator
type Cmd struct {
	Command string
	Stdin   deadlineReaderCloser
	Stdout  deadlineWriterCloser
	Stderr  deadlineWriterCloser
}

// Run cmd on fake ssh server with working directory dir.
//
// Closes cmd.Stdin, cmd.Stdout, and cmd.Stderr.
func RunCmd(
	ctx context.Context,
	dir string,
	cmd *Cmd,
) (exitCode int, err error) {

	exitCode = EXIT_FAILURE
	err = nil
	udsDir := filepath.Join(dir, UDSPath)

	// dial Server
	cli, err := rpc.DialHTTP("unix", udsDir)
	if err != nil {
		return
	}

	// create pipes
	inpipe, err := makePipe("stdin")
	if err != nil {
		return
	}
	defer inpipe.Close()
	outpipe, err := makePipe("stdout")
	if err != nil {
		return
	}
	defer outpipe.Close()
	errpipe, err := makePipe("stderr")
	if err != nil {
		return
	}
	defer errpipe.Close()

	// open client read end
	outpipef, err := openReadPipe(ctx, outpipe.Dir)
	if err != nil {
		return EXIT_FAILURE, err
	}
	defer outpipef.Close()
	errpipef, err := openReadPipe(ctx, errpipe.Dir)
	if err != nil {
		return EXIT_FAILURE, err
	}
	defer errpipef.Close()

	c := &RpcCmd{
		Cmd:        cmd.Command,
		StdinPipe:  inpipe.Dir,
		StdoutPipe: outpipe.Dir,
		StderrPipe: errpipe.Dir,
	}

	// open server pipes
	err = cli.Call(ctx, "RpcSsh.OpenPipes", c, &exitCode)
	if err != nil {
		return EXIT_FAILURE, err
	}

	// open client write end
	inpipef, err := openWritePipe(ctx, inpipe.Dir)
	if err != nil {
		return EXIT_FAILURE, err
	}
	// inpipef is a writer, so it is closed in ctxCopy

	// Error channels to signal Copy completed
	inCopyErr := make(chan error)
	outCopyErr := make(chan error)
	errCopyErr := make(chan error)

	// copy, and close writer
	ctxCopy := func(
		ctx context.Context,
		c chan<- error,
		w deadlineWriterCloser,
		r ctxio.DeadlineReader,
	) {
		_, err := io.Copy(
			ctxio.WriterAdapter(ctx, w),
			ctxio.ReaderAdapter(ctx, r),
		)
		closeerr := w.Close()
		if err == nil {
			err = closeerr
		}
		c <- err
		close(c)
	}
	go ctxCopy(ctx, inCopyErr, inpipef, cmd.Stdin)
	defer cmd.Stdin.Close()
	go ctxCopy(ctx, outCopyErr, cmd.Stdout, outpipef)
	go ctxCopy(ctx, errCopyErr, cmd.Stderr, errpipef)

	// run Cmd
	err = cli.Call(ctx, "RpcSsh.Run", c, &exitCode)

	// wait for copiers to finish
	inerr := <-inCopyErr
	outerr := <-outCopyErr
	errerr := <-errCopyErr
	if err != nil {
		return EXIT_FAILURE, err
	}
	if inerr != nil && !os.IsTimeout(inerr) {
		err = inerr
		return
	}
	if outerr != nil && !os.IsTimeout(outerr) {
		err = outerr
		return
	}
	if errerr != nil && !os.IsTimeout(errerr) {
		err = errerr
		return exitCode, errerr
	}

	return
}

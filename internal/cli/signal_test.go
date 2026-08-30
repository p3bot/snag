// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestSignalExitCode(t *testing.T) {
	t.Cleanup(func() { caughtSignal.Store(0) })

	caughtSignal.Store(0)
	if got := SignalExitCode(); got != 0 {
		t.Errorf("no signal → %d want 0", got)
	}
	caughtSignal.Store(int32(syscall.SIGINT))
	if got := SignalExitCode(); got != 130 {
		t.Errorf("SIGINT → %d want 130", got)
	}
	caughtSignal.Store(int32(syscall.SIGTERM))
	if got := SignalExitCode(); got != 143 {
		t.Errorf("SIGTERM → %d want 143", got)
	}
}

func TestSignalContext_SecondInterruptTerminates(t *testing.T) {
	if os.Getenv("SNAG_SIGNAL_HELPER") == "1" {
		ctx, stop := signalContext()
		defer stop()
		fmt.Println("ready")
		<-ctx.Done()
		fmt.Println("cancelled")
		select {}
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSignalContext_SecondInterruptTerminates$")
	cmd.Env = append(os.Environ(), "SNAG_SIGNAL_HELPER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	r := bufio.NewReader(stdout)
	readLine := func(want string) {
		t.Helper()
		line, err := r.ReadString('\n')
		if err != nil {
			_ = cmd.Process.Kill()
			t.Fatalf("helper %s: %v", want, err)
		}
		if line != want+"\n" {
			_ = cmd.Process.Kill()
			t.Fatalf("helper line = %q, want %s", line, want)
		}
	}

	readLine("ready")
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		_ = cmd.Process.Kill()
		t.Fatal(err)
	}
	readLine("cancelled")
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatal("helper exited on first SIGINT; want it to hang until the second")
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		_ = cmd.Process.Kill()
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("second SIGINT did not terminate helper")
	}
}

func TestAbortErr(t *testing.T) {
	if abortErr(context.Background(), nil) != nil {
		t.Fatal("no cancel, no error: want nil")
	}
	if abortErr(context.Background(), errors.New("boom")) != nil {
		t.Fatal("ordinary error is not abort")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(abortErr(ctx, errors.New("boom")), context.Canceled) {
		t.Fatal("cancelled ctx should abort even if err is unrelated")
	}
	if !errors.Is(abortErr(context.Background(), context.Canceled), context.Canceled) {
		t.Fatal("Canceled err should abort")
	}
	if abortErr(context.Background(), context.DeadlineExceeded) != nil {
		t.Fatal("deadline is not process cancel")
	}
}

func TestCLI_PageLoadTimeout_NoSignal(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hang(r, 15*time.Second)
	}))
	t.Cleanup(server.Close)

	_, stderr, err := runSnag("--force-headless", "--timeout", "2", server.URL)
	assertExitCode(t, err, ExitCodeError)
	assertContains(t, stderr, "Error:")
	if !bytes.Contains([]byte(stderr), []byte("timeout")) && !bytes.Contains([]byte(stderr), []byte("Timeout")) {
		t.Errorf("page-load timeout should mention timeout, stderr=%q", stderr)
	}
}

func TestCLI_SIGINT_DuringFetch(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-started:
		default:
			close(started)
		}
		hang(r, 30*time.Second)
	}))
	t.Cleanup(server.Close)

	stdout, stderr, err := runSnagSignal(t, os.Interrupt, started, "--force-headless", "--timeout", "25", server.URL)
	assertExitCode(t, err, ExitCodeInterrupt)
	assertInterruptStderr(t, stderr)
	_ = stdout
}

func TestCLI_SIGTERM_DuringFetch(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-started:
		default:
			close(started)
		}
		hang(r, 30*time.Second)
	}))
	t.Cleanup(server.Close)

	_, stderr, err := runSnagSignal(t, syscall.SIGTERM, started, "--force-headless", "--timeout", "25", server.URL)
	assertExitCode(t, err, ExitCodeSIGTERM)
	assertInterruptStderr(t, stderr)
}

func TestCLI_SIGINT_StopsRemainingURLs(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	var hits atomic.Int32
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a", "/b":
			if hits.Add(1) == 1 {
				close(started)
			}
			hang(r, 30*time.Second)
		}
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	_, stderr, err := runSnagSignal(t, os.Interrupt, started,
		"--force-headless", "--timeout", "25", "-d", dir,
		server.URL+"/a", server.URL+"/b")
	assertExitCode(t, err, ExitCodeInterrupt)
	assertInterruptStderr(t, stderr)
	assertNotContains(t, stderr, "remain open")
	if got := hits.Load(); got > 1 {
		t.Errorf("remaining URL was started: path hits=%d", got)
	}
}

func TestCLI_SIGINT_DuringConnect(t *testing.T) {
	port, accepted := hangingDebugPort(t)
	_, stderr, err := runSnagSignal(t, os.Interrupt, accepted, "--list-tabs", "--port", strconv.Itoa(port))
	assertExitCode(t, err, ExitCodeInterrupt)
	assertInterruptStderr(t, stderr)
}

func TestCLI_SIGTERM_DuringConnect(t *testing.T) {
	port, accepted := hangingDebugPort(t)
	_, stderr, err := runSnagSignal(t, syscall.SIGTERM, accepted, "--list-tabs", "--port", strconv.Itoa(port))
	assertExitCode(t, err, ExitCodeSIGTERM)
	assertInterruptStderr(t, stderr)
}

func TestCLI_SIGINT_DuringPDF(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.NotFound(w, r)
			return
		}
		select {
		case <-started:
		default:
			close(started)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<!doctype html><html><body><h1>pdf interrupt</h1></body></html>")
	}))
	t.Cleanup(server.Close)

	out := filepath.Join(t.TempDir(), "page.pdf")
	_, stderr, err := runSnagSignalOnStderr(t, os.Interrupt, started, "Generating PDF...",
		"--force-headless", "--verbose", "--format", "pdf", "-o", out, "--timeout", "25", server.URL)
	assertExitCode(t, err, ExitCodeInterrupt)
	assertInterruptStderr(t, stderr)
}

func hangingDebugPort(t *testing.T) (port int, accepted <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	ch := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			select {
			case <-ch:
			default:
				close(ch)
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1)
				_, _ = c.Read(buf)
				_, _ = c.Read(buf)
			}(conn)
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, ch
}

func hang(r *http.Request, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-r.Context().Done():
	case <-timer.C:
	}
}

func assertInterruptStderr(t *testing.T, stderr string) {
	t.Helper()
	assertNotContains(t, stderr, "Received")
	assertNotContains(t, stderr, "cleaning up")
	assertNotContains(t, stderr, "context canceled")
	assertNotContains(t, stderr, "Error:")
}

func runSnagSignal(t *testing.T, sig os.Signal, started <-chan struct{}, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	return runSnagSignalled(t, sig, started, "", args...)
}

func runSnagSignalOnStderr(t *testing.T, sig os.Signal, started <-chan struct{}, needle string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	return runSnagSignalled(t, sig, started, needle, args...)
}

func runSnagSignalled(t *testing.T, sig os.Signal, started <-chan struct{}, needle string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(snagBin, args...)
	var outBuf bytes.Buffer
	var errMu sync.Mutex
	var errBuf bytes.Buffer
	stderrPipe, pipeErr := cmd.StderrPipe()
	if pipeErr != nil {
		t.Fatalf("stderr pipe: %v", pipeErr)
	}
	cmd.Stdout = &outBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("start snag: %v", err)
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buf := make([]byte, 4096)
		for {
			n, rerr := stderrPipe.Read(buf)
			if n > 0 {
				errMu.Lock()
				errBuf.Write(buf[:n])
				errMu.Unlock()
			}
			if rerr != nil {
				return
			}
		}
	}()

	stderrSnapshot := func() string {
		errMu.Lock()
		defer errMu.Unlock()
		return errBuf.String()
	}

	killWait := func(msg string) {
		t.Helper()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		<-readDone
		t.Fatalf("%s; stderr=%q", msg, stderrSnapshot())
	}

	timer := time.NewTimer(25 * time.Second)
	defer timer.Stop()
	select {
	case <-started:
	case <-timer.C:
		killWait("timed out waiting for work to start")
	}

	if needle != "" {
		deadline := time.NewTimer(25 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			if strings.Contains(stderrSnapshot(), needle) {
				break
			}
			select {
			case <-deadline.C:
				killWait("timed out waiting for " + needle)
			case <-readDone:
				killWait("stderr closed before " + needle)
			case <-ticker.C:
			}
		}
	}

	if err := cmd.Process.Signal(sig); err != nil {
		killWait("signal: " + err.Error())
	}

	err = cmd.Wait()
	<-readDone
	return outBuf.String(), stderrSnapshot(), err
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	llsyscall "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

// requireSandbox skips the test when unprivileged user namespaces are not
// available (common on hardened CI). It returns nil when the sandbox can run.
func requireSandbox(t *testing.T) {
	t.Helper()
	if err := checkUserNamespaces(); err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func TestRunEcho(t *testing.T) {
	requireSandbox(t)

	shBin := lookBin(t, "cat", "/bin/cat", "/usr/bin/cat")

	var out bytes.Buffer
	p := Profile{
		RootFS:   RootMinimal,
		WorkDir:  "/",
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Env:      []string{"PATH=/usr/bin:/bin"},
	}
	res, err := Run(context.Background(), p, Cmd{
		Path:   shBin,
		Args:   nil,
		Stdin:  strings.NewReader("hello-sbox"),
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code = %d", res.ExitCode)
	}
	if got := out.String(); got != "hello-sbox" {
		t.Fatalf("output = %q", got)
	}
}

// TestSecretsUnreachable is the core security assertion: a path that is NOT
// bound into the sandbox must not be readable from inside it.
func TestSecretsUnreachable(t *testing.T) {
	requireSandbox(t)

	// Create a secret file on the host, outside any bind.
	secretDir := t.TempDir()
	secret := secretDir + "/secret.txt"
	if err := os.WriteFile(secret, []byte("top-secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	catBin := lookBin(t, "cat", "/bin/cat", "/usr/bin/cat")

	var out, errBuf bytes.Buffer
	p := Profile{
		RootFS:   RootMinimal,
		WorkDir:  "/",
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Env:      []string{"PATH=/usr/bin:/bin"},
	}
	res, err := Run(context.Background(), p, Cmd{
		Path:   catBin,
		Args:   []string{secret}, // absolute host path
		Stdout: &out,
		Stderr: &errBuf,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected cat to fail reading the secret, but it succeeded: %q", out.String())
	}
	if strings.Contains(out.String(), "top-secret") {
		t.Fatalf("SECRET LEAKED into sandbox output: %q", out.String())
	}
}

func TestTimeoutKills(t *testing.T) {
	requireSandbox(t)

	sleepBin := lookBin(t, "sleep", "/bin/sleep", "/usr/bin/sleep")

	p := Profile{
		RootFS:   RootMinimal,
		WorkDir:  "/",
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Env:      []string{"PATH=/usr/bin:/bin"},
		Limits:   Limits{Wall: 500 * time.Millisecond},
	}
	start := time.Now()
	res, err := Run(context.Background(), p, Cmd{
		Path: sleepBin,
		Args: []string{"30"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.TimedOut {
		t.Fatalf("expected TimedOut")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("timeout did not fire promptly")
	}
}

func lookBin(t *testing.T, name string, candidates ...string) string {
	t.Helper()
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Skipf("%s not found in %v", name, candidates)
	return ""
}

func TestContextCancelKills(t *testing.T) {
	requireSandbox(t)

	sleepBin := lookBin(t, "sleep", "/bin/sleep", "/usr/bin/sleep")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	p := Profile{
		RootFS:   RootMinimal,
		WorkDir:  "/",
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Env:      []string{"PATH=/usr/bin:/bin"},
	}
	_, err := Run(ctx, p, Cmd{Path: sleepBin, Args: []string{"30"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestSeccompBlocksUnshare verifies the strict seccomp policy denies unshare,
// which is one of the dangerous syscalls used for sandbox escapes.
func TestSeccompBlocksUnshare(t *testing.T) {
	requireSandbox(t)

	unshareBin := lookBin(t, "unshare", "/usr/bin/unshare", "/bin/unshare")

	p := Profile{
		RootFS:   RootMinimal,
		WorkDir:  "/",
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Env:      []string{"PATH=/usr/bin:/bin"},
	}
	res, err := Run(context.Background(), p, Cmd{
		Path: unshareBin,
		Args: []string{"--user", "true"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected unshare to be blocked by seccomp, but it succeeded")
	}
}

// TestNetNoneHasNoLoopback verifies that NetNone yields a network namespace
// where loopback is down, while NetLoopback brings it up. The check runs by
// re-executing the test binary inside the sandbox.
func TestNetModes(t *testing.T) {
	requireSandbox(t)

	if os.Getenv("SBOX_CHECK_LO") == "1" {
		// This branch runs inside the sandbox.
		ifs, _ := net.Interfaces()
		for _, i := range ifs {
			if i.Name == "lo" && i.Flags&net.FlagUp != 0 {
				os.Exit(10) // lo is up
			}
		}
		os.Exit(11) // lo is down / absent
		return
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		net      NetMode
		wantUp   bool
		wantExit int
	}{
		{NetNone, false, 11},
		{NetLoopback, true, 10},
	}
	for _, tc := range cases {
		p := Profile{
			RootFS:   RootMinimal,
			WorkDir:  "/",
			Net:      tc.net,
			Seccomp:  SeccompStrict,
			Landlock: true,
			Env:      []string{"PATH=/usr/bin:/bin", "SBOX_CHECK_LO=1"},
			Binds:    []Bind{{Host: exe}},
		}
		res, err := Run(context.Background(), p, Cmd{
			Path: exe,
			Args: []string{"-test.run", "TestNetModes"},
		})
		if err != nil {
			t.Fatalf("net=%v Run: %v", tc.net, err)
		}
		if res.ExitCode != tc.wantExit {
			t.Fatalf("net=%v: got exit %d, want %d (loopback up=%v expected)",
				tc.net, res.ExitCode, tc.wantExit, tc.wantUp)
		}
	}
}

// landlockAvailable skips the test when landlock cannot be enforced on this
// kernel (e.g. Docker Desktop's LinuxKit VM has no landlock LSM).
func landlockAvailable(t *testing.T) {
	t.Helper()
	v, err := llsyscall.LandlockGetABIVersion()
	if err != nil || v < 1 {
		t.Skipf("landlock unavailable (abi=%d, err=%v)", v, err)
	}
}

// TestLandlockOnlySecretsUnreachable is the core proof for IsolationLandlockOnly:
// it runs WITHOUT any namespaces (so it works under a hardened confinement) and
// asserts that a path that is not in the allowlist cannot be read.
func TestLandlockOnlySecretsUnreachable(t *testing.T) {
	landlockAvailable(t)

	secretDir := t.TempDir()
	secret := secretDir + "/secret.txt"
	if err := os.WriteFile(secret, []byte("top-secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	catBin := lookBin(t, "cat", "/bin/cat", "/usr/bin/cat")
	workDir := t.TempDir()

	var out, errBuf bytes.Buffer
	p := Profile{
		Isolation: IsolationLandlockOnly,
		Binds:     []Bind{{Host: workDir, Writable: true}},
		WorkDir:   workDir,
		Seccomp:   SeccompStrict,
		Landlock:  true,
		Env:       []string{"PATH=/usr/bin:/bin"},
	}
	res, err := Run(context.Background(), p, Cmd{
		Path:   catBin,
		Args:   []string{secret},
		Stdout: &out,
		Stderr: &errBuf,
	})
	if err != nil {
		t.Fatalf("Run: %v (stderr=%q)", err, errBuf.String())
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected cat to be denied, but it succeeded: %q", out.String())
	}
	if bytes.Contains(out.Bytes(), []byte("top-secret")) {
		t.Fatalf("SECRET LEAKED in landlock-only mode: %q", out.String())
	}
}

// TestLandlockOnlyAllowsBind verifies that an explicitly bound path IS readable
// in landlock-only mode.
func TestLandlockOnlyAllowsBind(t *testing.T) {
	landlockAvailable(t)

	workDir := t.TempDir()
	allowed := workDir + "/data.txt"
	if err := os.WriteFile(allowed, []byte("visible"), 0o644); err != nil {
		t.Fatal(err)
	}

	catBin := lookBin(t, "cat", "/bin/cat", "/usr/bin/cat")

	var out bytes.Buffer
	p := Profile{
		Isolation: IsolationLandlockOnly,
		Binds:     []Bind{{Host: workDir, Writable: false}},
		WorkDir:   workDir,
		Seccomp:   SeccompStrict,
		Landlock:  true,
		Env:       []string{"PATH=/usr/bin:/bin"},
	}
	res, err := Run(context.Background(), p, Cmd{
		Path:   catBin,
		Args:   []string{allowed},
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit 0 reading allowed bind, got %d", res.ExitCode)
	}
	if got := out.String(); got != "visible" {
		t.Fatalf("output = %q, want %q", got, "visible")
	}
}

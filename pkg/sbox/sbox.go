// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sbox

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrUnsupported is returned by [Run] when the current platform cannot provide
// a real sandbox. On non-linux platforms a passthrough stub is used instead
// (see the package documentation), which only returns this error when isolation
// is explicitly required via the SBOX_REQUIRE_ISOLATION environment variable.
var ErrUnsupported = errors.New("sbox: sandbox not supported on this platform")

// ErrNoUserNamespace is returned when unprivileged user namespaces are not
// available on the host. This is a hard failure: sbox never silently degrades
// to a weaker isolation, because the whole point is to protect the calling
// application's secrets and database from untrusted child processes.
var ErrNoUserNamespace = errors.New("sbox: unprivileged user namespaces are not available")

// ErrNotInitialized is returned by [Run] when [Init] was not called at the very
// beginning of main. Without Init the re-exec trampoline cannot take over the
// child process and set up the sandbox.
var ErrNotInitialized = errors.New("sbox: Init was not called at the start of main")

// NetMode controls the network isolation of the sandboxed process.
type NetMode int

const (
	// NetHost shares the host network namespace. This is the default because
	// the initial use cases (go toolchain, git) commonly require network
	// access. The sandboxed process can still not reach the application's
	// secrets, but it can talk to the network.
	NetHost NetMode = iota

	// NetLoopback creates a fresh network namespace with only a configured
	// loopback interface. Useful for inspecting untrusted webservers that bind
	// to localhost without giving them outbound connectivity.
	NetLoopback

	// NetNone creates a fresh, empty network namespace. The process has no
	// network interfaces at all (not even loopback).
	NetNone
)

// RootFSKind selects how the sandbox root filesystem is assembled.
type RootFSKind int

const (
	// RootTmpfs mounts a fresh, empty tmpfs as the new root. Only the paths
	// explicitly added via [Profile.Binds] become visible. This is the most
	// hermetic option.
	RootTmpfs RootFSKind = iota

	// RootMinimal is like RootTmpfs but additionally bind-mounts a small set of
	// common read-only system paths (/usr, /bin, /lib, /lib64, /etc/ssl,
	// /etc/resolv.conf) so that dynamically linked tools and TLS just work.
	RootMinimal
)

// SeccompMode selects the seccomp-bpf syscall policy.
type SeccompMode int

const (
	// SeccompStrict installs a restrictive allowlist that blocks dangerous
	// syscalls (ptrace, bpf, keyctl, mount after setup, kexec, ...). This is
	// the recommended default.
	SeccompStrict SeccompMode = iota

	// SeccompPermissive installs a minimal blocklist only. Use when a workload
	// legitimately needs syscalls the strict profile denies.
	SeccompPermissive

	// SeccompOff disables seccomp entirely. Only namespaces and landlock apply.
	SeccompOff
)

// Bind describes a host path that is made available inside the sandbox.
type Bind struct {
	// Host is the absolute path on the host filesystem.
	Host string
	// Target is the absolute path inside the sandbox. If empty, Host is used.
	Target string
	// Writable makes the bind mount read-write. Defaults to read-only.
	Writable bool
	// Optional skips the bind if the host path does not exist instead of
	// failing the sandbox setup.
	Optional bool
}

// Limits describes resource limits applied to the sandboxed process via
// setrlimit and a wall-clock timeout. A zero value means "no explicit limit".
type Limits struct {
	// CPUTime is the maximum CPU time (RLIMIT_CPU). Rounded down to seconds.
	CPUTime time.Duration
	// AddressSpace is the maximum virtual memory size in bytes (RLIMIT_AS).
	AddressSpace uint64
	// NoFile is the maximum number of open file descriptors (RLIMIT_NOFILE).
	NoFile uint64
	// NProc is the maximum number of processes/threads (RLIMIT_NPROC).
	NProc uint64
	// FileSize is the maximum size of files the process may create
	// (RLIMIT_FSIZE) in bytes.
	FileSize uint64
	// Wall is the wall-clock timeout. When exceeded, the whole sandbox process
	// tree is killed. A zero value means no timeout.
	Wall time.Duration
}

// Profile describes a sandbox declaratively and is meant to be reusable across
// many invocations (e.g. one profile for "go build", one for "git").
//
// A profile never grants access to the calling application's data directory
// unless the caller explicitly adds a corresponding Bind, so secrets and the
// database remain structurally unreachable.
type Profile struct {
	// RootFS selects how the sandbox root is assembled.
	RootFS RootFSKind
	// Binds lists host paths exposed inside the sandbox.
	Binds []Bind
	// Env is the explicit environment for the child. The host environment is
	// never inherited automatically.
	Env []string
	// WorkDir is the working directory inside the sandbox. Must be reachable
	// through one of the Binds (or the minimal rootfs).
	WorkDir string
	// Hostname sets the UTS namespace hostname. Defaults to "sandbox".
	Hostname string
	// Net selects the network isolation mode.
	Net NetMode
	// Limits applies resource limits.
	Limits Limits
	// Seccomp selects the syscall policy.
	Seccomp SeccompMode
	// Landlock enables the additional landlock filesystem hardening layer on
	// top of the mount isolation. Enabled by default (the zero value is true
	// via the NewProfile constructor); when constructing a Profile literal set
	// this explicitly.
	Landlock bool
	// AllowNewUserNS permits the sandboxed process to create nested user
	// namespaces. Disabled by default.
	AllowNewUserNS bool
}

// Cmd is a concrete program execution within a [Profile].
type Cmd struct {
	// Path is the absolute path of the executable inside the sandbox.
	Path string
	// Args are the arguments, excluding the program name (Args[0] is set to
	// Path automatically).
	Args []string
	// Stdin, Stdout and Stderr are wired to the child process. Nil means
	// /dev/null for stdin and discard for stdout/stderr.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Result carries the outcome of a sandboxed execution.
type Result struct {
	// ExitCode is the exit code of the sandboxed process. It is -1 if the
	// process was killed by a signal or the sandbox failed before exec.
	ExitCode int
	// TimedOut is true if the process was killed because it exceeded
	// Limits.Wall or the context deadline.
	TimedOut bool
}

// Run executes cmd inside the sandbox described by p and blocks until the
// process exits, the context is cancelled, or the wall-clock timeout elapses.
//
// Run requires that [Init] has been called at the start of main. On linux it
// re-executes the current binary into a locked-down trampoline that applies
// namespaces, mount isolation, landlock and seccomp before exec'ing cmd.Path.
//
// On non-linux platforms Run falls back to an UNSANDBOXED passthrough (see the
// package documentation); this exists purely to allow local development on dev
// hosts and must never be relied upon in production.
func Run(ctx context.Context, p Profile, cmd Cmd) (Result, error) {
	return run(ctx, p, cmd)
}

// Init must be the first thing called in main. In the normal (parent) process
// it is a no-op and returns immediately. When the process was re-executed by
// [Run] as a sandbox trampoline, Init takes over, sets up the sandbox and
// exec's the target program, never returning to the caller.
func Init() {
	initTrampoline()
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsyscall "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"golang.org/x/sys/unix"
)

// initTrampoline is called from Init. In the parent process it is a no-op. In
// the re-executed child it reads the spec, builds the sandbox, applies all
// isolation layers and finally execve's the target program, never returning.
func initTrampoline() {
	if os.Getenv(envInitMode) != "1" {
		return
	}

	// From here on we are the trampoline: on any error we must die loudly
	// rather than continue running untrusted-adjacent code with partial
	// isolation.
	if err := trampoline(); err != nil {
		fmt.Fprintf(os.Stderr, "sbox-init: %v\n", err)
		os.Exit(127)
	}
	// trampoline execs on success and never reaches here.
	os.Exit(127)
}

func trampoline() error {
	// Lock to the OS thread: all the namespace/mount/landlock/seccomp state we
	// set up must belong to the thread that ultimately calls execve.
	runtime.LockOSThread()

	sp, err := readSpec()
	if err != nil {
		return err
	}

	if sp.Isolation == int(IsolationLandlockOnly) {
		return trampolineLandlockOnly(sp)
	}

	if err := setupNet(sp); err != nil {
		return fmt.Errorf("net setup: %w", err)
	}

	if err := setupMounts(sp); err != nil {
		return fmt.Errorf("mount setup: %w", err)
	}

	if err := unix.Sethostname([]byte(sp.Hostname)); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	if err := os.Chdir(sp.WorkDir); err != nil {
		return fmt.Errorf("chdir %q: %w", sp.WorkDir, err)
	}

	if err := applyRLimits(sp.RLimits); err != nil {
		return fmt.Errorf("rlimits: %w", err)
	}

	if sp.Landlock {
		if err := applyLandlock(sp); err != nil {
			return fmt.Errorf("landlock: %w", err)
		}
	}

	// no_new_privs must be set before seccomp and prevents any setuid binary
	// from gaining privileges.
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("no_new_privs: %w", err)
	}

	if sp.Seccomp != int(SeccompOff) {
		if err := applySeccomp(SeccompMode(sp.Seccomp), sp.AllowNewUserNS); err != nil {
			return fmt.Errorf("seccomp: %w", err)
		}
	}

	argv := append([]string{sp.Path}, sp.Args...)
	env := sp.Env
	if env == nil {
		env = []string{}
	}
	if err := unix.Exec(sp.Path, argv, env); err != nil {
		return fmt.Errorf("exec %q: %w", sp.Path, err)
	}
	return nil // unreachable
}

// trampolineLandlockOnly applies the weaker, privilege-free isolation: no
// namespaces, no mounts, no pivot_root. It restricts filesystem access via a
// landlock allowlist over the real (host) paths, sets rlimits, no_new_privs and
// seccomp, then execs. It is used when the calling process runs under a
// hardened confinement that forbids namespace creation and mount.
func trampolineLandlockOnly(sp spec) error {
	if sp.WorkDir != "" && sp.WorkDir != "/" {
		if err := os.Chdir(sp.WorkDir); err != nil {
			return fmt.Errorf("chdir %q: %w", sp.WorkDir, err)
		}
	}

	if err := applyRLimits(sp.RLimits); err != nil {
		return fmt.Errorf("rlimits: %w", err)
	}

	if sp.Landlock {
		if err := applyLandlockAllowlist(sp); err != nil {
			return fmt.Errorf("landlock: %w", err)
		}
	}

	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("no_new_privs: %w", err)
	}

	if sp.Seccomp != int(SeccompOff) {
		if err := applySeccomp(SeccompMode(sp.Seccomp), sp.AllowNewUserNS); err != nil {
			return fmt.Errorf("seccomp: %w", err)
		}
	}

	argv := append([]string{sp.Path}, sp.Args...)
	env := sp.Env
	if env == nil {
		env = []string{}
	}
	if err := unix.Exec(sp.Path, argv, env); err != nil {
		return fmt.Errorf("exec %q: %w", sp.Path, err)
	}
	return nil // unreachable
}

// applyLandlockAllowlist builds a strict landlock allowlist over the real host
// paths. Unlike the namespace-mode landlock (which trusts mount isolation and
// may grant RODirs("/")), this is the ONLY filesystem boundary, so it must
// never grant "/". It exposes only the minimal read-only system paths needed by
// common toolchains plus the explicit binds. Any path not listed here — notably
// the application's data directory — is inaccessible.
func applyLandlockAllowlist(sp spec) error {
	// Landlock is the sole filesystem boundary in this mode. If the kernel does
	// not support it at all, a BestEffort call would silently succeed and leave
	// the process unconfined, so we verify availability first and fail hard.
	if v, err := llsyscall.LandlockGetABIVersion(); err != nil || v < 1 {
		return fmt.Errorf("landlock unavailable (abi=%d, err=%v): cannot enforce landlock-only isolation", v, err)
	}

	var rules []landlock.Rule
	for _, p := range minimalROPaths {
		rules = append(rules, landlock.RODirs(p).IgnoreIfMissing())
	}
	for _, b := range sp.Binds {
		if b.Writable {
			rules = append(rules, landlock.RWDirs(b.Host).IgnoreIfMissing())
		} else {
			rules = append(rules, landlock.RODirs(b.Host).IgnoreIfMissing())
		}
	}
	// A private, writable scratch space. Under a systemd unit with
	// PrivateTmp=yes this /tmp is already service-private.
	rules = append(rules, landlock.RWDirs("/tmp").IgnoreIfMissing())

	// Not BestEffort-optional here: landlock is the sole boundary, so if the
	// kernel cannot enforce it we must fail rather than run unconfined.
	// BestEffort still degrades gracefully across landlock ABI versions.
	return landlock.V4.BestEffort().RestrictPaths(rules...)
}

func readSpec() (spec, error) {
	f := os.NewFile(specFD, "sbox-spec")
	if f == nil {
		return spec{}, fmt.Errorf("spec fd %d not available", specFD)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return spec{}, fmt.Errorf("read spec: %w", err)
	}
	var sp spec
	if err := json.Unmarshal(data, &sp); err != nil {
		return spec{}, fmt.Errorf("unmarshal spec: %w", err)
	}
	return sp, nil
}

// setupMounts builds a fresh root filesystem: make all mounts private, create a
// tmpfs root, bind the requested paths into it, mount a private /proc, minimal
// /dev and /tmp, then pivot_root into it.
func setupMounts(sp spec) error {
	// Make the whole mount tree private so nothing propagates back to the host.
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("make-rprivate: %w", err)
	}

	// Resolve all bind sources to O_PATH file descriptors BEFORE mounting the
	// new root tmpfs. Mounting the tmpfs shadows whatever directory it is
	// placed on (commonly /tmp), which would otherwise hide bind sources that
	// live under that path. Referring to the pre-opened fds via /proc/self/fd
	// makes bind mounting immune to this shadowing.
	if sp.RootFS == int(RootMinimal) {
		for i := range minimalROPaths {
			resolveBindFD(&sp.resolvedBinds, specBind{Host: minimalROPaths[i], Target: minimalROPaths[i], Optional: true})
		}
	}
	for _, b := range sp.Binds {
		if err := resolveBindFD(&sp.resolvedBinds, b); err != nil {
			return err
		}
	}

	// Create a fresh tmpfs to serve as the new root. We mount it at a private
	// directory. Bind sources were already resolved to fds above, so it does
	// not matter whether this shadows their original path.
	newRoot, err := os.MkdirTemp("", ".sbox-root-*")
	if err != nil {
		return fmt.Errorf("mkdir newroot: %w", err)
	}
	if err := unix.Mount("tmpfs", newRoot, "tmpfs", 0, "mode=0755"); err != nil {
		return fmt.Errorf("mount newroot tmpfs: %w", err)
	}

	for _, rb := range sp.resolvedBinds {
		if err := bindResolvedInto(newRoot, rb); err != nil {
			return err
		}
	}

	// Prepare mount points that live inside newRoot.
	for _, d := range []string{"proc", "dev", "tmp"} {
		if err := os.MkdirAll(filepath.Join(newRoot, d), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	// Private /proc (works because we are in a new PID namespace).
	if !bindCovers(sp.resolvedBinds, "/proc") {
		if err := unix.Mount("proc", filepath.Join(newRoot, "proc"), "proc",
			unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC, ""); err != nil {
			return fmt.Errorf("mount proc: %w", err)
		}
	}
	// Minimal /dev as tmpfs with the common device nodes bind-mounted in.
	if !bindCovers(sp.resolvedBinds, "/dev") {
		devPath := filepath.Join(newRoot, "dev")
		if err := unix.Mount("tmpfs", devPath, "tmpfs", unix.MS_NOSUID, "mode=0755"); err != nil {
			return fmt.Errorf("mount dev: %w", err)
		}
		for _, dev := range []string{"null", "zero", "full", "random", "urandom", "tty"} {
			src := "/dev/" + dev
			dst := filepath.Join(devPath, dev)
			if _, err := os.Stat(src); err != nil {
				continue
			}
			f, err := os.OpenFile(dst, os.O_CREATE, 0o666)
			if err != nil {
				return fmt.Errorf("create dev node %s: %w", dev, err)
			}
			f.Close()
			if err := unix.Mount(src, dst, "", unix.MS_BIND, ""); err != nil {
				return fmt.Errorf("bind dev %s: %w", dev, err)
			}
		}
	}
	// Writable /tmp inside the sandbox, unless a bind already provides /tmp or
	// something under it (mounting here would shadow such a bind).
	if !bindTouches(sp.resolvedBinds, "/tmp") {
		if err := unix.Mount("tmpfs", filepath.Join(newRoot, "tmp"), "tmpfs",
			unix.MS_NOSUID|unix.MS_NODEV, "mode=1777"); err != nil {
			return fmt.Errorf("mount sandbox tmp: %w", err)
		}
	}

	// pivot_root: move newRoot to / and the old root to newRoot/.oldroot.
	oldRoot := filepath.Join(newRoot, ".oldroot")
	if err := os.MkdirAll(oldRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir oldroot: %w", err)
	}
	if err := unix.PivotRoot(newRoot, oldRoot); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir /: %w", err)
	}
	// Detach the old root and remove the mount point.
	if err := unix.Unmount("/.oldroot", unix.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount oldroot: %w", err)
	}
	if err := os.Remove("/.oldroot"); err != nil {
		// best effort
		_ = err
	}
	return nil
}

// bindCovers reports whether a bind target equals prefix or is an ancestor of
// it (i.e. the bind provides prefix itself).
func bindCovers(binds []resolvedBind, prefix string) bool {
	for _, b := range binds {
		if b.target == prefix {
			return true
		}
	}
	return false
}

// bindTouches reports whether any bind target equals prefix or lives under it,
// meaning a default mount at prefix would shadow that bind.
func bindTouches(binds []resolvedBind, prefix string) bool {
	p := prefix + "/"
	for _, b := range binds {
		if b.target == prefix || strings.HasPrefix(b.target, p) {
			return true
		}
	}
	return false
}

// resolvedBind is a bind whose source has been opened as an O_PATH fd so it can
// be bind-mounted via /proc/self/fd even after its original path is shadowed.
type resolvedBind struct {
	fd       int
	isDir    bool
	target   string
	writable bool
}

// resolveBindFD opens b.Host as an O_PATH descriptor and appends the resolved
// bind to dst. Missing optional binds are skipped.
func resolveBindFD(dst *[]resolvedBind, b specBind) error {
	fi, err := os.Stat(b.Host)
	if err != nil {
		if b.Optional && os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat bind %q: %w", b.Host, err)
	}
	fd, err := unix.Open(b.Host, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		if b.Optional {
			return nil
		}
		return fmt.Errorf("open bind %q: %w", b.Host, err)
	}
	target := b.Target
	if target == "" {
		target = b.Host
	}
	*dst = append(*dst, resolvedBind{
		fd: fd, isDir: fi.IsDir(), target: target, writable: b.Writable,
	})
	return nil
}

// bindResolvedInto bind-mounts a pre-opened bind source into newRoot+target.
func bindResolvedInto(newRoot string, rb resolvedBind) error {
	target := filepath.Join(newRoot, rb.target)
	if rb.isDir {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("mkdir bind target %q: %w", target, err)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("mkdir bind parent %q: %w", target, err)
		}
		f, err := os.OpenFile(target, os.O_CREATE, 0o644)
		if err != nil {
			return fmt.Errorf("create bind target %q: %w", target, err)
		}
		f.Close()
	}

	src := fmt.Sprintf("/proc/self/fd/%d", rb.fd)
	if err := unix.Mount(src, target, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("bind fd %d -> %q: %w", rb.fd, target, err)
	}
	unix.Close(rb.fd)
	if !rb.writable {
		flags := uintptr(unix.MS_BIND | unix.MS_REMOUNT | unix.MS_RDONLY | unix.MS_REC)
		if err := unix.Mount("", target, "", flags, ""); err != nil {
			return fmt.Errorf("remount ro %q: %w", target, err)
		}
	}
	return nil
}

// minimalROPaths are the common system paths exposed read-only in RootMinimal.
var minimalROPaths = []string{
	"/usr",
	"/bin",
	"/sbin",
	"/lib",
	"/lib64",
	"/etc/ssl",
	"/etc/ca-certificates",
	"/etc/resolv.conf",
	"/etc/nsswitch.conf",
	"/etc/passwd",
	"/etc/group",
}

// setupNet configures networking inside the (possibly new) network namespace.
// For NetHost we inherit the host network and do nothing. For NetLoopback we
// bring the loopback interface up so localhost sockets work. For NetNone we
// leave the namespace empty (not even loopback).
func setupNet(sp spec) error {
	if sp.Net == int(NetLoopback) {
		if err := bringLoopbackUp(); err != nil {
			return err
		}
	}
	return nil
}

// bringLoopbackUp sets the "lo" interface UP inside the current network
// namespace using a raw netlink RTM_NEWLINK request. This avoids any external
// tooling (ip/ifconfig).
func bringLoopbackUp() error {
	ifi, err := netInterfaceByName("lo")
	if err != nil {
		return fmt.Errorf("find lo: %w", err)
	}
	return setLinkUp(ifi)
}

func applyRLimits(l specRLimits) error {
	set := func(res int, val uint64) error {
		if val == 0 {
			return nil
		}
		rl := unix.Rlimit{Cur: val, Max: val}
		return unix.Setrlimit(res, &rl)
	}
	if err := set(unix.RLIMIT_CPU, l.CPUSeconds); err != nil {
		return fmt.Errorf("cpu: %w", err)
	}
	if err := set(unix.RLIMIT_AS, l.AddressSpace); err != nil {
		return fmt.Errorf("as: %w", err)
	}
	if err := set(unix.RLIMIT_NOFILE, l.NoFile); err != nil {
		return fmt.Errorf("nofile: %w", err)
	}
	if err := set(unix.RLIMIT_NPROC, l.NProc); err != nil {
		return fmt.Errorf("nproc: %w", err)
	}
	if err := set(unix.RLIMIT_FSIZE, l.FileSize); err != nil {
		return fmt.Errorf("fsize: %w", err)
	}
	return nil
}

// applyLandlock adds a landlock filesystem layer on top of the mount isolation.
// After pivot_root the visible paths are the bind targets, /tmp, /dev and
// /proc; we grant read-execute on the whole (now minimal) root and read-write
// on the writable binds plus /tmp.
func applyLandlock(sp spec) error {
	rules := []landlock.Rule{
		landlock.RODirs("/").IgnoreIfMissing(),
		landlock.RWDirs("/tmp").IgnoreIfMissing(),
		landlock.RWDirs("/dev").IgnoreIfMissing(),
	}
	for _, b := range sp.Binds {
		if b.Writable {
			rules = append(rules, landlock.RWDirs(b.Target).IgnoreIfMissing())
		}
	}
	// BestEffort so that older kernels degrade gracefully; mount isolation
	// remains the primary guarantee.
	return landlock.V4.BestEffort().RestrictPaths(rules...)
}

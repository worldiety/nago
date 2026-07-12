// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// run is the linux parent implementation. It re-executes the current binary
// into the sandbox trampoline (see initTrampoline) inside a fresh set of
// namespaces, hands over the spec via a pipe, and waits for completion.
func run(ctx context.Context, p Profile, cmd Cmd) (Result, error) {
	if cmd.Path == "" {
		return Result{ExitCode: -1}, errors.New("sbox: cmd.Path is empty")
	}

	if err := checkUserNamespaces(); err != nil {
		return Result{ExitCode: -1}, err
	}

	sp, err := buildSpec(p, cmd)
	if err != nil {
		return Result{ExitCode: -1}, err
	}

	specData, err := json.Marshal(sp)
	if err != nil {
		return Result{ExitCode: -1}, fmt.Errorf("sbox: marshal spec: %w", err)
	}

	// Pipe carrying the spec to the child on specFD.
	pr, pw, err := os.Pipe()
	if err != nil {
		return Result{ExitCode: -1}, fmt.Errorf("sbox: pipe: %w", err)
	}
	defer pr.Close()

	exe, err := os.Executable()
	if err != nil {
		pw.Close()
		return Result{ExitCode: -1}, fmt.Errorf("sbox: resolve executable: %w", err)
	}

	c := exec.Command(exe)
	c.Args = []string{"sbox-init"}
	c.Env = []string{envInitMode + "=1"}
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	c.Stderr = cmd.Stderr
	if c.Stdout == nil {
		c.Stdout = io.Discard
	}
	if c.Stderr == nil {
		c.Stderr = io.Discard
	}
	// pr becomes fd 3 (specFD) in the child; ExtraFiles[0] -> fd 3.
	c.ExtraFiles = []*os.File{pr}

	cloneFlags := uintptr(unix.CLONE_NEWUSER |
		unix.CLONE_NEWNS |
		unix.CLONE_NEWPID |
		unix.CLONE_NEWIPC |
		unix.CLONE_NEWUTS)
	if p.Net != NetHost {
		cloneFlags |= unix.CLONE_NEWNET
	}

	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
		// Map the current uid/gid to root inside the new user namespace so the
		// trampoline can mount, pivot_root and chown within the sandbox.
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
		GidMappingsEnableSetgroups: false,
		// Put the child in its own process group so we can kill the whole tree.
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	if err := c.Start(); err != nil {
		pw.Close()
		return Result{ExitCode: -1}, fmt.Errorf("sbox: start trampoline: %w", err)
	}
	// pr is owned by the child now.
	pr.Close()

	// Write the spec, then close to signal EOF to the child reader.
	if _, err := pw.Write(specData); err != nil {
		pw.Close()
		_ = c.Process.Kill()
		_, _ = c.Process.Wait()
		return Result{ExitCode: -1}, fmt.Errorf("sbox: write spec: %w", err)
	}
	pw.Close()

	return wait(ctx, c, p.Limits.Wall)
}

// wait blocks for the process to finish, enforcing ctx cancellation and the
// optional wall-clock timeout by killing the whole process group.
func wait(ctx context.Context, c *exec.Cmd, wall time.Duration) (Result, error) {
	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	var timer <-chan time.Time
	if wall > 0 {
		t := time.NewTimer(wall)
		defer t.Stop()
		timer = t.C
	}

	pgid := c.Process.Pid

	kill := func() {
		// Negative pid kills the whole process group.
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_ = c.Process.Kill()
	}

	select {
	case err := <-done:
		return result(c, err, false)
	case <-ctx.Done():
		kill()
		<-done
		res, _ := result(c, nil, true)
		return res, ctx.Err()
	case <-timer:
		kill()
		<-done
		res, _ := result(c, nil, true)
		return res, nil
	}
}

func result(c *exec.Cmd, waitErr error, timedOut bool) (Result, error) {
	res := Result{ExitCode: -1, TimedOut: timedOut}
	if c.ProcessState != nil {
		res.ExitCode = c.ProcessState.ExitCode()
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		return res, waitErr
	}
	return res, nil
}

func buildSpec(p Profile, cmd Cmd) (spec, error) {
	binds := make([]specBind, 0, len(p.Binds))
	for _, b := range p.Binds {
		if b.Host == "" {
			return spec{}, errors.New("sbox: bind with empty host path")
		}
		target := b.Target
		if target == "" {
			target = b.Host
		}
		binds = append(binds, specBind{
			Host: b.Host, Target: target, Writable: b.Writable, Optional: b.Optional,
		})
	}

	hostname := p.Hostname
	if hostname == "" {
		hostname = "sandbox"
	}
	workdir := p.WorkDir
	if workdir == "" {
		workdir = "/"
	}

	var cpuSeconds uint64
	if p.Limits.CPUTime > 0 {
		cpuSeconds = uint64(p.Limits.CPUTime / time.Second)
		if cpuSeconds == 0 {
			cpuSeconds = 1
		}
	}

	return spec{
		RootFS:         int(p.RootFS),
		Binds:          binds,
		Env:            p.Env,
		WorkDir:        workdir,
		Hostname:       hostname,
		Net:            int(p.Net),
		AllowNewUserNS: p.AllowNewUserNS,
		Seccomp:        int(p.Seccomp),
		Landlock:       p.Landlock,
		RLimits: specRLimits{
			CPUSeconds:   cpuSeconds,
			AddressSpace: p.Limits.AddressSpace,
			NoFile:       p.Limits.NoFile,
			NProc:        p.Limits.NProc,
			FileSize:     p.Limits.FileSize,
		},
		Path: cmd.Path,
		Args: cmd.Args,
	}, nil
}

// checkUserNamespaces verifies that unprivileged user namespaces are available.
// sbox hard-fails when they are not, rather than degrading isolation.
func checkUserNamespaces() error {
	// A best-effort probe: the definitive check happens at clone time, but this
	// gives a clearer, earlier error on the common Ubuntu/AppArmor lockdown.
	if data, err := os.ReadFile("/proc/sys/kernel/unprivileged_userns_clone"); err == nil {
		if len(data) > 0 && data[0] == '0' {
			return ErrNoUserNamespace
		}
	}
	if data, err := os.ReadFile("/proc/sys/user/max_user_namespaces"); err == nil {
		if len(data) > 0 && data[0] == '0' {
			return ErrNoUserNamespace
		}
	}
	return nil
}

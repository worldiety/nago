// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !linux

package sbox

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

// initTrampoline is a no-op on non-linux platforms. There is no re-exec
// trampoline because there is no sandbox to set up.
func initTrampoline() {}

// run is the non-linux passthrough implementation. It executes the command
// WITHOUT ANY ISOLATION and logs a loud warning on every invocation. This
// exists purely so that a NAGO application remains buildable and runnable on
// developer machines (macOS, Windows). Production always runs on linux, where
// this file is never compiled.
//
// If SBOX_REQUIRE_ISOLATION=1 is set, the passthrough refuses to run and
// returns ErrUnsupported, so environments that must not run unsandboxed can
// enforce that.
func run(ctx context.Context, p Profile, cmd Cmd) (Result, error) {
	if os.Getenv("SBOX_REQUIRE_ISOLATION") == "1" {
		return Result{ExitCode: -1}, ErrUnsupported
	}

	slog.Warn("sbox: running UNSANDBOXED passthrough on a non-linux dev host — "+
		"NO isolation is applied, secrets are NOT protected. Never use in production.",
		slog.String("path", cmd.Path),
		slog.Any("args", cmd.Args),
	)

	c := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
	c.Dir = p.WorkDir
	c.Env = p.Env
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	c.Stderr = cmd.Stderr
	if c.Stdin == nil {
		c.Stdin = nil
	}
	if c.Stdout == nil {
		c.Stdout = io.Discard
	}
	if c.Stderr == nil {
		c.Stderr = io.Discard
	}

	err := c.Run()
	res := Result{ExitCode: c.ProcessState.ExitCode()}
	if ctx.Err() != nil {
		res.TimedOut = true
	}
	var exitErr *exec.ExitError
	if err != nil && !asExitError(err, &exitErr) {
		return res, err
	}
	return res, nil
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

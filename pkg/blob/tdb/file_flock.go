// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd

package tdb

import (
	"os"
	"syscall"
)

// lockFile locks the data file exclusively for the open file and waits for it. Former versions locked the compacted
// file and the log the same way instead of tdb.lock, thus they cannot open a directory at the same time. A compaction
// locks the new files before they replace the former ones.
func lockFile(f *os.File) error {
	conn, err := f.SyscallConn()
	if err != nil {
		return err
	}

	var lockErr error
	err = conn.Control(func(fd uintptr) {
		for {
			lockErr = syscall.Flock(int(fd), syscall.LOCK_EX)
			if lockErr != syscall.EINTR {
				return
			}
		}
	})

	if err == nil && lockErr != nil {
		err = &os.PathError{Op: "flock", Path: f.Name(), Err: lockErr}
	}

	return err
}

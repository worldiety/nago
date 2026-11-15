// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build darwin

package ndb

import (
	"os"
	"syscall"
)

// On macOS fsync(2) moves the data of a file to the drive, but not through the cache of the drive; F_FULLFSYNC does
// that too, for everything the drive holds. SyncAll therefore does an fsync(2) of each file and one F_FULLFSYNC.
const syncBarrier = true

// syncData is fsync(2), without F_FULLFSYNC.
func syncData(f *os.File) error {
	rc, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var serr error
	if err := rc.Control(func(fd uintptr) {
		for {
			if serr = syscall.Fsync(int(fd)); serr != syscall.EINTR {
				return
			}
		}
	}); err != nil {
		return err
	}
	return serr
}

// syncFull is Go's File.Sync, which is F_FULLFSYNC on macOS.
func syncFull(f *os.File) error { return f.Sync() }

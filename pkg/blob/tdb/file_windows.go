// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build windows

package tdb

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// openDataFile opens the file with FILE_SHARE_DELETE, which os.OpenFile does not set. Without it, a file cannot be
// renamed or removed while it is open, but a compaction moves the files away, which readers may still use.
func openDataFile(path string, flag int) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}

	disposition := uint32(syscall.OPEN_EXISTING)
	switch {
	case flag&os.O_CREATE != 0 && flag&os.O_TRUNC != 0:
		disposition = syscall.CREATE_ALWAYS
	case flag&os.O_CREATE != 0:
		disposition = syscall.OPEN_ALWAYS
	case flag&os.O_TRUNC != 0:
		disposition = syscall.TRUNCATE_EXISTING
	}

	h, err := syscall.CreateFile(p, syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, disposition,
		syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}

	return os.NewFile(uintptr(h), path), nil
}

// replaceFile moves src to dst. Windows cannot replace a file, which is still open, thus the former dst is moved
// away first and removed afterwards, which becomes effective when its last handle closes. A crash in between leaves
// dst missing, which the next Open repairs, see recoverFiles.
func replaceFile(dst, src string) error {
	old := dst + oldSuffix
	_ = os.Remove(old)
	if err := os.Rename(dst, old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		if rerr := os.Rename(old, dst); rerr != nil {
			return errors.Join(err, rerr)
		}

		return err
	}

	_ = os.Remove(old)
	return nil
}

// syncDir does nothing, because Windows cannot sync a directory.
func syncDir(dir string) error {
	return nil
}

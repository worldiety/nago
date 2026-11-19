// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !windows

package tdb

import (
	"os"
)

func openDataFile(path string, flag int) (*os.File, error) {
	return os.OpenFile(path, flag, 0666)
}

// replaceFile moves src to dst atomically. Open handles of the former dst keep reading its content.
func replaceFile(dst, src string) error {
	return os.Rename(src, dst)
}

// syncDir makes the renames within the directory durable.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}

	err = d.Sync()
	if cerr := d.Close(); err == nil {
		err = cerr
	}

	return err
}

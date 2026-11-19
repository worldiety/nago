// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !(darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd)

package tdb

import (
	"os"
)

// lockFile does nothing on this platform, see the other implementation. On Windows, the locks of former versions are
// mandatory, thus a data file, which a former version holds, fails to load.
func lockFile(f *os.File) error {
	return nil
}

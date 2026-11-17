// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !darwin

package ndb

import "os"

// Elsewhere a sync of a file is all it takes; SyncAll needs no flush of the drive at the end.
const syncBarrier = false

func syncData(f *os.File) error { return f.Sync() }

func syncFull(f *os.File) error { return f.Sync() }

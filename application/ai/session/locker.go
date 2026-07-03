// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"sync"

	"go.wdy.de/nago/pkg/std/concurrent"
)

// locker hands out one mutex per session id so that mutating operations serialize per session instead of
// globally. This matters because [Append] holds its lock for the whole (potentially long-running) provider
// call - including a blocking agentic ask_user round-trip - and must not block operations on unrelated
// sessions.
//
// Entries are created lazily and intentionally never removed: a *sync.Mutex is tiny, and reclaiming a keyed
// lock safely (without racing a concurrent lock/unlock on the same id) is notoriously error prone. The
// unbounded-growth risk is negligible for the number of distinct sessions a process realistically touches.
type locker struct {
	mutexes concurrent.RWMap[ID, *sync.Mutex]
}

// lock acquires and returns the release function for the given session id. Usage:
//
//	defer l.lock(id)()
func (l *locker) lock(id ID) func() {
	m, _ := l.mutexes.LoadOrStore(id, &sync.Mutex{})
	m.Lock()
	return m.Unlock
}

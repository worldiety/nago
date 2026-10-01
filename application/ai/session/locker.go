// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"

	"go.wdy.de/nago/pkg/std/concurrent"
)

// locker hands out one lock per session id so that mutating operations serialize per session instead of
// globally. This matters because [Append] holds its lock for the whole (potentially long-running) provider
// call and must not block operations on unrelated sessions.
//
// Entries are created lazily and intentionally never removed: a lock is tiny, and reclaiming a keyed lock
// safely (without racing a concurrent lock/unlock on the same id) is notoriously error prone. The
// unbounded-growth risk is negligible for the number of distinct sessions a process realistically touches.
type locker struct {
	locks concurrent.RWMap[ID, chan struct{}]
}

// lock acquires the lock of the given session id and returns its release function. It gives up with the error
// of ctx, if that is done first, e.g. because the user stopped a run which waits for a run of the same
// session in another window. A nil ctx waits forever. Usage:
//
//	release, err := l.lock(ctx, id)
//	if err != nil {
//		return err
//	}
//	defer release()
func (l *locker) lock(ctx context.Context, id ID) (release func(), err error) {
	sem, _ := l.locks.LoadOrStore(id, make(chan struct{}, 1))
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case sem <- struct{}{}:
		return func() { <-sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

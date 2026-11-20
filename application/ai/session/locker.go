// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	"sync"
)

// locker hands out one lock per session id so that mutating operations serialize per session instead of
// globally. This matters because [Append] holds its lock for the whole (potentially long-running) provider
// call and must not block operations on unrelated sessions.
//
// An entry exists only while somebody holds or waits for the lock of its session, thus the locker keeps nothing of
// the sessions it has locked once.
type locker struct {
	mu    sync.Mutex
	locks map[ID]*sessionLock
}

// sessionLock is the lock of a session and the number of its holders and waiters.
type sessionLock struct {
	sem  chan struct{}
	refs int
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
	l.mu.Lock()
	if l.locks == nil {
		l.locks = map[ID]*sessionLock{}
	}

	e := l.locks[id]
	if e == nil {
		e = &sessionLock{sem: make(chan struct{}, 1)}
		l.locks[id] = e
	}
	e.refs++
	l.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case e.sem <- struct{}{}:
		return sync.OnceFunc(func() {
			<-e.sem
			l.drop(id, e)
		}), nil
	case <-ctx.Done():
		l.drop(id, e)
		return nil, ctx.Err()
	}
}

// drop releases the entry of a holder or a waiter.
func (l *locker) drop(id ID, e *sessionLock) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e.refs--
	if e.refs == 0 {
		delete(l.locks, id)
	}
}

// len returns the number of sessions, whose lock is held or awaited.
func (l *locker) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.locks)
}

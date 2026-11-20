// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package blob

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

// InvalidKey is the key, by which [Invalidate] marks a derived store as invalid, see [OpenStoreOptions.Derived].
// While it exists, the owner of the store must not trust the derived data. The owner removes it, once it has rebuilt
// the store from scratch. It sorts behind every key, which starts with a letter, a digit or one of "!-./_~" followed by
// anything but a second "~", so that [DeleteAll], which deletes in ascending order, removes it last.
const InvalidKey = "~~invalid"

// ErrInvalidated is the cause, with which [Invalidate] cancels the context of a running [Derive].
var ErrInvalidated = errors.New("the derived store has been invalidated")

// derivation is the work on a derived store in this process. It exists while a [Derive] or an [Invalidate] of the
// store is in progress, so that no store stays referenced afterwards.
type derivation struct {
	refs int
	// invalidations counts the [Invalidate] calls, which have not ended yet
	invalidations int
	// ended is closed, when the last invalidation has ended
	ended chan struct{}
	// running is closed, when the running Derive has returned, nil without one
	running chan struct{}
	cancel  context.CancelCauseFunc
}

var derivations = struct {
	mu sync.Mutex
	m  map[any]*derivation
}{m: map[any]*derivation{}}

// derivationKey identifies the store within the process: by itself, if it can be compared, otherwise by its name.
func derivationKey(store Store) any {
	if reflect.ValueOf(store).Comparable() {
		return store
	}

	return store.Name()
}

// acquireDerivation returns the derivation of the key with one more reference. The caller holds derivations.mu.
func acquireDerivation(key any) *derivation {
	d := derivations.m[key]
	if d == nil {
		d = &derivation{}
		derivations.m[key] = d
	}

	d.refs++
	return d
}

// releaseDerivation drops a reference of the derivation. The caller holds derivations.mu.
func releaseDerivation(key any, d *derivation) {
	d.refs--
	if d.refs == 0 {
		delete(derivations.m, key)
	}
}

// Derive runs fn, which derives the data of the derived store from other stores, e.g. it rebuilds an index. It runs at
// most one fn per store at a time, and none while an [Invalidate] of the store is in progress: it waits until those
// have ended, or until ctx is done. An Invalidate, which begins while fn runs, cancels the context of fn with the cause
// [ErrInvalidated] and waits, until fn has returned. Thus, fn checks its context before each step, which it must not
// take any more after an invalidation, and returns [context.Cause] of it.
func Derive(ctx context.Context, store Store, fn func(ctx context.Context) error) error {
	key := derivationKey(store)

	for {
		derivations.mu.Lock()
		d := acquireDerivation(key)

		var wait chan struct{}
		switch {
		case d.invalidations > 0:
			wait = d.ended
		case d.running != nil:
			wait = d.running
		}

		if wait != nil {
			releaseDerivation(key, d)
			derivations.mu.Unlock()

			select {
			case <-wait:
				continue
			case <-ctx.Done():
				return context.Cause(ctx)
			}
		}

		fnCtx, cancel := context.WithCancelCause(ctx)
		running := make(chan struct{})
		d.running, d.cancel = running, cancel
		derivations.mu.Unlock()

		err := fn(fnCtx)
		cancel(nil)

		derivations.mu.Lock()
		close(running)
		d.running, d.cancel = nil, nil
		releaseDerivation(key, d)
		derivations.mu.Unlock()

		return err
	}
}

// Invalidate marks the derived store as invalid with [InvalidKey], e.g. before a restore clears the stores, from which
// it is derived. Before, it cancels a running [Derive] of the store and waits until it has returned. From then on, no
// Derive of the store runs until end is called, e.g. at the end of the restore, whether it succeeded or failed. A
// failed mark ends the invalidation at once. Calling end more than once does nothing.
func Invalidate(store Store) (end func(), err error) {
	key := derivationKey(store)

	derivations.mu.Lock()
	d := acquireDerivation(key)
	d.invalidations++
	if d.invalidations == 1 {
		d.ended = make(chan struct{})
	}

	running := d.running
	if d.cancel != nil {
		d.cancel(ErrInvalidated)
	}
	derivations.mu.Unlock()

	if running != nil {
		<-running
	}

	end = sync.OnceFunc(func() {
		derivations.mu.Lock()
		defer derivations.mu.Unlock()

		d.invalidations--
		if d.invalidations == 0 {
			close(d.ended)
		}
		releaseDerivation(key, d)
	})

	if err := Put(store, InvalidKey, []byte(time.Now().UTC().Format(time.RFC3339Nano))); err != nil {
		end()
		return nil, err
	}

	return end, nil
}

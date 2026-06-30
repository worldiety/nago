// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package msgstore

import (
	"log/slog"
	"runtime/debug"
	"sync"

	"go.wdy.de/nago/pkg/ndb"
)

// Notification is the engine-neutral live signal, re-exported so the hook sites
// in Append/Put can construct it without a conversion.
type Notification = ndb.Notification

// notifyRegistry is the set of active subscribers for a DB.
//
// Delivery is synchronous: publish invokes each interested subscriber's callback
// inline (outside the per-type write lock). This keeps the implementation small
// and predictable — no per-subscriber goroutines, channels, or buffering. The
// trade-off is that a slow callback delays the writer, so subscribers must do
// only trivial work in the callback (set a flag, signal a worker) and must not
// write back into the same store. See [ndb.Notifier].
type notifyRegistry struct {
	mu     sync.RWMutex
	subs   map[uint64]*subscriber
	nextID uint64
	closed bool
}

type subscriber struct {
	types map[TypeID]bool // nil/empty ⇒ all types
	fn    func(Notification)
}

func (s *subscriber) wants(typeID TypeID) bool {
	if len(s.types) == 0 {
		return true
	}
	return s.types[typeID]
}

func newNotifyRegistry() *notifyRegistry {
	return &notifyRegistry{subs: make(map[uint64]*subscriber)}
}

func (r *notifyRegistry) subscribe(types []TypeID, fn func(Notification)) (close func()) {
	sub := &subscriber{fn: fn}
	if len(types) > 0 {
		sub.types = make(map[TypeID]bool, len(types))
		for _, t := range types {
			sub.types[t] = true
		}
	}

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return func() {} // registry closed: nothing will ever fire
	}
	id := r.nextID
	r.nextID++
	r.subs[id] = sub
	r.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.subs, id)
			r.mu.Unlock()
		})
	}
}

// publish invokes every interested subscriber synchronously. A panicking
// callback is recovered so it cannot take down the writer.
func (r *notifyRegistry) publish(n Notification) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, sub := range r.subs {
		if sub.wants(n.Type) {
			deliver(sub.fn, n)
		}
	}
}

func deliver(fn func(Notification), n Notification) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("msgstore: panic in notification subscriber", "recover", rec)
			debug.PrintStack()
		}
	}()
	fn(n)
}

// closeAll drops all subscribers. Subsequent subscribe calls return a no-op
// closer.
func (r *notifyRegistry) closeAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	r.subs = make(map[uint64]*subscriber)
}

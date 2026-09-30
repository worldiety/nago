// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"log/slog"
	"slices"
	"time"

	"go.wdy.de/nago/presentation/proto"
)

// A CallbackKey identifies a callback across renders by what it acts on, see [MountKeyedCallback].
//
// The key must name the target of the action, not just the action: a key {"remove", rowID} hits the same row
// after the rows have been reordered, whereas a key {"next", ""} would hit whatever "next" means now. Two
// strings avoid any allocation for the usual case of a literal name and an existing id.
type CallbackKey struct {
	// Name is the action, e.g. "remove", or the column of a cell.
	Name string
	// ID is the target, e.g. the id of a row. It is empty for a singular action of the page, like "save".
	ID string
}

// IsZero reports whether the key is empty.
func (k CallbackKey) IsZero() bool {
	return k == CallbackKey{}
}

// keyedCallbackMounter is implemented by a [RenderContext] which resolves stale calls by key.
type keyedCallbackMounter interface {
	MountKeyedCallback(key CallbackKey, f func()) proto.Ptr
}

// MountKeyedCallback is like [RenderContext.MountCallback], but the callback can be found again by its key.
//
// A callback pointer is only valid for the tree which has been rendered last, so a call from a former tree is
// discarded, e.g. because the window has been rendered again by a ticker, by a background task or by a value
// which an input has sent when it lost its focus right before the click. With a key, such a call is redirected
// to the callback which the current tree mounted under the same key, so the click is not lost, and it always
// runs the newest closure. Nothing is redirected, if the key is gone (e.g. the row has been removed), if it is
// ambiguous, if the former tree is older than [MaxFormerTreeAge], or if the former tree has already been used:
// each tree may cause at most one call after it has been superseded, and none at all, if one of its own calls
// has superseded it, which discards the second click of a double click.
//
// A zero key mounts a plain callback, and a context without support for keys does so as well.
func MountKeyedCallback(ctx RenderContext, key CallbackKey, f func()) proto.Ptr {
	if f == nil {
		return 0
	}

	if key.IsZero() {
		return ctx.MountCallback(f)
	}

	if m, ok := ctx.(keyedCallbackMounter); ok {
		return m.MountKeyedCallback(key, f)
	}

	return ctx.MountCallback(f)
}

const (
	// MaxFormerTrees is how many superseded trees keep the keys of their callbacks, see [MountKeyedCallback].
	MaxFormerTrees = 4
	// MaxFormerTreeAge is how long a superseded tree keeps the keys of its callbacks, see [MountKeyedCallback].
	MaxFormerTreeAge = 2 * time.Second
	// minFormerKeys is the least amount of keys kept for superseded trees, see maxFormerKeys.
	minFormerKeys = 4096
	// selfCallWindow is how long before a tree has been superseded one of its own calls counts as the cause,
	// e.g. the click which changed a state and rendered the next tree. A further call of that tree is then
	// the second click of a double click and never redirected.
	selfCallWindow = 500 * time.Millisecond
)

// formerSegment keeps the keys of a superseded tree, see [MountKeyedCallback]. Its closures are gone.
type formerSegment struct {
	base proto.Ptr
	keys []CallbackKey
	// supersededAt is when the next tree replaced this one.
	supersededAt time.Time
	// calledAt is when a callback of this tree was called last, while it was the current one.
	calledAt time.Time
	// consumed is set once a call of this tree has been redirected.
	consumed bool
}

func (f *formerSegment) contains(ptr proto.Ptr) bool {
	return ptr >= f.base && ptr < f.base+proto.Ptr(len(f.keys))
}

// MountKeyedCallback implements keyedCallbackMounter, see [MountKeyedCallback].
func (s *scopeWindow) MountKeyedCallback(key CallbackKey, f func()) proto.Ptr {
	if f == nil {
		return 0
	}

	ptr := s.callbacks.mountKeyed(key, f)
	s.parent.ids.callback = s.callbacks.next()

	return ptr
}

// resolveCallback returns the callback of ptr: the one of the current tree, or the current callback with the
// same key for a pointer of a former tree, see [MountKeyedCallback]. It returns nil for a stale pointer.
// Only for the event loop.
func (s *scopeWindow) resolveCallback(ptr proto.Ptr, now time.Time) func() {
	if fn := s.callbacks.lookup(ptr); fn != nil {
		s.callbacks.calledAt = now
		return fn
	}

	idx := slices.IndexFunc(s.former, func(f formerSegment) bool { return f.contains(ptr) })
	if idx < 0 {
		return nil
	}

	f := &s.former[idx]
	if now.Sub(f.supersededAt) > MaxFormerTreeAge {
		return nil
	}

	if f.consumed || (!f.calledAt.IsZero() && f.supersededAt.Sub(f.calledAt) <= selfCallWindow) {
		// the second click of a double click, or the tree has been used up
		return nil
	}

	key := f.keys[ptr-f.base]
	if key.IsZero() {
		return nil
	}

	cur, ok := s.callbacks.byKey[key]
	if !ok || cur == 0 {
		// the target is gone or ambiguous
		return nil
	}

	f.consumed = true
	// The client has applied this tree, so the trees before it are of no use any more. The tree itself stays
	// consumed, thus a further call of it is recognized as stale rather than as unknown.
	s.dropFormer(idx)
	s.callbacks.calledAt = now
	slog.Debug("redirected call of a former tree to the current callback of its key", "scope", s.parent.id, "ptr", ptr, "key", key)

	return s.callbacks.lookup(cur)
}

// supersedeCallbacks releases the callbacks of the current tree, which the next tree replaces, and keeps their
// keys for a while, see [MountKeyedCallback]. Only for the event loop.
func (s *scopeWindow) supersedeCallbacks(now time.Time) {
	if s.callbacks.hasKeys {
		keys := s.callbacks.keys
		s.callbacks.keys = s.spareKeys
		s.spareKeys = nil
		s.former = append(s.former, formerSegment{
			base:         s.callbacks.base,
			keys:         keys,
			supersededAt: now,
			calledAt:     s.callbacks.calledAt,
		})
	}

	s.dropCallbacks()
	s.pruneFormer(now)
}

// dropFormer forgets the former trees before idx.
func (s *scopeWindow) dropFormer(idx int) {
	if idx <= 0 {
		return
	}

	for i := range idx {
		clear(s.former[i].keys)
		s.spareKeys = s.former[i].keys[:0]
	}

	s.former = slices.Delete(s.former, 0, idx)
}

// pruneFormer forgets the former trees which are too old or beyond the budget. Only for the event loop.
func (s *scopeWindow) pruneFormer(now time.Time) {
	// right after a tree has been superseded, the next one is still empty, so the superseded one tells the size
	current := len(s.callbacks.keys)
	if n := len(s.former); n > 0 {
		current = max(current, len(s.former[n-1].keys))
	}

	budget := maxFormerKeys(current)
	total := 0
	for _, f := range s.former {
		total += len(f.keys)
	}

	drop := 0
	for drop < len(s.former) {
		f := s.former[drop]
		if now.Sub(f.supersededAt) <= MaxFormerTreeAge && len(s.former)-drop <= MaxFormerTrees && total <= budget {
			break
		}

		total -= len(f.keys)
		drop++
	}

	s.dropFormer(drop)
}

// maxFormerKeys is how many keys the former trees keep at most: twice the keys of the current tree, so a page
// keeps its share whatever its size, but at least minFormerKeys.
func maxFormerKeys(current int) int {
	return max(2*current, minFormerKeys)
}

// forgetFormer drops the keys of all former trees, e.g. because the window has been destroyed.
func (s *scopeWindow) forgetFormer() {
	s.dropFormer(len(s.former))
}

// discardKeys drops the keys of all former trees and of the current one, so that nothing rendered so far is
// redirected any more: the keys have changed their meaning, e.g. because another subject took over the window
// or because the window shows another item of the same route. Only for the event loop.
func (s *scopeWindow) discardKeys() {
	s.forgetFormer()
	s.callbacks.dropKeys()
}

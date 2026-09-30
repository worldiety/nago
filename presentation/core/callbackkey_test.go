// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"testing"
	"time"

	"go.wdy.de/nago/presentation/proto"
)

// keyedWindow is a window without a frontend, whose trees are rendered by hand.
func keyedWindow() *scopeWindow {
	s := &scopeWindow{parent: &Scope{}}
	s.parent.ids.init()
	s.callbacks.restart(s.parent.ids.callback)
	return s
}

func TestKeyedCallbackIsRedirected(t *testing.T) {
	w := keyedWindow()
	now := time.Now()
	key := CallbackKey{Name: "save"}

	calls := ""
	old := w.MountKeyedCallback(key, func() { calls += "old" })
	plain := w.MountCallback(func() { calls += "plain" })

	// e.g. a ticker renders again
	w.supersedeCallbacks(now)
	cur := w.MountKeyedCallback(key, func() { calls += "new" })

	// callbacks without key stay strict
	if w.resolveCallback(plain, now) != nil {
		t.Fatal("a callback without key must not be redirected")
	}

	if fn := w.resolveCallback(old, now); fn == nil {
		t.Fatal("the keyed call of the former tree must be redirected")
	} else {
		fn()
	}

	if calls != "new" {
		t.Fatalf("expected the newest closure, got %q", calls)
	}

	// each former tree redirects at most once
	if w.resolveCallback(old, now) != nil {
		t.Fatal("a former tree must redirect at most once")
	}

	if w.resolveCallback(cur, now) == nil {
		t.Fatal("the current callback must resolve")
	}
}

func TestKeyedCallbackDoubleClickIsDiscarded(t *testing.T) {
	w := keyedWindow()
	now := time.Now()
	key := CallbackKey{Name: "next"}

	ptr := w.MountKeyedCallback(key, func() {})
	// the first click runs and renders the next tree right away
	w.resolveCallback(ptr, now)()
	w.supersedeCallbacks(now.Add(10 * time.Millisecond))
	w.MountKeyedCallback(key, func() {})

	if w.resolveCallback(ptr, now.Add(20*time.Millisecond)) != nil {
		t.Fatal("the second click of a double click must be discarded")
	}
}

func TestKeyedCallbackIsNotRedirectedWhenTargetIsGoneOrAmbiguous(t *testing.T) {
	w := keyedWindow()
	now := time.Now()

	gone := w.MountKeyedCallback(CallbackKey{Name: "remove", ID: "b"}, func() {})
	twice := w.MountKeyedCallback(CallbackKey{Name: "save"}, func() {})
	w.supersedeCallbacks(now)

	w.MountKeyedCallback(CallbackKey{Name: "remove", ID: "c"}, func() {})
	w.MountKeyedCallback(CallbackKey{Name: "save"}, func() {})
	w.MountKeyedCallback(CallbackKey{Name: "save"}, func() {})

	if w.resolveCallback(gone, now) != nil {
		t.Fatal("a key which is gone must not be redirected")
	}

	if w.resolveCallback(twice, now) != nil {
		t.Fatal("an ambiguous key must not be redirected")
	}
}

func TestKeyedCallbackExpires(t *testing.T) {
	w := keyedWindow()
	now := time.Now()
	key := CallbackKey{Name: "save"}

	ptr := w.MountKeyedCallback(key, func() {})
	w.supersedeCallbacks(now)
	w.MountKeyedCallback(key, func() {})

	if w.resolveCallback(ptr, now.Add(MaxFormerTreeAge+time.Second)) != nil {
		t.Fatal("an old tree must not be redirected")
	}

	// an idle window drops the keys of its former trees
	w.pruneFormer(now.Add(MaxFormerTreeAge + time.Second))
	if len(w.former) != 0 {
		t.Fatalf("expected no former trees, got %d", len(w.former))
	}
}

func TestFormerTreesAreBounded(t *testing.T) {
	w := keyedWindow()
	now := time.Now()

	for range MaxFormerTrees + 2 {
		w.MountKeyedCallback(CallbackKey{Name: "save"}, func() {})
		w.supersedeCallbacks(now)
	}

	if len(w.former) != MaxFormerTrees {
		t.Fatalf("expected %d former trees, got %d", MaxFormerTrees, len(w.former))
	}

	// the keys of the former trees are bounded relative to the current tree
	w = keyedWindow()
	for range 3 {
		for i := range 3000 {
			w.MountKeyedCallback(CallbackKey{Name: "row", ID: string(rune('a' + i%26))}, func() {})
		}
		w.supersedeCallbacks(now)
	}
	for range 3000 {
		w.MountKeyedCallback(CallbackKey{Name: "row"}, func() {})
	}
	w.pruneFormer(now)

	total := 0
	for _, f := range w.former {
		total += len(f.keys)
	}
	if total > maxFormerKeys(3000) || len(w.former) != 2 {
		t.Fatalf("expected at most %d keys in 2 trees, got %d in %d", maxFormerKeys(3000), total, len(w.former))
	}

	// a tree without keys costs nothing
	w = keyedWindow()
	w.MountCallback(func() {})
	w.supersedeCallbacks(now)
	if len(w.former) != 0 || len(w.callbacks.keys) != 0 {
		t.Fatal("a tree without keys must not be kept")
	}
}

func TestRedirectDropsOlderTrees(t *testing.T) {
	w := keyedWindow()
	now := time.Now()
	key := CallbackKey{Name: "save"}

	var ptrs []proto.Ptr
	for range 3 {
		ptrs = append(ptrs, w.MountKeyedCallback(key, func() {}))
		w.supersedeCallbacks(now)
	}
	w.MountKeyedCallback(key, func() {})

	// the client applied the second tree, thus the first one is of no use any more
	if w.resolveCallback(ptrs[1], now) == nil {
		t.Fatal("expected a redirect")
	}

	if w.resolveCallback(ptrs[0], now) != nil {
		t.Fatal("an older tree must have been dropped")
	}

	if w.resolveCallback(ptrs[2], now) == nil {
		t.Fatal("a newer tree must still redirect")
	}

	w.forgetFormer()
	if len(w.former) != 0 {
		t.Fatal("expected no former trees")
	}
}

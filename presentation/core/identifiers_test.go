// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"testing"

	"go.wdy.de/nago/presentation/proto"
)

func TestCallbackSegment(t *testing.T) {
	var seg callbackSegment
	seg.restart(100)

	calls := 0
	a := seg.mount(func() { calls += 1 })
	b := seg.mount(func() { calls += 10 })
	if a != 100 || b != 101 || seg.next() != 102 {
		t.Fatalf("unexpected pointers %d %d %d", a, b, seg.next())
	}

	seg.lookup(a)()
	seg.lookup(b)()
	if calls != 11 {
		t.Fatalf("unexpected calls %d", calls)
	}

	for _, ptr := range []proto.Ptr{0, 99, 102, 1 << 52} {
		if seg.lookup(ptr) != nil {
			t.Fatalf("pointer %d must not resolve", ptr)
		}
	}

	// the next tree continues after the former one, whose pointers do not resolve any more
	seg.restart(seg.next())
	if seg.lookup(a) != nil || seg.lookup(b) != nil {
		t.Fatal("pointers of a dropped tree must not resolve")
	}

	if c := seg.mount(func() {}); c != 102 {
		t.Fatalf("expected the next pointer, got %d", c)
	}
}

func TestIdentifiersStartWithinTheirRanges(t *testing.T) {
	for range 100 {
		var ids identifiers
		ids.init()

		if ids.callback < minCallbackPtr || ids.callback >= maxCallbackBase {
			t.Fatalf("callback base %d out of range", ids.callback)
		}

		if ids.state < minStatePtr || ids.state >= maxStateBase {
			t.Fatalf("state base %d out of range", ids.state)
		}

		if a := ids.async.Load(); a < 1 || a >= maxSeqBase {
			t.Fatalf("async base %d out of range", a)
		}

		if ids.file < 1 || ids.file >= maxSeqBase {
			t.Fatalf("file base %d out of range", ids.file)
		}
	}

	// two scopes do not share their bases
	var a, b identifiers
	a.init()
	b.init()
	if a.callback == b.callback || a.state == b.state {
		t.Fatal("expected random bases")
	}
}

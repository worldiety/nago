// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestFilePoolKeepsAHandleOpenWhileItIsUsed: a handle, on which a read, write or stat runs, leaves the pool when another
// path evicts it, by Evict or by Close, but its file stays open until that operation ends, and then it is closed. The
// counter rows: a handle without a running operation is closed at once. The pool has one shard with one handle, so that
// every other path evicts.
func TestFilePoolKeepsAHandleOpenWhileItIsUsed(t *testing.T) {
	for _, c := range []struct {
		name  string
		inUse bool
		evict func(p *FilePool, used, other string) error
	}{
		{"evicted by another path", true, func(p *FilePool, _, other string) error { _, err := p.WriteAt(other, []byte("b"), 0); return err }},
		{"Evict", true, func(p *FilePool, used, _ string) error { p.Evict(used); return nil }},
		{"Close", true, func(p *FilePool, _, _ string) error { return p.Close() }},
		{"counter row: evicted by another path", false, func(p *FilePool, _, other string) error { _, err := p.WriteAt(other, []byte("b"), 0); return err }},
		{"counter row: Evict", false, func(p *FilePool, used, _ string) error { p.Evict(used); return nil }},
		{"counter row: Close", false, func(p *FilePool, _, _ string) error { return p.Close() }},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := NewFilePool(1)
			if len(p.shards) != 1 || p.shards[0].maxOpen != 1 {
				t.Fatalf("premise: %d shards with %d handles, want 1 with 1", len(p.shards), p.shards[0].maxOpen)
			}
			s := p.shards[0]
			dir := t.TempDir()
			used, other := filepath.Join(dir, "a"), filepath.Join(dir, "b")
			s.mu.Lock()
			e, err := s.acquire(used, os.O_RDWR|os.O_CREATE)
			if err == nil && !c.inUse {
				s.release(e) // the operation ended before the eviction
			}
			s.mu.Unlock()
			if err != nil {
				t.Fatal(err)
			}
			if err := c.evict(p, used, other); err != nil {
				t.Fatal(err)
			}
			_, err = e.file.WriteAt([]byte("a"), 0)
			if !c.inUse {
				if !errors.Is(err, os.ErrClosed) {
					t.Fatalf("handle without an operation not closed by the eviction: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("handle closed while its operation runs: %v", err)
			}
			s.mu.Lock()
			s.release(e)
			s.mu.Unlock()
			if _, err := e.file.WriteAt([]byte("a"), 0); !errors.Is(err, os.ErrClosed) {
				t.Fatalf("handle still open after its operation ended: %v", err)
			}
			_ = p.Close()
		})
	}
}

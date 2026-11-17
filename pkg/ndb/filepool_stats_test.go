// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
)

// TestFilePoolStatsCount: the pool counts each opening of a file and the bytes read and written through it. Each row is
// one call on a pool of one handle, so that every other path evicts, and says by how much each counter grows. A cached
// handle opens nothing, the same call on a path that is not cached opens it once: the rows with a growth of 0 stand next
// to their counter row, which shows that the call would have been counted. A read past the end counts what it read, not
// what it asked for. No counter ever shrinks, also not by Evict and Close.
func TestFilePoolStatsCount(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	pool := ndb.NewFilePool(1) // one shard with one handle
	t.Cleanup(func() { _ = pool.Close() })

	type growth struct{ opens, read, written int64 }
	readAt := func(path string, n int, off int64) func() (int, error) {
		return func() (int, error) { return pool.ReadAt(path, make([]byte, n), off) }
	}
	writeAt := func(path string, n int, off int64) func() (int, error) {
		return func() (int, error) { return pool.WriteAt(path, make([]byte, n), off) }
	}
	rows := []struct {
		name    string
		call    func() (int, error)
		wantErr error // io.EOF for a short read; nil otherwise
		want    growth
	}{
		{name: "write of a new file opens it", call: writeAt(a, 100, 0), want: growth{opens: 1, written: 100}},
		{name: "write to the cached file", call: writeAt(a, 50, 100), want: growth{written: 50}},
		{name: "read of the cached file", call: readAt(a, 30, 10), want: growth{read: 30}},
		{name: "another path evicts it and is opened", call: writeAt(b, 7, 0), want: growth{opens: 1, written: 7}},
		{name: "the evicted path is opened again", call: readAt(a, 150, 0), want: growth{opens: 1, read: 150}},
		{name: "a read past the end counts what it read", call: readAt(a, 200, 100), wantErr: io.EOF, want: growth{read: 50}},
		{name: "a read at the end reads nothing", call: readAt(a, 10, 150), wantErr: io.EOF, want: growth{}},
		{name: "Stat of the cached file", call: func() (int, error) { _, err := pool.Stat(a); return 0, err }, want: growth{}},
		{name: "Evict", call: func() (int, error) { pool.Evict(a); return 0, nil }, want: growth{}},
		{name: "after Evict the file is opened again", call: readAt(a, 1, 0), want: growth{opens: 1, read: 1}},
		{name: "Close keeps the counters", call: func() (int, error) { return 0, pool.Close() }, want: growth{}},
	}
	prev := pool.Stats()
	for _, r := range rows {
		_, err := r.call()
		if !errors.Is(err, r.wantErr) {
			t.Fatalf("%s: %v, want %v", r.name, err, r.wantErr)
		}
		got := pool.Stats()
		if got.Opens < prev.Opens || got.ReadBytes < prev.ReadBytes || got.WrittenBytes < prev.WrittenBytes {
			t.Errorf("%s: the counters shrank from %+v to %+v", r.name, prev, got)
		}
		g := growth{opens: got.Opens - prev.Opens, read: got.ReadBytes - prev.ReadBytes, written: got.WrittenBytes - prev.WrittenBytes}
		if g != r.want {
			t.Errorf("%s: the counters grew by %+v, want %+v", r.name, g, r.want)
		}
		prev = got
	}
}

// TestFilePoolStatsConcurrent: readers and writers on many paths of a small pool, which evicts all the time, are each
// counted exactly once. Run with -race.
func TestFilePoolStatsConcurrent(t *testing.T) {
	const workers, perWorker, size = 8, 400, 64
	dir := t.TempDir()
	pool := ndb.NewFilePool(8)
	t.Cleanup(func() { _ = pool.Close() })
	paths := make([]string, 32)
	for i := range paths {
		paths[i] = filepath.Join(dir, "f"+string(rune('a'+i%26))+string(rune('a'+i/26)))
		if err := os.WriteFile(paths[i], make([]byte, size), 0644); err != nil {
			t.Fatal(err)
		}
	}
	before := pool.Stats()
	var wg sync.WaitGroup
	errs := make([]error, workers)
	for w := range workers {
		wg.Go(func() {
			buf := make([]byte, size)
			for i := range perWorker {
				p := paths[(w*perWorker+i)%len(paths)]
				var err error
				if i%2 == 0 {
					_, err = pool.ReadAt(p, buf, 0)
				} else {
					_, err = pool.WriteAt(p, buf, 0)
				}
				if err != nil {
					errs[w] = err
					return
				}
			}
		})
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatal(err)
	}
	got := pool.Stats()
	wantEach := int64(workers * perWorker / 2 * size)
	if r, w := got.ReadBytes-before.ReadBytes, got.WrittenBytes-before.WrittenBytes; r != wantEach || w != wantEach {
		t.Errorf("read %d and written %d bytes, want %d each", r, w, wantEach)
	}
	// every call opens its path at most once; with 32 paths in 8 handles nearly every call opens one
	if o := got.Opens - before.Opens; o < int64(len(paths)) || o > workers*perWorker {
		t.Errorf("%d openings for %d calls on %d paths", o, workers*perWorker, len(paths))
	}
}

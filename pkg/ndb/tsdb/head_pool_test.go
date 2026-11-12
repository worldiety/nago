// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
)

// openDescriptors counts the open file descriptors of the process. The listing itself holds one, in every count alike.
func openDescriptors(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("no /dev/fd: %v", err)
	}
	return len(entries)
}

// TestOpenColumnsStayWithinTheFilePool: the write-ahead log of the head of every column goes through the shared
// FilePool like the chunks, so that 500 open columns, each with a correction in its head, hold at most the descriptors
// of the pool, also after a reopen that replays every head. The counter row with a pool larger than the columns shows
// that the count sees the descriptors of the pool; against 39b9dd36e, which opened every head.wal with os.OpenFile next
// to the pool, the row with the small pool held one descriptor per column.
func TestOpenColumnsStayWithinTheFilePool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no /dev/fd")
	}
	const columns, slack = 500, 16
	for _, c := range []struct {
		name     string
		pool     int
		atMost   int // more open descriptors than before, at most; -1 for no bound
		atLeast  int // at least; 0 for no bound, since the garbage collector may close files of earlier tests meanwhile
		reopened bool
	}{
		{"pool of 32", 32, 32 + slack, 0, false},
		{"pool of 32, reopened", 32, 32 + slack, 0, true},
		{"counter row: pool of 1024", 1024, -1, columns, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			pool := ndb.NewFilePool(c.pool)
			defer pool.Close()
			opts := Options{FilePool: pool}
			db, err := Open(dir, opts)
			if err != nil {
				t.Fatal(err)
			}
			before := openDescriptors(t)
			write := func(db *DB) {
				for i := range columns {
					col, err := db.Column("b", fmt.Sprintf("c%04d", i), Schema{Scheme: SchemeDecimal})
					if err != nil {
						t.Fatal(err)
					}
					if err := col.PutI64(10, 1); err != nil {
						t.Fatal(err)
					}
					if err := col.PutI64(20, 2); err != nil {
						t.Fatal(err)
					}
					if err := col.PutI64(10, -1); err != nil { // a correction: into the head and its WAL
						t.Fatal(err)
					}
				}
			}
			check := func(db *DB) {
				for i := range columns {
					col, err := db.Column("b", fmt.Sprintf("c%04d", i), Schema{Scheme: SchemeDecimal})
					if err != nil {
						t.Fatal(err)
					}
					ts, vs := collectI64(t, col, 0, 30)
					if len(ts) != 2 || vs[0] != -1 || vs[1] != 2 {
						t.Fatalf("column %d: %v %v, want [10 20] [-1 2]", i, ts, vs)
					}
				}
			}
			write(db)
			if c.reopened {
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
				if db, err = Open(dir, opts); err != nil {
					t.Fatal(err)
				}
			}
			check(db)
			grown := openDescriptors(t) - before
			if c.atMost >= 0 && grown > c.atMost {
				t.Errorf("%d open columns hold %d more descriptors, want at most %d with a pool of %d", columns, grown, c.atMost, c.pool)
			}
			if c.atLeast > 0 && grown < c.atLeast {
				t.Errorf("%d more descriptors, want at least %d: the count does not see the pool", grown, c.atLeast)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestHeadWritesSurviveEvictionByTheFilePool: four columns, whose heads share a pool of one descriptor, write corrections
// at the same time, so that each write evicts the handle of the other, also between the lookup in the pool and the
// write. Every correction lands and survives a reopen, because the pool keeps a handle open while a write runs on it. It
// checks the outcome under whatever interleaving the scheduler picks; against the pool of 39b9dd36e, which closed a
// handle under a running write, it failed with "file already closed" in 9 of 10 runs on an M1 Max.
func TestHeadWritesSurviveEvictionByTheFilePool(t *testing.T) {
	const n, columns = 1500, 4
	dir := t.TempDir()
	pool := ndb.NewFilePool(1)
	defer pool.Close()
	opts := Options{FilePool: pool}
	db, err := Open(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	cols := make([]*Column, columns)
	for i := range cols {
		if cols[i], err = db.Column("b", fmt.Sprintf("c%d", i), Schema{Scheme: SchemeDecimal}); err != nil {
			t.Fatal(err)
		}
		if err := cols[i].PutI64(n+1, 0); err != nil {
			t.Fatal(err)
		}
	}
	errs := make([]error, len(cols))
	done := make(chan int, len(cols))
	for i, col := range cols {
		go func() {
			defer func() { done <- i }()
			for ts := int64(1); ts <= n; ts++ {
				if err := col.PutI64(ts, -ts); err != nil { // older than the newest point: a correction in the head
					errs[i] = err
					return
				}
			}
		}()
	}
	for range cols {
		<-done
	}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("column %d: %v", i, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if db, err = Open(dir, opts); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := range cols {
		col, err := db.Column("b", fmt.Sprintf("c%d", i), Schema{Scheme: SchemeDecimal})
		if err != nil {
			t.Fatal(err)
		}
		ts, vs := collectI64(t, col, 1, n)
		if len(ts) != n {
			t.Fatalf("column %d: %d corrections after the reopen, want %d", i, len(ts), n)
		}
		for j := range ts {
			if vs[j] != -ts[j] {
				t.Fatalf("column %d: %d at %d, want %d", i, vs[j], ts[j], -ts[j])
			}
		}
	}
}

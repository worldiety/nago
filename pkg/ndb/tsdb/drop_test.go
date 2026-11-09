// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// dropOptions seals a chunk after 8 points in blocks of 4, so a column of points 0, 10, 20... has the chunks 0..70,
// 80..150 and so on.
func dropOptions() Options {
	return Options{BlockPoints: 4, Split: SplitByCount(8)}
}

func openDropDB(t *testing.T, dir string) *DB {
	t.Helper()
	db, err := Open(dir, dropOptions())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func decimalColumn(t *testing.T, db *DB, bucket, name string) *Column {
	t.Helper()
	c, err := db.Column(bucket, name, Schema{Scheme: SchemeDecimal})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func putEvery10(t *testing.T, c *Column, from, n int) {
	t.Helper()
	for i := range n {
		ts := int64(from+i) * 10
		if err := c.PutI64(ts, ts+1); err != nil {
			t.Fatal(err)
		}
	}
}

func every10(from, to int) []int64 {
	var out []int64
	for i := from; i < to; i++ {
		out = append(out, int64(i)*10)
	}
	return out
}

func chunkFiles(t *testing.T, c *Column) []string {
	t.Helper()
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), chunkExt) {
			names = append(names, e.Name())
		}
	}
	return names
}

func checkPoints(t *testing.T, c *Column, want []int64) {
	t.Helper()
	got, vals := collectI64(t, c, minInt64, maxInt64)
	if !slices.Equal(got, want) {
		t.Fatalf("points:\n got %v\nwant %v", got, want)
	}
	for i, ts := range got {
		if vals[i] != ts+1 {
			t.Fatalf("value of %d is %d", ts, vals[i])
		}
	}
}

// checkPointSet is checkPoints for a column with a point in the head, which a read returns behind the sealed points.
func checkPointSet(t *testing.T, c *Column, want []int64) {
	t.Helper()
	got, vals := collectI64(t, c, minInt64, maxInt64)
	byTime := make(map[int64]int64, len(got))
	for i, ts := range got {
		byTime[ts] = vals[i]
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("points:\n got %v\nwant %v", got, want)
	}
	for _, ts := range got {
		if byTime[ts] != ts+1 {
			t.Fatalf("value of %d is %d", ts, byTime[ts])
		}
	}
}

// TestDropBefore places the limit at each border of each kind of chunk. The state after it must also survive more
// points and a restart.
func TestDropBefore(t *testing.T) {
	const allFlushed, openPending = true, false
	for _, tt := range []struct {
		name    string
		points  int
		flushed bool
		cut     int64
		removed int
		from    int // index of the first remaining point, points means none
		files   int
	}{
		// finalized chunks 0..70, 80..150 and 160..190
		{name: "before all points", points: 20, flushed: allFlushed, cut: -1, removed: 0, from: 0, files: 3},
		{name: "at the first point", points: 20, flushed: allFlushed, cut: 0, removed: 0, from: 0, files: 3},
		{name: "in the first chunk", points: 20, flushed: allFlushed, cut: 30, removed: 0, from: 0, files: 3},
		{name: "at the newest point of a chunk", points: 20, flushed: allFlushed, cut: 70, removed: 0, from: 0, files: 3},
		{name: "after the newest point of a chunk", points: 20, flushed: allFlushed, cut: 71, removed: 1, from: 8, files: 2},
		{name: "at the first point of the next chunk", points: 20, flushed: allFlushed, cut: 80, removed: 1, from: 8, files: 2},
		{name: "after two chunks", points: 20, flushed: allFlushed, cut: 151, removed: 2, from: 16, files: 1},
		{name: "at the newest point of all", points: 20, flushed: allFlushed, cut: 190, removed: 2, from: 16, files: 1},
		{name: "after all points", points: 20, flushed: allFlushed, cut: 191, removed: 3, from: 20, files: 0},

		// the last chunk is pending with 4 points in a sealed block
		{name: "a pending chunk at its newest point stays", points: 20, flushed: openPending, cut: 190, removed: 2, from: 16, files: 1},
		{name: "a pending chunk after its newest point goes", points: 20, flushed: openPending, cut: 191, removed: 3, from: 20, files: 0},
		{name: "a pending chunk with a newer point stays whole", points: 20, flushed: openPending, cut: 175, removed: 2, from: 16, files: 1},

		// the last chunk is pending with its 2 points only in memory
		{name: "points in memory at the newest stay", points: 18, flushed: openPending, cut: 170, removed: 2, from: 16, files: 1},
		{name: "points in memory after the newest go", points: 18, flushed: openPending, cut: 171, removed: 3, from: 18, files: 0},
		{name: "points in memory with a newer one stay whole", points: 18, flushed: openPending, cut: 165, removed: 2, from: 16, files: 1},

		// the first pending chunk with a block on disk and a point in memory
		{name: "a block and points in memory go together", points: 5, flushed: openPending, cut: 41, removed: 1, from: 5, files: 0},
		{name: "a block and a newer point in memory stay together", points: 5, flushed: openPending, cut: 40, removed: 0, from: 0, files: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDropDB(t, dir)
			c := decimalColumn(t, db, "b", "c")
			putEvery10(t, c, 0, tt.points)
			if tt.flushed {
				if err := c.Flush(); err != nil {
					t.Fatal(err)
				}
			}

			removed, err := c.DropBefore(tt.cut)
			if err != nil {
				t.Fatal(err)
			}
			if removed != tt.removed {
				t.Errorf("removed: got %d, want %d", removed, tt.removed)
			}
			want := every10(tt.from, tt.points)
			checkPoints(t, c, want)
			if got := len(chunkFiles(t, c)); got != tt.files {
				t.Errorf("files: got %d %v, want %d", got, chunkFiles(t, c), tt.files)
			}
			if st := c.Stats(); st.Chunks > tt.files {
				t.Errorf("stats: %+v", st)
			}

			// appending goes on behind everything, which ever existed
			putEvery10(t, c, 100, 11)
			want = append(want, every10(100, 111)...)
			checkPoints(t, c, want)

			// an older point than the column has ever seen is still accepted
			if err := c.PutI64(5, 6); err != nil {
				t.Fatal(err)
			}
			withOld := append([]int64{5}, want...)
			slices.Sort(withOld)
			checkPointSet(t, c, withOld)

			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			again := openDropDB(t, dir)
			checkPoints(t, decimalColumn(t, again, "b", "c"), withOld)
		})
	}
}

// TestDropBeforeAndTheHead shows what happens to a change, which is not in a chunk yet, next to the same without one.
func TestDropBeforeAndTheHead(t *testing.T) {
	const cut = 151
	for _, tt := range []struct {
		name   string
		change func(t *testing.T, c *Column)
		flush  bool
		want   []int64
	}{
		{name: "no change", change: func(*testing.T, *Column) {}, want: every10(16, 20)},
		{name: "a late point stays in the head", change: func(t *testing.T, c *Column) {
			if err := c.PutI64(5, 6); err != nil {
				t.Fatal(err)
			}
		}, want: append([]int64{5}, every10(16, 20)...)},
		{name: "a flushed late point is an old chunk", flush: true, change: func(t *testing.T, c *Column) {
			if err := c.PutI64(5, 6); err != nil {
				t.Fatal(err)
			}
		}, want: every10(16, 20)},
		{name: "a deletion of a dropped point changes nothing", change: func(t *testing.T, c *Column) {
			if err := c.Delete(90); err != nil {
				t.Fatal(err)
			}
		}, want: every10(16, 20)},
		{name: "a deletion of a kept point stays", change: func(t *testing.T, c *Column) {
			if err := c.Delete(170); err != nil {
				t.Fatal(err)
			}
		}, want: []int64{160, 180, 190}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := openDropDB(t, t.TempDir())
			c := decimalColumn(t, db, "b", "c")
			putEvery10(t, c, 0, 20)
			if err := c.Flush(); err != nil {
				t.Fatal(err)
			}
			tt.change(t, c)
			if tt.flush {
				if err := c.Flush(); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := c.DropBefore(cut); err != nil {
				t.Fatal(err)
			}
			checkPointSet(t, c, tt.want)
			if err := c.Flush(); err != nil {
				t.Fatal(err)
			}
			if _, err := c.DropBefore(cut); err != nil {
				t.Fatal(err)
			}
			// after the flush a late point is in a chunk of its own, which is older than the limit
			got, _ := collectI64(t, c, minInt64, maxInt64)
			if len(got) == 0 || got[0] < 160 {
				t.Errorf("a flushed late point must go with the next drop: %v", got)
			}
		})
	}
}

// TestDropBeforeReadsOnlyNames takes the rights of a chunk away, so that opening it would fail.
func TestDropBeforeReadsOnlyNames(t *testing.T) {
	db := openDropDB(t, t.TempDir())
	c := decimalColumn(t, db, "b", "c")
	putEvery10(t, c, 0, 20)
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}
	c.pool.Evict(c.chunks[0].path)
	if err := os.Chmod(c.chunks[0].path, 0); err != nil {
		t.Fatal(err)
	}
	removed, err := c.DropBefore(71)
	if err != nil || removed != 1 {
		t.Fatalf("removed %d, %v", removed, err)
	}
	checkPoints(t, c, every10(8, 20))
}

// TestDropBeforeWithConcurrentWriters lets each writer also drop and read in between, so that all three interleave at
// a fixed cost. Only what is older than the limit may be gone. Run it with -race.
func TestDropBeforeWithConcurrentWriters(t *testing.T) {
	const writers, each, step = 4, 1000, 64
	for _, tt := range []struct {
		name string
		drop bool
	}{
		{name: "without dropping"},
		{name: "with dropping", drop: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// larger chunks than in the other tests, because a file per few points would dominate the time
			db, err := Open(t.TempDir(), Options{BlockPoints: 32, Split: SplitByCount(128)})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			c := decimalColumn(t, db, "b", "c")

			var next, cut, removed atomic.Int64
			dropOld := func() error {
				limit := next.Load() - 256
				for {
					old := cut.Load()
					if limit <= old || cut.CompareAndSwap(old, limit) {
						break
					}
				}
				n, err := c.DropBefore(cut.Load())
				removed.Add(int64(n))
				return err
			}
			readAll := func() error {
				seen := map[int64]bool{}
				var twice error
				err := c.ScanI64(minInt64, maxInt64, func(ts, _ []int64) bool {
					for _, v := range ts {
						if seen[v] {
							twice = fmt.Errorf("the point %d was read twice", v)
							return false
						}
						seen[v] = true
					}
					return true
				})
				return errors.Join(err, twice)
			}

			errs := make(chan error, writers)
			var wg sync.WaitGroup
			for range writers {
				wg.Go(func() {
					for i := range each {
						ts := next.Add(1)
						err := c.PutI64(ts, ts+1)
						if err == nil && i%step == 0 {
							err = readAll()
							if err == nil && tt.drop {
								err = dropOld()
							}
						}
						if err != nil {
							errs <- err
							return
						}
					}
				})
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				t.Fatal(err)
			}
			if err := c.Flush(); err != nil {
				t.Fatal(err)
			}

			// without dropping everything is there, which shows that the limit is what removes points
			limit := max(cut.Load(), 1)
			got, vals := collectI64(t, c, limit, maxInt64)
			if int64(len(got)) != writers*each-limit+1 {
				t.Fatalf("got %d points from %d, want %d", len(got), limit, writers*each-limit+1)
			}
			for i, ts := range got {
				if ts != limit+int64(i) || vals[i] != ts+1 {
					t.Fatalf("point %d: %d=%d", i, ts, vals[i])
				}
			}
			if tt.drop && removed.Load() == 0 {
				t.Error("nothing was dropped, so this proves nothing")
			}
		})
	}
}

func TestDropColumnAndBucket(t *testing.T) {
	type dropFn func(db *DB) error
	dropColumn := func(db *DB) error { return db.DropColumn("b", "c") }
	dropBucket := func(db *DB) error { return db.DropBucket("b") }

	for _, tt := range []struct {
		name         string
		drop         dropFn
		reopen       bool // the column exists only on disk
		otherColumn  bool // another column of the bucket survives
		bucketGone   bool
		wantDirGone  bool
		wantOldFails bool
	}{
		{name: "open column", drop: dropColumn, otherColumn: true, wantDirGone: true, wantOldFails: true},
		{name: "column only on disk", drop: dropColumn, reopen: true, otherColumn: true, wantDirGone: true},
		{name: "open bucket", drop: dropBucket, bucketGone: true, wantDirGone: true, wantOldFails: true},
		{name: "bucket only on disk", drop: dropBucket, reopen: true, bucketGone: true, wantDirGone: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDropDB(t, dir)
			c := decimalColumn(t, db, "b", "c")
			putEvery10(t, c, 0, 20)
			putEvery10(t, decimalColumn(t, db, "b", "keep"), 0, 3)
			putEvery10(t, decimalColumn(t, db, "other", "c"), 0, 3)
			if tt.reopen {
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
				db = openDropDB(t, dir)
			}

			if err := tt.drop(db); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(dir, "b", "c")); !os.IsNotExist(err) {
				t.Errorf("the column is still on disk: %v", err)
			}
			if _, ok, err := db.LookupColumn("b", "c"); ok || err != nil {
				t.Errorf("lookup: %v %v", ok, err)
			}

			// an old handle reports it, instead of writing to files of a new column
			if tt.wantOldFails {
				for name, err := range map[string]error{
					"put":    c.PutI64(10_000, 1),
					"delete": c.Delete(5),
					"flush":  c.Flush(),
					"sync":   c.Sync(),
					"scan":   c.ScanI64(minInt64, maxInt64, func(_, _ []int64) bool { return true }),
					"drop":   dropErr(c),
				} {
					if !errors.Is(err, errColumnGone) {
						t.Errorf("%s of an old handle: %v", name, err)
					}
				}
			}

			// the neighbours are untouched
			checkPoints(t, decimalColumn(t, db, "other", "c"), every10(0, 3))
			if tt.otherColumn {
				checkPoints(t, decimalColumn(t, db, "b", "keep"), every10(0, 3))
			}
			if tt.bucketGone {
				if _, err := os.Stat(filepath.Join(dir, "b")); !os.IsNotExist(err) {
					t.Errorf("the bucket is still on disk: %v", err)
				}
			}

			// the same name starts empty, even with another schema, and is written to new files
			fresh, err := db.Column("b", "c", Schema{Scheme: SchemeString})
			if err != nil {
				t.Fatal(err)
			}
			if ts, _ := collectStrings(t, fresh, minInt64, maxInt64); len(ts) != 0 {
				t.Fatalf("a new column must be empty: %v", ts)
			}
			if err := fresh.PutString(7, "seven"); err != nil {
				t.Fatal(err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			again := openDropDB(t, dir)
			reread, err := again.Column("b", "c", Schema{Scheme: SchemeString})
			if err != nil {
				t.Fatal(err)
			}
			if ts, vs := collectStrings(t, reread, minInt64, maxInt64); len(ts) != 1 || vs[0] != "seven" {
				t.Fatalf("after a restart: %v %v", ts, vs)
			}
		})
	}
}

func dropErr(c *Column) error {
	_, err := c.DropBefore(1)
	return err
}

// TestDropKeepsNewColumnApart reuses the name of a dropped column, while a handle of the old one still tries to write.
// The new value of an enum would go to the dictionary, which is a file of the new column now.
func TestDropKeepsNewColumnApart(t *testing.T) {
	text := func(ts int64) string { return fmt.Sprint("value-", ts) }
	readStrings := func(t *testing.T, c *Column) []string { _, vs := collectStrings(t, c, minInt64, maxInt64); return vs }
	for _, tt := range []struct {
		name   string
		schema Schema
		write  func(c *Column, ts int64) error
		read   func(t *testing.T, c *Column) []string
	}{
		{name: "decimal", schema: Schema{Scheme: SchemeDecimal}, write: func(c *Column, ts int64) error { return c.PutI64(ts, ts) },
			read: func(t *testing.T, c *Column) []string {
				_, vs := collectI64(t, c, minInt64, maxInt64)
				out := make([]string, len(vs))
				for i, v := range vs {
					out[i] = text(v)
				}
				return out
			}},
		{name: "text", schema: Schema{Scheme: SchemeString}, write: func(c *Column, ts int64) error { return c.PutString(ts, text(ts)) }, read: readStrings},
		{name: "enum", schema: Schema{Scheme: SchemeEnum}, write: func(c *Column, ts int64) error { return c.PutString(ts, text(ts)) }, read: readStrings},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			db := openDropDB(t, dir)
			old, err := db.Column("b", "c", tt.schema)
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.write(old, 1); err != nil {
				t.Fatal(err)
			}
			if err := db.DropColumn("b", "c"); err != nil {
				t.Fatal(err)
			}
			fresh, err := db.Column("b", "c", tt.schema)
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.write(fresh, 2); err != nil {
				t.Fatal(err)
			}
			if err := tt.write(old, 3); !errors.Is(err, errColumnGone) {
				t.Fatalf("a write of the old handle: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			again := openDropDB(t, dir)
			c, err := again.Column("b", "c", tt.schema)
			if err != nil {
				t.Fatal(err)
			}
			if got := tt.read(t, c); !slices.Equal(got, []string{"value-2"}) {
				t.Fatalf("points: %q", got)
			}
		})
	}
}

func TestDropOfNothing(t *testing.T) {
	db := openDropDB(t, t.TempDir())
	for name, tt := range map[string]struct {
		err  error
		want error
	}{
		"an unknown column is no error": {err: db.DropColumn("b", "unknown")},
		"an unknown bucket is no error": {err: db.DropBucket("unknown")},
		"an invalid bucket name":        {err: db.DropBucket("../outside"), want: errBadName},
		"an invalid column name":        {err: db.DropColumn("b", "../outside"), want: errBadName},
	} {
		if !errors.Is(tt.err, tt.want) {
			t.Errorf("%s: %v", name, tt.err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.DropColumn("b", "c"); !errors.Is(err, errClosed) {
		t.Errorf("a closed database: %v", err)
	}
}

// TestDropColumnWithConcurrentUse keeps using a column while it is dropped again and again. Run it with -race.
func TestDropColumnWithConcurrentUse(t *testing.T) {
	for _, tt := range []struct {
		name string
		drop func(db *DB) error
	}{
		{name: "column", drop: func(db *DB) error { return db.DropColumn("b", "c") }},
		{name: "bucket", drop: func(db *DB) error { return db.DropBucket("b") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := openDropDB(t, t.TempDir())
			schema := Schema{Scheme: SchemeDecimal}

			stop := make(chan struct{})
			errs := make(chan error, 4)
			var ok atomic.Int64
			var wg sync.WaitGroup
			for w := range 2 {
				wg.Go(func() {
					var ts int64
					for {
						select {
						case <-stop:
							return
						default:
						}
						ts++
						c, err := db.Column("b", "c", schema)
						if err != nil {
							errs <- err
							return
						}
						// losing a point is what a drop is for, but anything else than the report of it is a defect
						if err := c.PutI64(ts*2+int64(w), ts); err != nil {
							if !errors.Is(err, errColumnGone) {
								errs <- err
								return
							}
							continue
						}
						ok.Add(1)
					}
				})
			}
			for range 50 {
				if err := tt.drop(db); err != nil {
					t.Fatal(err)
				}
			}
			for ok.Load() == 0 {
				select {
				case err := <-errs:
					t.Fatal(err)
				default:
				}
			}
			close(stop)
			wg.Wait()
			select {
			case err := <-errs:
				t.Fatal(err)
			default:
			}

			c, err := db.Column("b", "c", schema)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.PutI64(1<<40, 1); err != nil {
				t.Fatal(err)
			}
			if err := c.Flush(); err != nil {
				t.Fatal(err)
			}
			if ts, _ := collectI64(t, c, 1<<40, 1<<40); len(ts) != 1 {
				t.Fatalf("after the drops: %v", ts)
			}
		})
	}
}

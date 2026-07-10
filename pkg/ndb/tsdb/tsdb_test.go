// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/timeseries"
)

func timeUTC() *time.Location { return time.UTC }

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(t.TempDir(), Options{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// collectI64 reads all points in range into slices.
func collectI64(t *testing.T, c *Column, min, max int64) ([]int64, []int64) {
	t.Helper()
	var ts, vs []int64
	err := c.ScanI64(min, max, func(bt []int64, bv []int64) bool {
		ts = append(ts, bt...)
		vs = append(vs, bv...)
		return true
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	return ts, vs
}

func collectStrings(t *testing.T, c *Column, min, max int64) ([]int64, []string) {
	t.Helper()
	var ts []int64
	var vs []string
	err := c.ScanString(min, max, func(bt []int64, bv []string) bool {
		ts = append(ts, bt...)
		vs = append(vs, bv...)
		return true
	})
	if err != nil {
		t.Fatalf("scan strings: %v", err)
	}
	return ts, vs
}

func TestDecimalWriteReadFromHead(t *testing.T) {
	db := openTestDB(t)
	c, err := db.Column("plant1", "temp", Schema{Scheme: SchemeDecimal, Decimals: 2})
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < 100; i++ {
		if err := c.PutF64(1000+i*20, 20.0+float64(i)/100); err != nil {
			t.Fatal(err)
		}
	}
	ts, vs := collectI64(t, c, 0, 1<<62)
	if len(ts) != 100 {
		t.Fatalf("want 100 points, got %d", len(ts))
	}
	if vs[0] != 2000 || vs[99] != 2099 {
		t.Fatalf("values wrong: first=%d last=%d", vs[0], vs[99])
	}
}

func TestOverwriteNewestWins(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	c.PutI64(500, 1)
	c.PutI64(500, 2)
	c.PutI64(500, 3)
	ts, vs := collectI64(t, c, 0, 1000)
	if len(ts) != 1 || vs[0] != 3 {
		t.Fatalf("overwrite failed: ts=%v vs=%v", ts, vs)
	}
}

func TestDeletePoint(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	c.PutI64(1, 10)
	c.PutI64(2, 20)
	c.PutI64(3, 30)
	if err := c.Delete(2); err != nil {
		t.Fatal(err)
	}
	ts, vs := collectI64(t, c, 0, 100)
	if len(ts) != 2 || ts[0] != 1 || ts[1] != 3 || vs[1] != 30 {
		t.Fatalf("delete failed: ts=%v vs=%v", ts, vs)
	}
}

func TestFlushPersistsAndReopens(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 1})
	for i := int64(0); i < 5000; i++ {
		c.PutI64(1000+i*10, 100+i%7)
	}
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}
	// after flush, head must be empty and data live in chunks
	if c.head.len() != 0 {
		t.Fatalf("head not reset: %d", c.head.len())
	}
	ts, vs := collectI64(t, c, 0, 1<<62)
	if len(ts) != 5000 {
		t.Fatalf("want 5000 after flush, got %d", len(ts))
	}
	if vs[7] != 100 {
		t.Fatalf("value corrupted after flush: %d", vs[7])
	}
	db.Close()

	// reopen and verify persistence
	db2, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	c2, ok, err := db2.LookupColumn("b", "c")
	if err != nil || !ok {
		t.Fatalf("lookup after reopen: ok=%v err=%v", ok, err)
	}
	ts2, _ := collectI64(t, c2, 0, 1<<62)
	if len(ts2) != 5000 {
		t.Fatalf("persistence lost: got %d", len(ts2))
	}
}

func TestBulkRewriteAfterFlush(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	for i := int64(0); i < 3000; i++ {
		c.PutI64(i, i)
	}
	c.Flush()
	// bulk overwrite an old range that now lives in sealed chunks
	for i := int64(1000); i < 1500; i++ {
		c.PutI64(i, i*10)
	}
	c.Flush()
	ts, vs := collectI64(t, c, 0, 1<<62)
	if len(ts) != 3000 {
		t.Fatalf("count changed after rewrite: %d", len(ts))
	}
	// spot check rewritten and untouched values
	idx := map[int64]int64{}
	for i := range ts {
		idx[ts[i]] = vs[i]
	}
	if idx[1200] != 12000 {
		t.Fatalf("rewrite not applied: %d", idx[1200])
	}
	if idx[2000] != 2000 {
		t.Fatalf("untouched value changed: %d", idx[2000])
	}
}

func TestDeleteRangeAcrossChunks(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	for i := int64(0); i < 2000; i++ {
		c.PutI64(i, i)
	}
	c.Flush()
	if err := c.DeleteRange(500, 1499); err != nil {
		t.Fatal(err)
	}
	c.Flush()
	ts, _ := collectI64(t, c, 0, 1<<62)
	if len(ts) != 1000 {
		t.Fatalf("want 1000 after range delete, got %d", len(ts))
	}
	for _, x := range ts {
		if x >= 500 && x <= 1499 {
			t.Fatalf("deleted ts %d still present", x)
		}
	}
}

func TestEnumColumn(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "state", Schema{Scheme: SchemeEnum})
	states := []string{"idle", "run", "run", "error", "idle"}
	for i, s := range states {
		if err := c.PutString(int64(i), s); err != nil {
			t.Fatal(err)
		}
	}
	c.Flush()
	ts, vs := collectStrings(t, c, 0, 100)
	if len(ts) != 5 {
		t.Fatalf("want 5, got %d", len(ts))
	}
	for i, s := range states {
		if vs[i] != s {
			t.Fatalf("enum[%d]=%q want %q", i, vs[i], s)
		}
	}
}

func TestStringColumn(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "msg", Schema{Scheme: SchemeString})
	for i := int64(0); i < 300; i++ {
		c.PutString(i, fmt.Sprintf("event-%d", i))
	}
	c.Flush()
	ts, vs := collectStrings(t, c, 0, 1000)
	if len(ts) != 300 || vs[42] != "event-42" {
		t.Fatalf("string column wrong: n=%d v42=%q", len(ts), vs[42])
	}
}

func TestSchemeMismatch(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	if err := c.PutString(1, "x"); err == nil {
		t.Fatal("expected scheme mismatch writing string to decimal column")
	}
	if err := c.ScanString(0, 1, func([]int64, []string) bool { return true }); err == nil {
		t.Fatal("expected scheme mismatch scanning strings")
	}
}

func TestColumnSchemaConflict(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 3}); err == nil {
		t.Fatal("expected conflict on differing decimals")
	}
	if _, err := db.Column("b", "c", Schema{Scheme: SchemeEnum}); err == nil {
		t.Fatal("expected conflict on differing scheme")
	}
}

func TestM4Composition(t *testing.T) {
	db := openTestDB(t)
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	for i := int64(0); i < 1000; i++ {
		c.PutI64(i, i%50)
	}
	rng := timeseries.NewRange(0, 999, timeUTC())
	var n int
	for range timeseries.M4(c.IterI64(0, 1000), rng, 10) {
		n++
	}
	if n == 0 {
		t.Fatal("M4 produced no points")
	}
	if n > 40 {
		t.Fatalf("M4 produced too many points: %d", n)
	}
}

// TestReadSeesUnflushedPendingChunk proves that a read observes data the append
// fast path has already sealed into the pending chunk file but not yet finalized
// (i.e. without an explicit Flush). During a large monotonic burst most data
// lives in the pending chunk; it must be visible to reads immediately.
func TestReadSeesUnflushedPendingChunk(t *testing.T) {
	db := openTestDB(t)
	c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	if err != nil {
		t.Fatal(err)
	}
	const n = 20000 // >> BlockPoints, so many blocks seal to the pending chunk
	for i := int64(0); i < n; i++ {
		if err := c.PutI64(1000+i*20, i); err != nil {
			t.Fatal(err)
		}
	}
	// no Flush: read must still see all n points (pending chunk + buffer).
	ts, vs := collectI64(t, c, 0, 1<<62)
	if int64(len(ts)) != n {
		t.Fatalf("read saw %d points without Flush, want %d", len(ts), n)
	}
	for i := int64(0); i < n; i++ {
		if ts[i] != 1000+i*20 || vs[i] != i {
			t.Fatalf("point %d = (%d,%d), want (%d,%d)", i, ts[i], vs[i], 1000+i*20, i)
		}
	}

	// a bounded range read across the pending chunk must be correct too.
	lo, hi := int64(1000+5000*20), int64(1000+15000*20)
	rts, _ := collectI64(t, c, lo, hi)
	if len(rts) != 10001 {
		t.Fatalf("range read saw %d points, want 10001", len(rts))
	}
	if rts[0] != lo || rts[len(rts)-1] != hi {
		t.Fatalf("range read bounds = [%d,%d], want [%d,%d]", rts[0], rts[len(rts)-1], lo, hi)
	}
}

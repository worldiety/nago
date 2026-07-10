// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHeadTornTailRecovery writes points to the head WAL, appends garbage
// (simulating a torn write), and verifies the column reopens ignoring the torn
// tail while keeping all valid records.
func TestHeadTornTailRecovery(t *testing.T) {
	dir := t.TempDir()
	db, _ := Open(dir, Options{})
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	for i := int64(0); i < 50; i++ {
		c.PutI64(i, i)
	}
	db.Close()

	walPath := filepath.Join(dir, "b", "c", headWALName)
	f, err := os.OpenFile(walPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal(err)
	}
	// append a partial/garbage record that cannot form a valid frame
	f.Write([]byte{0x00, 0xFF, 0xFF, 0xFF, 0x7F, 0xDE, 0xAD})
	f.Close()

	db2, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	c2, ok, err := db2.LookupColumn("b", "c")
	if err != nil || !ok {
		t.Fatalf("reopen: ok=%v err=%v", ok, err)
	}
	ts, _ := collectI64(t, c2, 0, 1000)
	if len(ts) != 50 {
		t.Fatalf("torn tail recovery: got %d points, want 50", len(ts))
	}
}

// TestChunkBlockCorruptionSkipped corrupts one block in a sealed chunk and
// verifies the reader forward-scans past it and still returns the other blocks.
func TestChunkBlockCorruptionSkipped(t *testing.T) {
	dir := t.TempDir()
	db, _ := Open(dir, Options{BlockPoints: 100})
	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})
	for i := int64(0); i < 500; i++ {
		c.PutI64(i, i)
	}
	c.Flush()
	chunkPath := c.chunks[0].path
	db.Close()

	// flip bytes in the middle of the file to corrupt one block body
	data, err := os.ReadFile(chunkPath)
	if err != nil {
		t.Fatal(err)
	}
	mid := len(data) / 2
	for i := mid; i < mid+8 && i < len(data); i++ {
		data[i] ^= 0xFF
	}
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	db2, _ := Open(dir, Options{BlockPoints: 100})
	defer db2.Close()
	c2, _, _ := db2.LookupColumn("b", "c")
	ts, _ := collectI64(t, c2, 0, 1000)
	// with 5 blocks of 100, corrupting one loses at most ~100 points but must
	// not lose everything and must not error.
	if len(ts) == 0 {
		t.Fatal("corruption lost all data")
	}
	if len(ts) == 500 {
		t.Log("note: corruption did not hit a block body (acceptable)")
	}
	if len(ts) < 300 {
		t.Fatalf("lost too much: %d points", len(ts))
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"io"
	"math/rand"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test_wal_set(t *testing.T) {
	testfname := filepath.Join(t.TempDir(), "test.WAL")
	f, err := OpenFile(testfname)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	type entryt struct {
		b, k, v []byte
	}

	wal, err := NewWAL(f, nil)
	if err != nil {
		t.Error(err)
	}
	rnd := rand.New(rand.NewSource(1234))
	var testset []entryt
	for i := range 10_000 {
		b := make([]byte, rnd.Intn(64))
		k := make([]byte, rnd.Intn(128))
		v := make([]byte, rnd.Intn(16*1024))
		rnd.Read(b)
		rnd.Read(k)
		rnd.Read(v)
		testset = append(testset, entryt{b, k, v})

		entry := Node{kind: setKeyValue, tx: uint64(i), bucket: b[:], key: k[:], val: v[:]}
		if _, err := wal.write(&entry); err != nil {
			t.Error(err)
		}
	}

	const reOpenTest = true

	if reOpenTest {
		wal, err = OpenWAL(testfname, nil)
		if err != nil {
			t.Error(err)
		}
	}

	idx := 0

	for entry, err := range wal.All() {
		if err != nil {
			t.Error(err)
		}

		expected := testset[idx]
		if !reflect.DeepEqual(expected.b, entry.bucket) {
			t.Fatalf("got %v\nwant %v", entry.bucket, expected.b)
		}

		if !reflect.DeepEqual(expected.k, entry.key) {
			t.Fatalf("got %v\nwant %v", entry.key, expected.k)
		}

		if !reflect.DeepEqual(expected.v, entry.val) {
			t.Fatalf("got %v\nwant %v", entry.val, expected.v)
		}

		valPtr := entry.Value()
		r := valPtr.NewReader()
		tmp, err := io.ReadAll(r)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(expected.v, tmp) {
			t.Fatalf("got %v\nwant %v", tmp, expected.v)
		}

		idx++
	}

	if idx != len(testset) {
		t.Errorf("failed to range over entries: expected %d, got %d", len(testset), idx)
	}

}

func Benchmark_wal_set(t *testing.B) {
	f, err := OpenFile(filepath.Join(t.TempDir(), "test.WAL"))
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	wal, err := NewWAL(f, nil)
	if err != nil {
		t.Error(err)
	}
	var b [16]byte
	var k [32]byte
	var v [512]byte
	rnd := rand.New(rand.NewSource(1234))
	rnd.Read(b[:])
	rnd.Read(k[:])
	rnd.Read(v[:])

	start := time.Now()
	writes := 0
	for i := 0; i < t.N; i++ {
		writes++
		entry := Node{kind: setKeyValue, tx: uint64(i), bucket: b[:], key: k[:], val: v[:]}
		if _, err := wal.write(&entry); err != nil {
			t.Error(err)
		}
	}

	t.Logf("wrote %d entries in %v\n", writes, time.Now().Sub(start))
}

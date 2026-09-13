// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type TestEntry struct {
	Bucket string
	Key    string
	Value  []byte
}

func TestDB_Bench(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	expectedSet := makeTestSet()
	start := time.Now()
	for _, entry := range expectedSet {
		if err := db.Set(entry.Bucket, entry.Key, entry.Value); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("written %d entries in %v\n", len(expectedSet), time.Since(start))

	start = time.Now()
	for _, entry := range expectedSet {
		optReader := db.Get(entry.Bucket, entry.Key)
		if optReader.IsNone() {
			t.Fatal("missing entry")
		}

		reader := optReader.Unwrap()
		_, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Logf("read %d entries in %v\n", len(expectedSet), time.Since(start))
}

func TestDB_Set(t *testing.T) {
	dbdir := filepath.Join(t.TempDir())
	db, err := Open(dbdir)
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	expectedSet := makeTestSet()
	for _, entry := range expectedSet {
		if err := db.Set(entry.Bucket, entry.Key, entry.Value); err != nil {
			t.Fatal(err)
		}

		optReader := db.Get(entry.Bucket, entry.Key)
		if optReader.IsNone() {
			t.Fatal("missing entry")
		}

		reader := optReader.Unwrap()
		buf, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf, entry.Value) {
			t.Fatalf("mismatched value")
		}
	}

	var entries []TestEntry
	for bucket := range db.Buckets() {
		for entry := range db.All(bucket) {
			buf, err := io.ReadAll(entry.val.NewReader())
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, TestEntry{
				Bucket: bucket,
				Key:    entry.key,
				Value:  buf,
			})
		}
	}

	if len(entries) != len(expectedSet) {
		t.Fatalf("mismatched number of entries")
	}

	sort(entries)
	sort(expectedSet)

	if !reflect.DeepEqual(entries, expectedSet) {
		t.Fatalf("mismatched entries")
	}

	// close and re-read
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(dbdir)
	if err != nil {
		t.Fatal(err)
	}

	entries = nil
	for bucket := range db.Buckets() {
		for entry := range db.All(bucket) {
			buf, err := io.ReadAll(entry.val.NewReader())
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, TestEntry{
				Bucket: bucket,
				Key:    entry.key,
				Value:  buf,
			})
		}
	}

	sort(entries)

	if !reflect.DeepEqual(entries, expectedSet) {
		t.Fatalf("mismatched entries")
	}

	// close and re-read
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(dbdir)
	if err != nil {
		t.Fatal(err)
	}

	entries = nil
	for bucket := range db.Buckets() {
		for entry := range db.All(bucket) {
			buf, err := io.ReadAll(entry.val.NewReader())
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, TestEntry{
				Bucket: bucket,
				Key:    entry.key,
				Value:  buf,
			})
		}
	}

	sort(entries)

	if !reflect.DeepEqual(entries, expectedSet) {
		t.Fatalf("mismatched entries")
	}
}

func TestDB_5m(t *testing.T) {
	// Five million writes with a read back after each one takes well over twenty minutes,
	// which exceeds the default go test timeout and makes a plain "go test ./..." unusable.
	// It never actually ran before, because a leaked handle panicked the binary long
	// beforehand. Keep it opt in, following the convention of the other expensive tests.
	if os.Getenv("NAGO_TDB_STRESS") == "" {
		t.Skip("set NAGO_TDB_STRESS=1 to run the five million entry stress test")
	}

	dbdir := filepath.Join(t.TempDir())
	db, err := Open(dbdir)
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	const maxEntries = 5_000_000
	bucket := "nums"
	for no := range maxEntries {
		key := strconv.Itoa(no)
		val := []byte(key)

		if err := db.Set(bucket, key, val); err != nil {
			t.Fatal(err)
		}

		optReader := db.Get(bucket, key)
		if optReader.IsNone() {
			t.Fatal("missing entry")
		}

		reader := optReader.Unwrap()
		buf, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf, val) {
			t.Fatalf("mismatched value")
		}
	}

	// close and re-read
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(dbdir)
	if err != nil {
		t.Fatal(err)
	}

	for no := range maxEntries {
		key := strconv.Itoa(no)
		val := []byte(key)
		optReader := db.Get(bucket, key)
		buf, err := io.ReadAll(optReader.Unwrap())
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, val) {
			t.Fatalf("mismatched value")
		}
	}

}

func sort(entries []TestEntry) {
	slices.SortFunc(entries, func(a, b TestEntry) int {
		if i := strings.Compare(a.Bucket, b.Bucket); i != 0 {
			return i
		}
		if i := strings.Compare(a.Key, b.Key); i != 0 {
			return i
		}
		return bytes.Compare(a.Value, b.Value)
	})
}

func makeTestSet() []TestEntry {
	var res []TestEntry
	r := rand.New(rand.NewSource(1234))

	for bidx := range 10 {
		b := []byte("bucket-" + strconv.Itoa(bidx))

		for kidx := range 10_000 {
			k := []byte("key-" + strconv.Itoa(kidx))
			v := make([]byte, r.Intn(1024*16))
			r.Read(v)

			res = append(res, TestEntry{
				Bucket: string(b),
				Key:    string(k),
				Value:  v,
			})

		}
	}

	return res
}

func TestDB_Races(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	expectedSet := makeTestSet()

	const threads = 10
	var wg sync.WaitGroup
	for range threads {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for _, entry := range expectedSet {
				if err := db.Set(entry.Bucket, entry.Key, entry.Value); err != nil {
					panic(err)
				}
			}
		}()
	}

	wg.Wait()

	var entries []TestEntry
	for bucket := range db.Buckets() {
		for entry := range db.All(bucket) {
			buf, err := io.ReadAll(entry.val.NewReader())
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, TestEntry{
				Bucket: bucket,
				Key:    entry.key,
				Value:  buf,
			})
		}
	}

	if len(entries) != len(expectedSet) {
		t.Fatalf("mismatched number of entries")
	}

	sort(entries)
	sort(expectedSet)

	if !reflect.DeepEqual(entries, expectedSet) {
		t.Fatalf("mismatched entries")
	}
}

// closeDB closes whichever database the given variable currently refers to.
//
// The tests reopen the database several times and only closed some of those handles. An
// unclosed handle is not merely untidy here: lockedfile installs a finalizer which panics when
// the file becomes unreachable without a Close, so the leak killed the whole test binary from
// inside the garbage collector, during whatever unrelated test happened to be running.
//
// A close error is deliberately ignored, because on the path where a test already failed the
// handle may have been closed explicitly beforehand.
func closeDB(db **DB) {
	if *db != nil {
		_ = (*db).Close()
	}
}

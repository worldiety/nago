// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"fmt"
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

// TestDB_ConcurrentReadWrite must be run with the race detector: snapshots of the iterators must not race with
// writers or with other snapshots of the same bucket.
func TestDB_ConcurrentReadWrite(t *testing.T) {
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	const n = 200
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range n {
			key := fmt.Sprintf("k%04d", i)
			if err := db.Set("runs", key, []byte("v")); err != nil {
				t.Error(err)
				return
			}

			if i%3 == 0 {
				if err := db.Delete("runs", key); err != nil {
					t.Error(err)
					return
				}
			}
		}
	})

	for range 4 {
		wg.Go(func() {
			for range n {
				for range db.Ascend("runs") {
				}
				for range db.Descend("runs") {
				}
				for range db.AscendRange("runs", "k0000", "k0100") {
				}
				for range db.DescendRange("runs", "k0000", "k0100") {
				}
			}
		})
	}

	wg.Wait()
}

// TestDB_CompactWithConcurrentWrites covers the Compact path where a write lands after the snapshot was taken, so
// the WAL must be kept and replayed on top of the new compacted file. That path used to re-open the WAL while its
// old handle was still open. The exclusive flock of lockedfile belongs to the open file description, so Open
// blocked forever on it, while Compact held every lock of the DB.
func TestDB_CompactWithConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	const bucket = "b"
	want := map[string]string{}
	for i := range 200 {
		key := fmt.Sprintf("old%05d", i)
		want[key] = "value of " + key
		if err := db.Set(bucket, key, []byte(want[key])); err != nil {
			t.Fatal(err)
		}
	}

	// The writer only adds fresh keys and deletes old ones, because neither has to read a stored value.
	next := 0
	write := func() error {
		key := fmt.Sprintf("new%06d", next)
		if err := db.Set(bucket, key, []byte(key)); err != nil {
			return err
		}
		want[key] = key

		if next%2 == 0 && next/2 < 200 {
			key := fmt.Sprintf("old%05d", next/2)
			if err := db.Delete(bucket, key); err != nil {
				return err
			}
			delete(want, key)
		}

		next++
		return nil
	}

	// A compaction without a write in that window removes the WAL instead, thus retry a few times. The ballast
	// makes the snapshot take a moment and tells both paths apart: a kept WAL still contains it, a new one cannot
	// grow that large from the few writes between the return of Compact and the stop of the writer.
	walFile := filepath.Join(dir, "tdb.wal")
	ballast := bytes.Repeat([]byte("x"), 1<<20)
	keptWAL := false
	for attempt := 0; attempt < 10 && !keptWAL; attempt++ {
		if err := db.Set("ballast", strconv.Itoa(attempt), ballast); err != nil {
			t.Fatal(err)
		}

		info, err := os.Stat(walFile)
		if err != nil {
			t.Fatal(err)
		}
		walSizeBefore := info.Size()

		stop := make(chan struct{})
		var wg sync.WaitGroup
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}

				if err := write(); err != nil {
					t.Error(err)
					return
				}
			}
		})

		compacted := make(chan error, 1)
		go func() { compacted <- db.Compact() }()

		select {
		case err := <-compacted:
			close(stop)
			wg.Wait()
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(30 * time.Second):
			// Compact still holds the locks of the DB, thus closing it would block as well.
			db = nil
			t.Fatal("Compact did not return, it blocks on re-opening the WAL")
		}

		info, err = os.Stat(walFile)
		if err != nil {
			t.Fatal(err)
		}
		keptWAL = info.Size() >= walSizeBefore
	}

	if !keptWAL {
		t.Fatal("no write landed between snapshot and file swap, the kept WAL path is not covered")
	}

	check := func(stage string) {
		t.Helper()
		if got := db.Len(bucket); got != len(want) {
			t.Fatalf("%s: expected %d entries, got %d", stage, len(want), got)
		}

		for key, val := range want {
			optReader := db.Get(bucket, key)
			if optReader.IsNone() {
				t.Fatalf("%s: missing entry %s", stage, key)
			}

			buf, err := io.ReadAll(optReader.Unwrap())
			if err != nil {
				t.Fatal(err)
			}

			if string(buf) != val {
				t.Fatalf("%s: mismatched value of %s: %q", stage, key, buf)
			}
		}
	}

	check("after compaction")

	// close and re-read, which replays the kept WAL once more
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	check("after re-open")
}

// openFiles counts the file descriptors of this process. This catches a leaked lockedfile right away, instead of by
// its finalizer panic at some random later point.
func openFiles(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("cannot count open files: %v", err)
	}

	return len(entries)
}

// TestDB_CompactCleansUpOnError checks that a failed compaction neither leaves its temp file open nor on disk.
func TestDB_CompactCleansUpOnError(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	if err := db.Set("b", "k", []byte("v")); err != nil {
		t.Fatal(err)
	}

	// the value lives in the WAL, thus a closed WAL fails the snapshot while it copies the value
	_ = db.wal.f.Close()

	before := openFiles(t)
	if err := db.Compact(); err == nil {
		t.Fatal("expected Compact to fail")
	}

	if after := openFiles(t); after != before {
		t.Fatalf("Compact left %d files open", after-before)
	}

	tmpFiles, err := filepath.Glob(filepath.Join(dir, "*.compact.tmp"))
	if err != nil {
		t.Fatal(err)
	}

	if len(tmpFiles) != 0 {
		t.Fatalf("Compact left its temp files behind: %v", tmpFiles)
	}
}

// TestDB_OpenCleansUpOnError checks that a failed Open leaves no file open and releases the directory, so that it
// can be opened again once the cause is gone.
func TestDB_OpenCleansUpOnError(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Set("b", "k", []byte("v")); err != nil {
		t.Fatal(err)
	}

	// move the entry into the compacted file, so that it survives the corrupted WAL below
	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	// a payload length beyond the limit fails the replay, when the compacted file is already open
	walFile := filepath.Join(dir, "tdb.wal")
	if err := os.WriteFile(walFile, []byte{0xff, 0xff, 0xff, 0xff, 0xff}, 0600); err != nil {
		t.Fatal(err)
	}

	before := openFiles(t)
	if _, err := Open(dir); err == nil {
		t.Fatal("expected Open to fail on the corrupted WAL")
	}

	if after := openFiles(t); after != before {
		t.Fatalf("Open left %d files open", after-before)
	}

	if err := os.Remove(walFile); err != nil {
		t.Fatal(err)
	}

	db, err = Open(dir)
	if err != nil {
		t.Fatalf("cannot open again after a failed Open: %v", err)
	}

	defer closeDB(&db)

	if !db.Exists("b", "k") {
		t.Fatal("entry of the compacted file is missing")
	}
}

// TestDB_CloseCleansUpOnError checks that Close still closes the compacted file and releases the directory, when
// the WAL fails.
func TestDB_CloseCleansUpOnError(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	// a closed WAL fails to sync and to close
	_ = db.wal.f.Close()

	before := openFiles(t)
	if err := db.Close(); err == nil {
		t.Fatal("expected Close to report the failed WAL")
	}

	if after := openFiles(t); after != before-1 {
		t.Fatal("Close left the compacted file open")
	}

	db, err = Open(dir)
	if err != nil {
		t.Fatalf("cannot open again after a failed Close: %v", err)
	}

	defer closeDB(&db)
}

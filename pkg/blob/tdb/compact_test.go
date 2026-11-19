// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testValue encodes the key and a version, so that a reader can tell, whether it got a value of the key.
func testValue(key string, version uint64, size int) []byte {
	v := make([]byte, 8+len(key)+size)
	binary.LittleEndian.PutUint64(v, version)
	copy(v[8:], key)
	return v
}

func checkTestValue(key string, v []byte) error {
	if len(v) < 8+len(key) || string(v[8:8+len(key)]) != key {
		return fmt.Errorf("value of %s belongs to another key: %q", key, v[:min(len(v), 8+len(key)+8)])
	}

	return nil
}

// TestDB_OnlineCompaction runs writers, readers, iterators, explicit and automatic compactions at the same time and
// compares the result with a model. Run it with the race detector.
func TestDB_OnlineCompaction(t *testing.T) {
	dir := t.TempDir()
	// a low limit lets automatic compactions run between the explicit ones
	db, err := OpenWithOptions(dir, Options{GCPercent: 20, MinGarbage: 64 * 1024})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	const writers = 3
	const keysPerWriter = 300
	const writesPerWriter = 6000
	models := make([]map[string]uint64, writers)
	var stop atomic.Bool
	var wg sync.WaitGroup
	var writersWG sync.WaitGroup
	errs := make(chan error, 100)
	report := func(err error) {
		select {
		case errs <- err:
		default:
		}
		stop.Store(true)
	}

	for w := range writers {
		models[w] = map[string]uint64{}
		writersWG.Go(func() {
			r := rand.New(rand.NewSource(int64(w)))
			model := models[w]
			for i := range writesPerWriter {
				if stop.Load() {
					return
				}

				key := fmt.Sprintf("w%d-%04d", w, r.Intn(keysPerWriter))
				bucket := fmt.Sprintf("bucket-%d", r.Intn(3))
				key = bucket + "/" + key
				if _, ok := model[key]; ok && r.Intn(5) == 0 {
					if err := db.Delete(bucket, key); err != nil {
						report(err)
						return
					}
					delete(model, key)
					continue
				}

				version := uint64(i + 1)
				if err := db.Set(bucket, key, testValue(key, version, r.Intn(2000))); err != nil {
					report(err)
					return
				}
				model[key] = version
			}
		})
	}

	// readers check, that a value belongs to its key, also when they read it slowly
	for rd := range 3 {
		wg.Go(func() {
			r := rand.New(rand.NewSource(int64(100 + rd)))
			for !stop.Load() {
				bucket := fmt.Sprintf("bucket-%d", r.Intn(3))
				key := fmt.Sprintf("%s/w%d-%04d", bucket, r.Intn(writers), r.Intn(keysPerWriter))
				opt := db.Get(bucket, key)
				if opt.IsNone() {
					continue
				}

				reader := opt.Unwrap()
				head := make([]byte, 3)
				if _, err := io.ReadFull(reader, head); err != nil {
					report(fmt.Errorf("read %s: %w", key, err))
					return
				}

				if r.Intn(10) == 0 {
					time.Sleep(time.Millisecond)
				}

				rest, err := io.ReadAll(reader)
				if err != nil {
					report(fmt.Errorf("read %s: %w", key, err))
					return
				}

				if err := checkTestValue(key, append(head, rest...)); err != nil {
					report(err)
					return
				}
			}
		})
	}

	// iterators check the order and the values of a snapshot, also while compactions replace the files
	for it := range 2 {
		wg.Go(func() {
			r := rand.New(rand.NewSource(int64(200 + it)))
			for !stop.Load() {
				bucket := fmt.Sprintf("bucket-%d", r.Intn(3))
				var last string
				for entry := range db.Ascend(bucket) {
					if entry.key <= last {
						report(fmt.Errorf("iterator not ascending: %q after %q", entry.key, last))
						return
					}

					last = entry.key
					if r.Intn(20) == 0 {
						buf, err := io.ReadAll(entry.val.NewReader())
						if err != nil {
							report(fmt.Errorf("iterated value of %s: %w", entry.key, err))
							return
						}

						if err := checkTestValue(entry.key, buf); err != nil {
							report(err)
							return
						}

						time.Sleep(100 * time.Microsecond)
					}
				}
			}
		})
	}

	explicit := 0
	wg.Go(func() {
		for !stop.Load() {
			if err := db.compact(explicit%2 == 1); err != nil {
				report(fmt.Errorf("compact: %w", err))
				return
			}
			explicit++
			// leave room for automatic compactions
			time.Sleep(50 * time.Millisecond)
		}
	})

	writersWG.Wait()
	stop.Store(true)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	stats := db.Stats()
	if explicit < 2 || stats.Compactions+stats.LogCompactions <= uint64(explicit) || stats.LogCompactions == 0 {
		t.Fatalf("expected explicit and automatic compactions of both kinds, got %d explicit of %d and %d of the log",
			explicit, stats.Compactions, stats.LogCompactions)
	}

	want := map[string]uint64{}
	for _, m := range models {
		for k, v := range m {
			want[k] = v
		}
	}

	check := func(stage string) {
		t.Helper()
		got := map[string]uint64{}
		for bucket := range db.Buckets() {
			for entry := range db.Ascend(bucket) {
				buf, err := io.ReadAll(entry.val.NewReader())
				if err != nil {
					t.Fatal(err)
				}

				if !strings.HasPrefix(entry.key, bucket+"/") {
					t.Fatalf("%s: key %s in bucket %s", stage, entry.key, bucket)
				}

				if err := checkTestValue(entry.key, buf); err != nil {
					t.Fatalf("%s: %v", stage, err)
				}

				got[entry.key] = binary.LittleEndian.Uint64(buf)
			}
		}

		if len(got) != len(want) {
			t.Fatalf("%s: expected %d entries, got %d", stage, len(want), len(got))
		}

		for k, v := range want {
			if got[k] != v {
				t.Fatalf("%s: entry %s has version %d, want %d", stage, k, got[k], v)
			}
		}
	}

	check("after the run")

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	check("after re-open")

	if err := db.compact(true); err != nil {
		t.Fatal(err)
	}

	check("after a compaction of the log")

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	check("after another compaction")
}

// TestDB_ReaderSurvivesCompaction checks that a reader, which has been opened before compactions replaced the files
// and the value, still reads the former value.
func TestDB_ReaderSurvivesCompaction(t *testing.T) {
	for _, logOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("log only %v", logOnly), func(t *testing.T) {
			testReaderSurvivesCompaction(t, logOnly)
		})
	}
}

func testReaderSurvivesCompaction(t *testing.T, logOnly bool) {
	db, err := OpenWithOptions(t.TempDir(), Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	before := bytes.Repeat([]byte("before"), 1000)
	if err := db.Set("b", "k", before); err != nil {
		t.Fatal(err)
	}

	reader := db.Get("b", "k").Unwrap()
	head := make([]byte, 10)
	if _, err := io.ReadFull(reader, head); err != nil {
		t.Fatal(err)
	}

	// the first compaction moves the value from the log into the compacted file, the second one replaces that file,
	// while a compaction of the log alone replaces the log twice
	for i := range 2 {
		if err := db.Set("b", "k", []byte(fmt.Sprintf("after %d", i))); err != nil {
			t.Fatal(err)
		}

		if err := db.compact(logOnly); err != nil {
			t.Fatal(err)
		}
	}

	rest, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(append(head, rest...), before) {
		t.Fatal("the reader got another value")
	}

	buf, err := io.ReadAll(db.Get("b", "k").Unwrap())
	if err != nil || string(buf) != "after 1" {
		t.Fatalf("got %q %v", buf, err)
	}
}

// TestDB_IteratorSurvivesCompaction checks that an iterator keeps its snapshot, while a compaction replaces the files.
func TestDB_IteratorSurvivesCompaction(t *testing.T) {
	for _, logOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("log only %v", logOnly), func(t *testing.T) {
			testIteratorSurvivesCompaction(t, logOnly)
		})
	}
}

func testIteratorSurvivesCompaction(t *testing.T, logOnly bool) {
	db, err := OpenWithOptions(t.TempDir(), Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	for i := range 100 {
		key := fmt.Sprintf("k%03d", i)
		if err := db.Set("b", key, []byte("old "+key)); err != nil {
			t.Fatal(err)
		}
	}

	n := 0
	for entry := range db.Ascend("b") {
		if n == 10 {
			for i := range 100 {
				key := fmt.Sprintf("k%03d", i)
				if err := db.Set("b", key, []byte("new "+key)); err != nil {
					t.Fatal(err)
				}
			}

			if err := db.Delete("b", "k050"); err != nil {
				t.Fatal(err)
			}

			if err := db.compact(logOnly); err != nil {
				t.Fatal(err)
			}
		}

		buf, err := io.ReadAll(entry.val.NewReader())
		if err != nil {
			t.Fatal(err)
		}

		if string(buf) != "old "+entry.key {
			t.Fatalf("snapshot entry %s has value %q", entry.key, buf)
		}

		n++
	}

	if n != 100 {
		t.Fatalf("the snapshot has %d entries", n)
	}

	if db.Len("b") != 99 {
		t.Fatalf("expected 99 entries, got %d", db.Len("b"))
	}
}

// TestDB_AutomaticCompaction checks that overwrites start compactions, which keep the files below about twice the
// live data.
func TestDB_AutomaticCompaction(t *testing.T) {
	dir := t.TempDir()
	const minGarbage = 256 * 1024
	db, err := OpenWithOptions(dir, Options{MinGarbage: minGarbage})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	const keys = 200
	const writes = 20_000
	value := make([]byte, 1000)
	var maxSize int64
	for i := range writes {
		binary.LittleEndian.PutUint64(value, uint64(i))
		if err := db.Set("b", fmt.Sprintf("k%03d", i%keys), value); err != nil {
			t.Fatal(err)
		}

		maxSize = max(maxSize, db.Stats().FileSize)

		// about 20 MB/s at most, a writer without a pause writes far more than a service and outruns any fsync
		if i%20 == 0 {
			time.Sleep(time.Millisecond)
		}
	}

	waitForCompaction(t, db)
	stats := db.Stats()
	live := stats.FileSize - stats.Garbage
	// the changes are spread over all entries, thus the log would hold most of them after a compaction of the log
	if stats.Compactions < 10 || stats.LogCompactions != 0 {
		t.Fatalf("expected many compactions of both files, got %d and %d of the log", stats.Compactions, stats.LogCompactions)
	}

	if live < keys*1000 || live > keys*1100 {
		t.Fatalf("estimated live data of %d is off", live)
	}

	// at rest, the garbage stays below the limit, because a compaction checks the limit again when it is done
	if stats.Garbage >= max(minGarbage, live) {
		t.Fatalf("garbage of %d is above the limit with live data of %d", stats.Garbage, live)
	}

	// The writes go on while a compaction runs, thus the files may grow beyond twice the live data for a moment, the
	// more the longer a sync takes. Without compactions, they would grow to all writes.
	if limit := int64(writes * 1000 / 2); maxSize > limit {
		t.Fatalf("the files grew to %d, more than %d with live data of %d", maxSize, limit, live)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if db.Len("b") != keys {
		t.Fatalf("expected %d entries, got %d", keys, db.Len("b"))
	}
}

func waitForCompaction(t *testing.T, db *DB) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for db.Stats().Compacting {
		if time.Now().After(deadline) {
			t.Fatal("compaction does not finish")
		}

		time.Sleep(time.Millisecond)
	}
}

// TestDB_GarbageAccounting checks the estimated garbage of overwrites and removals and that a compaction removes it.
func TestDB_GarbageAccounting(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenWithOptions(dir, Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	for i := range 100 {
		if err := db.Set("b", "k", []byte(fmt.Sprintf("value %03d", i))); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Set("b", "gone", []byte("short lived")); err != nil {
		t.Fatal(err)
	}

	if err := db.Delete("b", "gone"); err != nil {
		t.Fatal(err)
	}

	stats := db.Stats()
	live := recordSize(101, 1, 1, 9)
	if got := stats.FileSize - stats.Garbage; got != live {
		t.Fatalf("expected live data of %d, got %d (file size %d, garbage %d)", live, got, stats.FileSize, stats.Garbage)
	}

	// the garbage survives a re-open
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = OpenWithOptions(dir, Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	if reopened := db.Stats(); reopened.Garbage != stats.Garbage || reopened.FileSize != stats.FileSize {
		t.Fatalf("expected %+v after re-open, got %+v", stats, reopened)
	}

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	stats = db.Stats()
	if stats.Garbage != 0 || stats.FileSize != recordSize(102, 1, 1, 9) {
		t.Fatalf("expected only the live record after compaction, got %+v", stats)
	}
}

// TestDB_RecoverInterruptedCompaction checks the states, which a crash may leave while a compaction replaces the files.
func TestDB_RecoverInterruptedCompaction(t *testing.T) {
	// prepare writes entries, compacts and writes more, so that both files hold entries
	prepare := func(t *testing.T) (string, map[string]string) {
		dir := t.TempDir()
		db, err := OpenWithOptions(dir, Options{GCPercent: -1})
		if err != nil {
			t.Fatal(err)
		}

		want := map[string]string{}
		set := func(k, v string) {
			if err := db.Set("b", k, []byte(v)); err != nil {
				t.Fatal(err)
			}
			want[k] = v
		}

		for i := range 50 {
			set(fmt.Sprintf("k%02d", i), "first")
		}

		if err := db.Compact(); err != nil {
			t.Fatal(err)
		}

		for i := range 25 {
			set(fmt.Sprintf("k%02d", i*2), "second")
		}

		if err := db.Close(); err != nil {
			t.Fatal(err)
		}

		return dir, want
	}

	check := func(t *testing.T, dir string, want map[string]string) {
		t.Helper()
		db, err := Open(dir)
		if err != nil {
			t.Fatal(err)
		}

		defer closeDB(&db)

		if db.Len("b") != len(want) {
			t.Fatalf("expected %d entries, got %d", len(want), db.Len("b"))
		}

		for k, v := range want {
			buf, err := io.ReadAll(db.Get("b", k).Unwrap())
			if err != nil || string(buf) != v {
				t.Fatalf("entry %s: %q %v, want %q", k, buf, err, v)
			}
		}

		for _, suffix := range []string{nextSuffix, oldSuffix} {
			if left, _ := filepath.Glob(filepath.Join(dir, "*"+suffix)); len(left) > 0 {
				t.Fatalf("files left behind: %v", left)
			}
		}
	}

	path := func(dir, name string) string {
		return filepath.Join(dir, name)
	}

	rename := func(t *testing.T, from, to string) {
		t.Helper()
		if err := os.Rename(from, to); err != nil {
			t.Fatal(err)
		}
	}

	write := func(t *testing.T, name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("unfinished new files", func(t *testing.T) {
		dir, want := prepare(t)
		write(t, path(dir, snapName+nextSuffix), []byte("incomplete"))
		write(t, path(dir, logName+nextSuffix), []byte("incomplete"))
		check(t, dir, want)
	})

	t.Run("compacted file moved away", func(t *testing.T) {
		dir, want := prepare(t)
		rename(t, path(dir, snapName), path(dir, snapName+nextSuffix))
		check(t, dir, want)
	})

	t.Run("compacted file moved away and kept as old", func(t *testing.T) {
		dir, want := prepare(t)
		rename(t, path(dir, snapName), path(dir, snapName+oldSuffix))
		check(t, dir, want)
	})

	t.Run("log moved away", func(t *testing.T) {
		dir, want := prepare(t)
		rename(t, path(dir, logName), path(dir, logName+nextSuffix))
		write(t, path(dir, logName+oldSuffix), []byte("former log"))
		check(t, dir, want)
	})

	t.Run("former files left", func(t *testing.T) {
		dir, want := prepare(t)
		write(t, path(dir, snapName+oldSuffix), []byte("former compacted file"))
		write(t, path(dir, logName+oldSuffix), []byte("former log"))
		write(t, path(dir, "0123456789abcdef.compact.tmp"), []byte("temp file of the former compaction"))
		check(t, dir, want)
		if left, _ := filepath.Glob(filepath.Join(dir, "*.compact.tmp")); len(left) > 0 {
			t.Fatalf("files left behind: %v", left)
		}
	})

	// a crash after the compacted file has been replaced, but before the log: the former log holds the records of the
	// new compacted file as well, which a load must skip
	t.Run("new compacted file with the former log", func(t *testing.T) {
		dir := t.TempDir()
		db, err := OpenWithOptions(dir, Options{GCPercent: -1})
		if err != nil {
			t.Fatal(err)
		}

		want := map[string]string{}
		set := func(k, v string) {
			if err := db.Set("b", k, []byte(v)); err != nil {
				t.Fatal(err)
			}
			want[k] = v
		}

		for i := range 50 {
			set(fmt.Sprintf("k%02d", i), "first")
		}

		for i := range 10 {
			if err := db.Delete("b", fmt.Sprintf("k%02d", i)); err != nil {
				t.Fatal(err)
			}
			delete(want, fmt.Sprintf("k%02d", i))
		}

		formerLog, err := os.ReadFile(path(dir, logName))
		if err != nil {
			t.Fatal(err)
		}

		if err := db.Compact(); err != nil {
			t.Fatal(err)
		}

		for i := range 25 {
			set(fmt.Sprintf("k%02d", i*2), "second")
		}

		if err := db.Close(); err != nil {
			t.Fatal(err)
		}

		newLog, err := os.ReadFile(path(dir, logName))
		if err != nil {
			t.Fatal(err)
		}

		write(t, path(dir, logName), append(formerLog, newLog...))
		check(t, dir, want)
	})
}

// TestBuildTree checks that the two passes keep every item and that later items replace earlier ones.
func TestBuildTree(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for _, n := range []int{0, 1, 2, 3, 63, 64, 65, 1000, 4097} {
		var items []item
		for i := range n {
			items = append(items, item{key: fmt.Sprintf("%08d", i*2), off: uint64(i)})
		}

		check := func(items []item, want map[string]uint64) {
			t.Helper()
			tree := buildTree(items)
			if tree.Len() != len(want) {
				t.Fatalf("expected %d items, got %d", len(want), tree.Len())
			}

			var keys []string
			tree.Ascend(func(it item) bool {
				keys = append(keys, it.key)
				if want[it.key] != it.off {
					t.Fatalf("item %s has %d, want %d", it.key, it.off, want[it.key])
				}
				return true
			})

			if !slices.IsSorted(keys) {
				t.Fatal("not sorted")
			}
		}

		want := map[string]uint64{}
		for _, it := range items {
			want[it.key] = it.off
		}

		check(items, want)

		// unsorted items with duplicates, the later one wins
		if n > 1 {
			shuffled := slices.Clone(items)
			r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			dup := item{key: shuffled[0].key, off: 999999}
			shuffled = append(shuffled, dup)
			want[dup.key] = dup.off
			check(shuffled, want)
		}
	}
}

// TestKeyArena checks that the keys of an arena stay intact, while it fills more blocks.
func TestKeyArena(t *testing.T) {
	var a keyArena
	var keys, want []string
	buf := []byte{}
	for i := range 20000 {
		buf = fmt.Appendf(buf[:0], "key-%d-%s", i, strings.Repeat("x", i%300))
		keys = append(keys, a.add(unsafeString(buf)))
		want = append(want, string(buf))
	}

	big := strings.Repeat("y", arenaChunk)
	keys = append(keys, a.add(big))
	want = append(want, big)
	if !slices.Equal(keys, want) {
		t.Fatal("the arena has changed a key")
	}

	if a.add("") != "" {
		t.Fatal("empty key")
	}
}

// logRecords lists the records of the log as "set bucket/key" and "del bucket/key" in ascending order.
func logRecords(t *testing.T, db *DB) []string {
	t.Helper()
	db.mu.RLock()
	log := db.st.log
	db.mu.RUnlock()

	var res []string
	if _, err := scanRecords(log.f, 0, log.size.Load(), func(rec record, off int64, raw []byte) error {
		kind := "set"
		if rec.kind == removeKeyValue {
			kind = "del"
		}

		res = append(res, fmt.Sprintf("%s %s/%s", kind, rec.bucket, rec.key))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	slices.Sort(res)
	return res
}

// checkModel compares the entries of the DB with the model of bucket/key to value.
func checkModel(t *testing.T, db *DB, stage string, want map[string]string) {
	t.Helper()
	got := map[string]string{}
	for bucket := range db.Buckets() {
		for entry := range db.Ascend(bucket) {
			buf, err := io.ReadAll(entry.val.NewReader())
			if err != nil {
				t.Fatal(err)
			}

			got[bucket+"/"+entry.key] = string(buf)
		}
	}

	if len(got) != len(want) {
		t.Fatalf("%s: expected %d entries, got %d", stage, len(want), len(got))
	}

	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s: %s = %q, want %q", stage, k, got[k], v)
		}
	}
}

// TestDB_LogCompaction checks that a compaction of the log keeps the compacted file, the live records and the removals,
// which the entries of the compacted file need.
func TestDB_LogCompaction(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenWithOptions(dir, Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	model := map[string]string{}
	set := func(bucket, key, value string) {
		t.Helper()
		if err := db.Set(bucket, key, []byte(value)); err != nil {
			t.Fatal(err)
		}

		model[bucket+"/"+key] = value
	}

	del := func(bucket, key string) {
		t.Helper()
		if err := db.Delete(bucket, key); err != nil {
			t.Fatal(err)
		}

		delete(model, bucket+"/"+key)
	}

	for i := range 1000 {
		set("cold", fmt.Sprintf("c%04d", i), fmt.Sprintf("%s %d", strings.Repeat("c", 500), i))
	}

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	snapBefore, err := os.Stat(filepath.Join(dir, snapName))
	if err != nil {
		t.Fatal(err)
	}

	for round := range 50 {
		for i := range 20 {
			set("hot", fmt.Sprintf("h%02d", i), fmt.Sprintf("%d-%d", round, i))
		}
	}

	del("hot", "h00")
	del("cold", "c0001") // the compacted file holds the entry, thus the removal must stay
	del("cold", "c0002") // the new entry replaces the removal
	set("cold", "c0002", "again")
	set("cold", "c0003", "changed")
	set("temp", "t", "x") // a removal stays as long as the entry is removed, even if the compacted file lacks it
	del("temp", "t")

	db.mu.RLock()
	snapGarbage := db.st.snapGarbage
	db.mu.RUnlock()

	if err := db.compact(true); err != nil {
		t.Fatal(err)
	}

	stats := db.Stats()
	if stats.Compactions != 1 || stats.LogCompactions != 1 {
		t.Fatalf("expected one compaction of each kind, got %d and %d of the log", stats.Compactions, stats.LogCompactions)
	}

	snapAfter, err := os.Stat(filepath.Join(dir, snapName))
	if err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(snapBefore, snapAfter) || snapBefore.Size() != snapAfter.Size() || !snapBefore.ModTime().Equal(snapAfter.ModTime()) {
		t.Fatal("the compaction of the log changed the compacted file")
	}

	want := []string{"del cold/c0001", "del hot/h00", "del temp/t", "set cold/c0002", "set cold/c0003"}
	for i := 1; i < 20; i++ {
		want = append(want, fmt.Sprintf("set hot/h%02d", i))
	}

	slices.Sort(want)
	if got := logRecords(t, db); !slices.Equal(got, want) {
		t.Fatalf("log after the compaction:\n%v\nwant\n%v", got, want)
	}

	// the garbage of the log are its removals, the one of the compacted file is unchanged
	db.mu.RLock()
	st := db.st
	removals := st.logRemovals
	if st.snapGarbage != snapGarbage || st.logGarbage != removals || removals == 0 {
		t.Fatalf("unexpected garbage: compacted file %d (before %d), log %d, removals %d", st.snapGarbage, snapGarbage, st.logGarbage, removals)
	}
	db.mu.RUnlock()

	checkModel(t, db, "after the compaction of the log", model)

	// writes go to the new log
	set("hot", "h01", "after")
	del("cold", "c0004")
	checkModel(t, db, "after writes", model)

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = OpenWithOptions(dir, Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	checkModel(t, db, "after re-open", model)

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	checkModel(t, db, "after a compaction of both files", model)
	if got := logRecords(t, db); len(got) != 0 {
		t.Fatalf("expected an empty log, got %v", got)
	}

	if stats := db.Stats(); stats.Garbage != 0 {
		t.Fatalf("expected no garbage, got %d", stats.Garbage)
	}
}

// TestDB_PlanCompaction checks the choice between a compaction of the log alone and one of both files.
func TestDB_PlanCompaction(t *testing.T) {
	opts := Options{GCPercent: 100, MinGarbage: 1}
	setup := func(t *testing.T) *DB {
		db, err := OpenWithOptions(t.TempDir(), Options{GCPercent: -1})
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = db.Close() })
		value := strings.Repeat("v", 1000)
		for i := range 1000 {
			if err := db.Set("b", fmt.Sprintf("k%04d", i), []byte(value)); err != nil {
				t.Fatal(err)
			}
		}

		if err := db.Compact(); err != nil {
			t.Fatal(err)
		}

		return db
	}

	// overwrite changes count entries round by round and stops, when a compaction is due
	overwrite := func(t *testing.T, db *DB, count int, opts Options) (due, logOnly bool) {
		for round := 0; ; round++ {
			for i := range count {
				if err := db.Set("b", fmt.Sprintf("k%04d", i), []byte(fmt.Sprintf("%d %s", round, strings.Repeat("v", 1000)))); err != nil {
					t.Fatal(err)
				}

				db.mu.RLock()
				due, logOnly = db.st.plan(opts)
				db.mu.RUnlock()
				if due {
					return due, logOnly
				}
			}
		}
	}

	t.Run("few entries change often", func(t *testing.T) {
		if _, logOnly := overwrite(t, setup(t), 50, opts); !logOnly {
			t.Fatal("expected a compaction of the log")
		}
	})

	t.Run("changes spread over all entries", func(t *testing.T) {
		if _, logOnly := overwrite(t, setup(t), 1000, opts); logOnly {
			t.Fatal("expected a compaction of both files")
		}
	})

	t.Run("most entries in the log", func(t *testing.T) {
		// the compaction of the log would free more per written byte, but it would keep 60% of the entries in the log
		db := setup(t)
		if _, logOnly := overwrite(t, db, 600, Options{GCPercent: 200, MinGarbage: 1}); logOnly {
			t.Fatal("expected a compaction of both files")
		}

		db.mu.RLock()
		st := db.st
		garbage := st.garbage()
		live := st.snap.size.Load() + st.log.size.Load() - garbage
		logFree := st.logGarbage - st.logRemovals
		logWrite := st.log.size.Load() - logFree
		db.mu.RUnlock()
		if logFree*live < garbage*logWrite || 2*logWrite <= live {
			t.Fatalf("the test does not reach the case: log frees %d and writes %d, garbage %d, live %d", logFree, logWrite, garbage, live)
		}
	})

	t.Run("only removals", func(t *testing.T) {
		db := setup(t)
		for i := range 1000 {
			if err := db.Delete("b", fmt.Sprintf("k%04d", i)); err != nil {
				t.Fatal(err)
			}
		}

		db.mu.RLock()
		due, logOnly := db.st.plan(opts)
		db.mu.RUnlock()
		if !due || logOnly {
			t.Fatalf("expected a compaction of both files, got due %v, log only %v", due, logOnly)
		}
	})
}

// TestDB_AutomaticLogCompaction checks that a few entries, which change often, start compactions of the log alone,
// which leave the compacted file and the many entries in it as they are.
func TestDB_AutomaticLogCompaction(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenWithOptions(dir, Options{MinGarbage: 64 * 1024})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	cold := strings.Repeat("c", 1000)
	for i := range 2000 {
		if err := db.Set("cold", fmt.Sprintf("c%04d", i), []byte(cold)); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	snapBefore, err := os.Stat(filepath.Join(dir, snapName))
	if err != nil {
		t.Fatal(err)
	}

	const writes = 50_000
	value := make([]byte, 200)
	for i := range writes {
		binary.LittleEndian.PutUint64(value, uint64(i))
		if err := db.Set("hot", fmt.Sprintf("h%02d", i%50), value); err != nil {
			t.Fatal(err)
		}

		if i%50 == 0 {
			time.Sleep(time.Millisecond)
		}
	}

	waitForCompaction(t, db)
	stats := db.Stats()
	if stats.Compactions != 1 || stats.LogCompactions < 3 {
		t.Fatalf("expected only compactions of the log, got %d and %d of the log", stats.Compactions, stats.LogCompactions)
	}

	snapAfter, err := os.Stat(filepath.Join(dir, snapName))
	if err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(snapBefore, snapAfter) {
		t.Fatal("the compacted file has been replaced")
	}

	// at rest, the garbage stays below the live data
	if live := stats.FileSize - stats.Garbage; stats.Garbage >= live {
		t.Fatalf("garbage of %d is above the live data of %d", stats.Garbage, live)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	for i := range 50 {
		buf, err := io.ReadAll(db.Get("hot", fmt.Sprintf("h%02d", i)).Unwrap())
		if err != nil {
			t.Fatal(err)
		}

		if got, want := binary.LittleEndian.Uint64(buf), uint64(writes-50+i); got != want {
			t.Fatalf("h%02d has version %d, want %d", i, got, want)
		}
	}

	if db.Len("cold") != 2000 {
		t.Fatalf("expected 2000 cold entries, got %d", db.Len("cold"))
	}
}

// TestDB_ReaderSurvivesLogCompaction checks that a reader of a value in the log keeps reading it, while compactions
// of the log drop the value and replace the log.
func TestDB_ReaderSurvivesLogCompaction(t *testing.T) {
	db, err := OpenWithOptions(t.TempDir(), Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	before := bytes.Repeat([]byte("before"), 1000)
	if err := db.Set("b", "k", before); err != nil {
		t.Fatal(err)
	}

	reader := db.Get("b", "k").Unwrap()
	for i := range 3 {
		if err := db.Set("b", "k", []byte(fmt.Sprintf("after %d", i))); err != nil {
			t.Fatal(err)
		}

		if err := db.compact(true); err != nil {
			t.Fatal(err)
		}
	}

	buf, err := io.ReadAll(reader)
	if err != nil || !bytes.Equal(buf, before) {
		t.Fatalf("the reader got another value: %v", err)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"go.wdy.de/nago/pkg/xbytes"
)

// TestAppendRecordMatchesNode checks that appendRecord writes the format of Node byte by byte, which the former
// implementation has written, and that decodePayload reads it back.
func TestAppendRecordMatchesNode(t *testing.T) {
	r := rand.New(rand.NewSource(1234))
	for i := range 2000 {
		n := Node{
			kind:   nodeKind(r.Intn(2) + 1),
			tx:     r.Uint64() >> r.Intn(64),
			bucket: randBytes(r, 64),
			key:    randBytes(r, 300),
		}

		if n.kind == setKeyValue {
			n.val = randBytes(r, 20000)
		}

		var want xbytes.Buffer
		n.write(&want)

		got, valOff := appendRecord(nil, n.kind, n.tx, string(n.bucket), string(n.key), n.val)
		if !bytes.Equal(got, want.Buf[:want.Pos]) {
			t.Fatalf("record %d differs from the Node encoding", i)
		}

		if n.kind == setKeyValue && (valOff != int(n.valOffset) || !bytes.Equal(got[valOff:valOff+len(n.val)], n.val)) {
			t.Fatalf("record %d: wrong value offset %d, want %d", i, valOff, n.valOffset)
		}

		rec, err := decodePayload(n.kind, got[headerSize:])
		if err != nil {
			t.Fatal(err)
		}

		if rec.tx != n.tx || !bytes.Equal(rec.bucket, n.bucket) || !bytes.Equal(rec.key, n.key) || int(rec.valLen) != len(n.val) {
			t.Fatalf("record %d decoded wrongly", i)
		}
	}
}

func randBytes(r *rand.Rand, maxLen int) []byte {
	b := make([]byte, r.Intn(maxLen))
	r.Read(b)
	return b
}

// TestScanRecords writes records into a file and reads them back, also with a torn tail.
func TestScanRecords(t *testing.T) {
	type entry struct {
		kind      nodeKind
		b, k, v   []byte
		recordOff int64
	}

	r := rand.New(rand.NewSource(1234))
	var buf []byte
	var entries []entry
	for i := range 5000 {
		e := entry{kind: nodeKind(r.Intn(2) + 1), b: randBytes(r, 64), k: randBytes(r, 128), recordOff: int64(len(buf))}
		if e.kind == setKeyValue {
			e.v = randBytes(r, 16*1024)
		}

		entries = append(entries, e)
		buf, _ = appendRecord(buf, e.kind, uint64(i), string(e.b), string(e.k), e.v)
	}

	path := filepath.Join(t.TempDir(), "records")
	if err := os.WriteFile(path, buf, 0600); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	scan := func(to int64) (int, int64, error) {
		idx := 0
		end, err := scanRecords(f, 0, to, func(rec record, off int64, raw []byte) error {
			e := entries[idx]
			if off != e.recordOff || rec.kind != e.kind || rec.tx != uint64(idx) || !bytes.Equal(rec.bucket, e.b) || !bytes.Equal(rec.key, e.k) {
				t.Fatalf("record %d differs", idx)
			}

			if rec.kind == setKeyValue {
				val := make([]byte, rec.valLen)
				if _, err := f.ReadAt(val, rec.valOff); err != nil && !(errors.Is(err, io.EOF) && len(val) == 0) {
					t.Fatal(err)
				}

				if !bytes.Equal(val, e.v) {
					t.Fatalf("value of record %d differs", idx)
				}
			}

			idx++
			return nil
		})

		return idx, end, err
	}

	n, end, err := scan(int64(len(buf)))
	if err != nil || n != len(entries) || end != int64(len(buf)) {
		t.Fatalf("expected %d records up to %d, got %d up to %d: %v", len(entries), len(buf), n, end, err)
	}

	// a section, which ends within the last record, reads the others and reports the torn record
	last := entries[len(entries)-1].recordOff
	for _, cut := range []int64{last + 1, last + headerSize, int64(len(buf)) - 1} {
		n, end, err := scan(cut)
		if !errors.Is(err, errTornRecord) || n != len(entries)-1 || end != last {
			t.Fatalf("cut at %d: expected a torn record at %d after %d records, got %d at %d: %v", cut, last, len(entries)-1, n, end, err)
		}
	}
}

// TestDB_TornLogTail checks that a record, which a crash has left incomplete at the end of the log, is removed, so
// that later records do not follow it.
func TestDB_TornLogTail(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	for _, k := range []string{"a", "b", "c"} {
		if err := db.Set("b", k, []byte("value of "+k)); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	logFile := filepath.Join(dir, logName)
	torn, _ := appendRecord(nil, setKeyValue, 99, "b", "d", []byte("value of d"))
	for _, tail := range [][]byte{torn[:len(torn)-3], torn[:4], make([]byte, 100)} {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Write(tail); err != nil {
			t.Fatal(err)
		}

		_ = f.Close()

		db, err = Open(dir)
		if err != nil {
			t.Fatal(err)
		}

		if db.Exists("b", "d") {
			t.Fatal("the torn record must not be applied")
		}

		if err := db.Set("b", "e", []byte("value of e")); err != nil {
			t.Fatal(err)
		}

		if err := db.Close(); err != nil {
			t.Fatal(err)
		}

		db, err = Open(dir)
		if err != nil {
			t.Fatalf("a record after the torn one must be readable: %v", err)
		}

		for _, k := range []string{"a", "b", "c", "e"} {
			buf, err := io.ReadAll(db.Get("b", k).Unwrap())
			if err != nil || string(buf) != "value of "+k {
				t.Fatalf("entry %s: %q %v", k, buf, err)
			}
		}

		if err := db.Delete("b", "e"); err != nil {
			t.Fatal(err)
		}

		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}

	db = nil
}

func BenchmarkAppendRecord(b *testing.B) {
	val := make([]byte, 512)
	var buf []byte
	for b.Loop() {
		buf, _ = appendRecord(buf[:0], setKeyValue, 123456, "bucket-0123456789", "key-0123456789012345678901234567890", val)
	}
}

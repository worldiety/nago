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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// logBuilder writes a log, as the former versions have left it, and the entries, which it holds.
type logBuilder struct {
	buf   []byte
	tx    uint64
	model map[string]string
	offs  map[string]int // the offset of the last record of a key
}

func (l *logBuilder) set(key, value string) {
	l.tx++
	l.offs[key] = len(l.buf)
	l.buf, _ = appendRecord(l.buf, setKeyValue, l.tx, "b", key, []byte(value))
	l.model["b/"+key] = value
}

func (l *logBuilder) del(key string) {
	l.tx++
	l.offs[key] = len(l.buf)
	l.buf, _ = appendRecord(l.buf, removeKeyValue, l.tx, "b", key, nil)
	delete(l.model, "b/"+key)
}

func (l *logBuilder) setSome() {
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		l.set(k, "value of "+k)
	}
}

// torn appends the first n bytes of a record, as a crash leaves it. A former version has replayed the log up to it
// after a restart, thus its next record reuses the transaction.
func (l *logBuilder) torn(key string, valueLen, n int) {
	rec, _ := appendRecord(nil, setKeyValue, l.tx+1, "b", key, bytes.Repeat([]byte{'v'}, valueLen))
	if n < 0 {
		n += len(rec)
	}

	l.buf = append(l.buf, rec[:n]...)
}

// TestDB_BadLogEnd checks the logs, which a crash may leave behind, together with the former versions, which have
// appended behind an incomplete record after a restart. Open must never remove valid records, and fail on corruption.
func TestDB_BadLogEnd(t *testing.T) {
	tests := []struct {
		name    string
		build   func(l *logBuilder)
		wantErr error // Open fails and leaves the log as it is
		skipped bool  // Open rewrites the log
	}{
		{
			name: "incomplete record",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, -5)
			},
		},
		{
			name: "incomplete record, behind which a former version appended into its size",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, 50)
				l.set("g", "value of g")
			},
			skipped: true,
		},
		{
			name: "incomplete record, behind which a former version appended beyond its size",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, 150)
				for _, k := range []string{"g", "h", "i", "j"} {
					l.set(k, "value of "+k)
				}

				l.del("a")
			},
			skipped: true,
		},
		{
			name: "two incomplete records, behind which a former version appended",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, 150)
				l.set("g", "value of g")
				l.torn("h", 300, 20)
				l.set("i", "value of i")
				l.set("j", "value of j")
				l.torn("k", 10, 12)
			},
			skipped: true,
		},
		{
			name: "incomplete record filled with zeros",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, -40)
				l.buf = append(l.buf, make([]byte, 5000)...)
			},
		},
		{
			name: "incomplete record filled with zeros, behind which a former version appended",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, -40)
				l.buf = append(l.buf, make([]byte, 300)...)
				l.set("g", "value of g")
				l.set("h", "value of h")
			},
			skipped: true,
		},
		{
			name: "complete record filled with zeros",
			build: func(l *logBuilder) {
				l.setSome()
				l.torn("f", 200, -1)
				l.buf = append(l.buf, 'v')
				clear(l.buf[len(l.buf)-20:])
			},
		},
		{
			name: "incomplete header of garbage",
			build: func(l *logBuilder) {
				l.setSome()
				l.buf = append(l.buf, 0xff, 0x01, 0x02)
			},
		},
		{
			name: "header of garbage beyond the end",
			build: func(l *logBuilder) {
				l.setSome()
				l.buf = append(l.buf, 0x07, 0xe8, 0x03, 0, 0, 1, 2, 3, 4)
				l.buf = append(l.buf, bytes.Repeat([]byte{0x55}, 10)...)
			},
		},
		{
			name: "size beyond the end within the log",
			build: func(l *logBuilder) {
				l.setSome()
				l.set("f", "value of f")
				l.set("g", "value of g")
				l.set("h", "value of h")
				binary.LittleEndian.PutUint32(l.buf[l.offs["f"]+1:], 0x00ffffff)
				delete(l.model, "b/f")
			},
			skipped: true,
		},
		{
			name: "corrupt record within the log",
			build: func(l *logBuilder) {
				l.setSome()
				l.buf[l.offs["d"]-1] ^= 0xff
			},
			wantErr: InvalidNodeCRC,
		},
		{
			name: "corrupt last record",
			build: func(l *logBuilder) {
				l.setSome()
				l.buf[len(l.buf)-1] ^= 0xff
			},
			wantErr: InvalidNodeCRC,
		},
		{
			name: "invalid kind within the log",
			build: func(l *logBuilder) {
				l.setSome()
				l.buf[l.offs["c"]] = 0x07
			},
			wantErr: InvalidNodeType,
		},
		{
			name: "set turned into a removal",
			build: func(l *logBuilder) {
				l.setSome()
				l.set("f", "value of f")
				l.buf[l.offs["c"]] = byte(removeKeyValue)
			},
			wantErr: errCorruptPayload,
		},
		{
			name: "removal turned into a set",
			build: func(l *logBuilder) {
				l.setSome()
				l.del("b")
				l.set("f", "value of f")
				l.buf[l.offs["b"]] = byte(setKeyValue)
			},
			wantErr: errCorruptPayload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &logBuilder{model: map[string]string{}, offs: map[string]int{}}
			tt.build(l)
			dir := t.TempDir()
			logFile := filepath.Join(dir, logName)
			if err := os.WriteFile(logFile, l.buf, 0600); err != nil {
				t.Fatal(err)
			}

			db, err := Open(dir)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}

				if buf, err := os.ReadFile(logFile); err != nil || !bytes.Equal(buf, l.buf) {
					t.Fatalf("a failed Open must leave the log as it is: %v", err)
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			defer closeDB(&db)

			checkModel(t, db, "after Open", l.model)
			if compactions := db.Stats().Compactions; compactions != map[bool]uint64{false: 0, true: 1}[tt.skipped] {
				t.Fatalf("expected a rewrite of the log %v, got %d compactions", tt.skipped, compactions)
			}

			// the log holds no bad record anymore, which a compaction of the log would read
			log := db.st.log
			if end, err := scanRecordsNop(log.f, 0, log.size.Load()); err != nil || end != log.size.Load() {
				t.Fatalf("the log ends at %d of %d: %v", end, log.size.Load(), err)
			}

			if err := db.Set("b", "z", []byte("value of z")); err != nil {
				t.Fatal(err)
			}

			l.model["b/z"] = "value of z"
			if err := db.compact(true); err != nil {
				t.Fatal(err)
			}

			if err := db.Close(); err != nil {
				t.Fatal(err)
			}

			if db, err = Open(dir); err != nil {
				t.Fatal(err)
			}

			checkModel(t, db, "after reopen", l.model)
		})
	}
}

// scanRecordsNop scans the records for checks, which only need the end.
func scanRecordsNop(f *os.File, from, to int64) (int64, error) {
	return scanRecords(f, from, to, func(record, int64, []byte) error { return nil })
}

// TestDB_OversizedRecordOfFormerVersion checks that a record larger than the limit of Set, which former versions have
// written, can be read.
func TestDB_OversizedRecordOfFormerVersion(t *testing.T) {
	big := bytes.Repeat([]byte("0123456789abcdef"), maxPayloadSize/16)
	l := &logBuilder{model: map[string]string{}, offs: map[string]int{}}
	l.set("a", "value of a")
	l.tx++
	l.buf, _ = appendRecord(l.buf, setKeyValue, l.tx, "b", "big", big)
	l.set("c", "value of c")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, logName), l.buf, 0600); err != nil {
		t.Fatal(err)
	}

	l.buf = nil
	check := func(stage string) *DB {
		t.Helper()
		db, err := Open(dir)
		if err != nil {
			t.Fatalf("%s: %v", stage, err)
		}

		checkModel(t, db, stage, map[string]string{"b/a": "value of a", "b/big": string(big), "b/c": "value of c"})
		return db
	}

	db := check("after Open")
	defer closeDB(&db)

	if err := db.Set("b", "big", append(big, 'x')); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("Set must keep its limit: %v", err)
	}

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db = check("after a compaction")
}

// TestDB_SyncAcrossCompactions runs Sync while compactions replace the log, which syncs the directory afterwards.
func TestDB_SyncAcrossCompactions(t *testing.T) {
	db, err := OpenWithOptions(t.TempDir(), Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}

			if err := db.Set("b", fmt.Sprintf("k%d", i%100), fmt.Appendf(nil, "value %d", i)); err != nil {
				errs <- err
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}

			if err := db.Sync(); err != nil {
				errs <- err
				return
			}
		}
	}()

	const compactions = 20
	for i := range compactions {
		if err := db.compact(i%2 == 1); err != nil {
			t.Fatal(err)
		}
	}

	close(stop)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	if err := db.Sync(); err != nil {
		t.Fatal(err)
	}

	if renames := db.renames.Load(); renames != compactions || db.syncedDir.Load() != renames {
		t.Fatalf("expected %d synced renames, got %d of %d", compactions, db.syncedDir.Load(), renames)
	}

	if log := db.st.log; log.synced.Load() != log.size.Load() {
		t.Fatalf("Sync has synced %d of %d bytes", log.synced.Load(), log.size.Load())
	}

	if _, err := io.Copy(io.Discard, db.Get("b", "k1").Unwrap()); err != nil {
		t.Fatal(err)
	}
}

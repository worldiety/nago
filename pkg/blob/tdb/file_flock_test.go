// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd

package tdb

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// flockFile opens the file like a former version and locks it. It returns an error, if another one holds the lock.
func flockFile(t *testing.T, path string, wait bool) (*os.File, error) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	how := syscall.LOCK_EX
	if !wait {
		how |= syscall.LOCK_NB
	}

	if err := syscall.Flock(int(f.Fd()), how); err != nil {
		_ = f.Close()
		return nil, err
	}

	return f, nil
}

// TestDB_LocksDataFiles checks that a DB holds the locks of the data files like the former versions, also of the files,
// which a compaction writes.
func TestDB_LocksDataFiles(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenWithOptions(dir, Options{GCPercent: -1})
	if err != nil {
		t.Fatal(err)
	}

	defer closeDB(&db)

	checkLocked := func(stage string) {
		t.Helper()
		for _, name := range []string{snapName, logName} {
			if f, err := flockFile(t, filepath.Join(dir, name), false); !errors.Is(err, syscall.EWOULDBLOCK) {
				if f != nil {
					_ = f.Close()
				}

				t.Fatalf("%s: %s must be locked: %v", stage, name, err)
			}
		}
	}

	checkLocked("after Open")
	for i := range 3 {
		if err := db.Set("b", "k", []byte{byte(i)}); err != nil {
			t.Fatal(err)
		}

		if err := db.compact(i%2 == 1); err != nil {
			t.Fatal(err)
		}

		checkLocked("after a compaction")
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{snapName, logName} {
		f, err := flockFile(t, filepath.Join(dir, name), false)
		if err != nil {
			t.Fatalf("Close must release %s: %v", name, err)
		}

		_ = f.Close()
	}
}

// TestDB_OpenWaitsForFormerVersion checks that Open waits, while a former version holds a data file, also when it
// replaces the file meanwhile, and that it leaves the temporary files of the former version alone.
func TestDB_OpenWaitsForFormerVersion(t *testing.T) {
	for _, name := range []string{snapName, logName} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, name)
			holder, err := flockFile(t, path, true)
			if err != nil {
				t.Fatal(err)
			}

			tmp := filepath.Join(dir, "0123456789abcdef.compact.tmp")
			if err := os.WriteFile(tmp, []byte("compaction of a former version"), 0600); err != nil {
				t.Fatal(err)
			}

			var db *DB
			opened := make(chan error, 1)
			go func() {
				var err error
				db, err = Open(dir)
				opened <- err
			}()

			waiting := func(stage string) {
				t.Helper()
				select {
				case err := <-opened:
					t.Fatalf("%s: Open must wait for the former version: %v", stage, err)
				case <-time.After(200 * time.Millisecond):
				}

				if !exists(tmp) {
					t.Fatalf("%s: Open has removed the file of the running former version", stage)
				}
			}

			waiting("while it holds the file")

			// the former version replaces the file and releases the former one
			replacement, err := flockFile(t, path+".tmp", true)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.Rename(path+".tmp", path); err != nil {
				t.Fatal(err)
			}

			_ = holder.Close()
			waiting("while it holds the replacement")

			_ = replacement.Close()
			select {
			case err := <-opened:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("Open still waits after the former version has released the file")
			}

			defer closeDB(&db)
			if exists(tmp) {
				t.Fatal("the temporary file of the former version must be removed")
			}

			if err := db.Set("b", "k", []byte("v")); err != nil {
				t.Fatal(err)
			}
		})
	}
}

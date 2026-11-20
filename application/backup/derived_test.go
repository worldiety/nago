// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package backup_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
)

// TestBackup_LeavesOutDerivedStores (wokoda R-DEC-ASSISTANT-SESSION-HEADS, „Sie ist als abgeleitet geöffnet; die
// Sicherung lässt sie aus“): a backup of the local stores contains no store opened with blob.OpenStoreOptions.Derived.
// Counter row: the same store opened without Derived, with the same content, is in the backup.
func TestBackup_LeavesOutDerivedStores(t *testing.T) {
	stores, err := application.NewLocalStores(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stores.Close() })

	for _, c := range []struct {
		name    string
		derived bool
	}{{"x.derived", true}, {"x.plain", false}} {
		s, err := stores.Open(c.name, blob.OpenStoreOptions{Type: blob.EntityStore, Derived: c.derived})
		if err != nil {
			t.Fatal(err)
		}
		if err := blob.Put(s, "k", []byte(`{"v":1}`)); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := backup.NewBackup(stores)(context.Background(), user.SU(), &buf); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	f, err := zr.Open("index.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var idx backup.Index
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, s := range idx.Stores {
		names = append(names, s.Name)
	}
	if slices.Contains(names, "x.derived") {
		t.Errorf("the backup contains the derived store: %q", names)
	}
	if !slices.Contains(names, "x.plain") {
		t.Errorf("counter row: the backup lacks the store opened without Derived: %q", names)
	}
}

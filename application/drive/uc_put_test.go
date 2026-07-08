// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"errors"
	"os"
	"testing"

	"go.wdy.de/nago/application/user"
)

// TestPutUsesNameAsFilename verifies that the name argument (not OriginalFilename) is authoritative for the
// stored file name within the parent directory.
func TestPutUsesNameAsFilename(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// name differs from OriginalFilename on purpose
	if err := uc.Put(owner, root.ID, "report.pdf", stringReader("data"), PutOptions{
		Mode:             0600,
		OriginalFilename: "some-uploaded-name.pdf",
	}); err != nil {
		t.Fatalf("put: %v", err)
	}

	entries := entriesOf(t, uc, root.ID)
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 entry, got %d", len(entries))
	}

	file := statFile(t, uc, entries[0])
	if file.Filename != "report.pdf" {
		t.Fatalf("filename should be the name argument, got %q", file.Filename)
	}
	// OriginalFilename must still be preserved as metadata on the version
	versions := file.Versions()
	if len(versions) != 1 || versions[0].FileInfo.OriginalFilename != "some-uploaded-name.pdf" {
		t.Fatalf("OriginalFilename metadata not preserved: %+v", versions)
	}
}

// TestPutSameNameUpdatesInsteadOfDuplicating verifies that putting twice under the same name updates the
// existing file (new version) rather than creating a duplicate entry. This used to break when name differed
// from OriginalFilename, because the stored Filename was set from OriginalFilename and the EntryByName(name)
// lookup on the second put could not find the existing file.
func TestPutSameNameUpdatesInsteadOfDuplicating(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	put := func(content string) {
		if err := uc.Put(owner, root.ID, "doc.txt", stringReader(content), PutOptions{
			Mode:             0600,
			OriginalFilename: "unrelated-original.txt", // differs from name
			KeepVersion:      true,
		}); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	put("v1")
	put("v2")

	entries := entriesOf(t, uc, root.ID)
	if len(entries) != 1 {
		t.Fatalf("expected a single entry (update, not duplicate), got %d", len(entries))
	}

	file := statFile(t, uc, entries[0])
	if file.Filename != "doc.txt" {
		t.Fatalf("filename should be %q, got %q", "doc.txt", file.Filename)
	}
	if got := len(file.Versions()); got != 2 {
		t.Fatalf("expected 2 versions on the same file, got %d", got)
	}
}

// TestPutRejectsInvalidName verifies that an invalid file name is rejected up front.
func TestPutRejectsInvalidName(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// contains a path separator which is not a valid file name
	if err := uc.Put(owner, root.ID, "bad/name.txt", stringReader("data"), PutOptions{Mode: 0600}); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("expected os.ErrInvalid for invalid name, got %v", err)
	}

	// nothing must have been created
	if entries := entriesOf(t, uc, root.ID); len(entries) != 0 {
		t.Fatalf("no entry must be created on invalid name, got %d", len(entries))
	}
}

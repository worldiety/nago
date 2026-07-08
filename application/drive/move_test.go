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

// entriesOf returns the current entry FIDs of the given directory (freshly stat'd).
func entriesOf(t *testing.T, uc UseCases, dir FID) []FID {
	t.Helper()
	d := statFile(t, uc, dir)
	var out []FID
	for fid := range d.Entries.All() {
		out = append(out, fid)
	}
	return out
}

func containsFID(list []FID, fid FID) bool {
	for _, f := range list {
		if f == fid {
			return true
		}
	}
	return false
}

// putFile creates a file in the given parent and returns its FID (the single non-directory entry that
// appeared). Put derives the stored filename from OriginalFilename.
func putFile(t *testing.T, uc UseCases, parent FID, name, content string) FID {
	t.Helper()
	before := map[FID]bool{}
	for _, fid := range entriesOf(t, uc, parent) {
		before[fid] = true
	}

	if err := uc.Put(user.SU(), parent, name, stringReader(content), PutOptions{Mode: 0600, OriginalFilename: name}); err != nil {
		t.Fatalf("put %q: %v", name, err)
	}

	for _, fid := range entriesOf(t, uc, parent) {
		if !before[fid] {
			return fid
		}
	}
	t.Fatalf("created file %q not found in parent %s", name, parent)
	return ""
}

// TestMoveHappyPath moves a file from one directory to another and verifies the structural changes.
func TestMoveHappyPath(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, err := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	dst, err := uc.MkDir(owner, root.ID, "dst", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir dst: %v", err)
	}

	fileID := putFile(t, uc, src.ID, "file.txt", "hello")

	if err := uc.Move(owner, fileID, dst.ID); err != nil {
		t.Fatalf("move: %v", err)
	}

	// removed from src, added to dst
	if containsFID(entriesOf(t, uc, src.ID), fileID) {
		t.Fatal("file still referenced by source directory")
	}
	if !containsFID(entriesOf(t, uc, dst.ID), fileID) {
		t.Fatal("file not referenced by destination directory")
	}

	// parent backreference updated, FID unchanged
	moved := statFile(t, uc, fileID)
	if moved.Parent != dst.ID {
		t.Fatalf("file parent not updated: got %s want %s", moved.Parent, dst.ID)
	}
	if moved.ID != fileID {
		t.Fatalf("file id must not change: got %s want %s", moved.ID, fileID)
	}
}

// TestMoveKeepsVersionHistoryAndACL verifies that a move preserves the blob version history and the per-file
// ACL grants (both are keyed by the FID, which is stable across a move).
func TestMoveKeepsVersionHistoryAndACL(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, _ := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	dst, _ := uc.MkDir(owner, root.ID, "dst", MkDirOptions{Mode: 0700})

	fileID := putFile(t, uc, src.ID, "file.txt", "v1")
	// add a second version
	if err := uc.Put(owner, src.ID, "file.txt", stringReader("v2"), PutOptions{Mode: 0600, OriginalFilename: "file.txt"}); err != nil {
		t.Fatalf("put v2: %v", err)
	}

	before := statFile(t, uc, fileID)
	versionsBefore := len(before.Versions())
	if versionsBefore < 2 {
		t.Fatalf("expected at least 2 versions before move, got %d", versionsBefore)
	}

	// grant a user + a group on the file
	if err := uc.GrantFileAccess(owner, fileID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant user: %v", err)
	}
	if err := uc.GrantFileAccess(owner, fileID, GroupGrantee("team"), PermsReader...); err != nil {
		t.Fatalf("grant group: %v", err)
	}

	grantsBefore, err := uc.ReadFileGrants(owner, fileID)
	if err != nil {
		t.Fatalf("read grants before: %v", err)
	}

	if err := uc.Move(owner, fileID, dst.ID); err != nil {
		t.Fatalf("move: %v", err)
	}

	after := statFile(t, uc, fileID)
	if got := len(after.Versions()); got != versionsBefore {
		t.Fatalf("version history changed by move: before %d after %d", versionsBefore, got)
	}

	grantsAfter, err := uc.ReadFileGrants(owner, fileID)
	if err != nil {
		t.Fatalf("read grants after: %v", err)
	}
	if len(grantsAfter) != len(grantsBefore) || len(grantsAfter) != 2 {
		t.Fatalf("acl grants changed by move: before %+v after %+v", grantsBefore, grantsAfter)
	}
}

// TestMoveRejectsCycle verifies that moving a directory into itself or one of its descendants is rejected.
func TestMoveRejectsCycle(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	a, _ := uc.MkDir(owner, root.ID, "a", MkDirOptions{Mode: 0700})
	b, _ := uc.MkDir(owner, a.ID, "b", MkDirOptions{Mode: 0700})
	c, _ := uc.MkDir(owner, b.ID, "c", MkDirOptions{Mode: 0700})

	// move a into its own descendant c -> cycle
	if err := uc.Move(owner, a.ID, c.ID); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("expected os.ErrInvalid for cycle move, got %v", err)
	}

	// move a into itself -> cycle
	if err := uc.Move(owner, a.ID, a.ID); err != nil {
		// a.ID == a.ID is caught by the fid==newParent guard, but Parent==newParent no-op also applies to
		// direct child moves; moving a into a is fid==newParent -> ErrInvalid
		if !errors.Is(err, os.ErrInvalid) {
			t.Fatalf("expected os.ErrInvalid for self move, got %v", err)
		}
	} else {
		t.Fatal("expected error moving a directory into itself")
	}

	// structure must be intact: b still under a, c still under b
	if !containsFID(entriesOf(t, uc, a.ID), b.ID) {
		t.Fatal("b should still be under a")
	}
	if !containsFID(entriesOf(t, uc, b.ID), c.ID) {
		t.Fatal("c should still be under b")
	}
}

// TestMoveRejectsNameCollision verifies that a move into a directory that already contains an entry with the
// same name is rejected.
func TestMoveRejectsNameCollision(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, _ := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	dst, _ := uc.MkDir(owner, root.ID, "dst", MkDirOptions{Mode: 0700})

	fileID := putFile(t, uc, src.ID, "dup.txt", "a")
	// destination already has a file with the same stored name
	_ = putFile(t, uc, dst.ID, "dup.txt", "b")

	if err := uc.Move(owner, fileID, dst.ID); !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected os.ErrExist for name collision, got %v", err)
	}

	// file must remain in src
	if !containsFID(entriesOf(t, uc, src.ID), fileID) {
		t.Fatal("file must remain in source on collision")
	}
}

// TestMoveRequiresWriteOnBothParents verifies the unix-style permission model: write on both the source and
// the destination directory is required.
func TestMoveRequiresWriteOnBothParents(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, _ := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	dst, _ := uc.MkDir(owner, root.ID, "dst", MkDirOptions{Mode: 0700})
	fileID := putFile(t, uc, src.ID, "file.txt", "hello")

	bob := fakeSubject{id: "bob", rdb: rdb}

	// bob can write only the source, not the destination
	if err := uc.GrantFileAccess(owner, src.ID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant src: %v", err)
	}
	if err := uc.Move(bob, fileID, dst.ID); err == nil {
		t.Fatal("expected permission denied when destination is not writable")
	}

	// now grant write on the destination too -> move succeeds
	if err := uc.GrantFileAccess(owner, dst.ID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant dst: %v", err)
	}
	if err := uc.Move(bob, fileID, dst.ID); err != nil {
		t.Fatalf("move with write on both parents: %v", err)
	}
	if !containsFID(entriesOf(t, uc, dst.ID), fileID) {
		t.Fatal("file should be in destination after successful move")
	}
}

// TestMoveWritesMovedAuditEntry verifies that a Moved audit entry with correct old/new parents is appended and
// exposed via LogEntry.Unwrap.
func TestMoveWritesMovedAuditEntry(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, _ := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	dst, _ := uc.MkDir(owner, root.ID, "dst", MkDirOptions{Mode: 0700})
	fileID := putFile(t, uc, src.ID, "file.txt", "hello")

	if err := uc.Move(owner, fileID, dst.ID); err != nil {
		t.Fatalf("move: %v", err)
	}

	file := statFile(t, uc, fileID)
	var found *Moved
	for entry := range file.AuditLog.All() {
		if entry.Moved.IsSome() {
			m := entry.Moved.Unwrap()
			found = &m
		}
		// ensure Unwrap dispatches Moved as an Activity
		if act, ok := entry.Unwrap(); ok {
			if mv, ok := act.(Moved); ok {
				if mv.NewParent != dst.ID {
					t.Fatalf("Moved activity new parent mismatch: %s", mv.NewParent)
				}
			}
		}
	}

	if found == nil {
		t.Fatal("no Moved audit entry found on the file")
	}
	if found.OldParent != src.ID || found.NewParent != dst.ID || found.FID != fileID {
		t.Fatalf("Moved entry has wrong fields: %+v (want old=%s new=%s fid=%s)", *found, src.ID, dst.ID, fileID)
	}
}

// TestMoveNoOpWhenSameParent verifies that moving into the current parent is a no-op and does not error.
func TestMoveNoOpWhenSameParent(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	src, _ := uc.MkDir(owner, root.ID, "src", MkDirOptions{Mode: 0700})
	fileID := putFile(t, uc, src.ID, "file.txt", "hello")

	before := statFile(t, uc, fileID)
	logLenBefore := before.AuditLog.Len()

	if err := uc.Move(owner, fileID, src.ID); err != nil {
		t.Fatalf("no-op move should not error: %v", err)
	}

	after := statFile(t, uc, fileID)
	if after.AuditLog.Len() != logLenBefore {
		t.Fatalf("no-op move must not append audit entries: before %d after %d", logLenBefore, after.AuditLog.Len())
	}
}

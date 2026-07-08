// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"testing"

	"go.wdy.de/nago/application/user"
)

// TestApplyStandardEntryOrderWithStaleEntry reproduces the situation where a directory references a child FID
// whose file no longer exists in the repository (a stale entry). Re-ordering the directory entries (as done by
// MkDir/Put/Move) must not panic and must not duplicate the stale reference.
func TestApplyStandardEntryOrderWithStaleEntry(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// create a child directory so root has a real entry
	child, err := uc.MkDir(owner, root.ID, "child", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir child: %v", err)
	}

	// make the entry stale: delete the child's record directly from the repo while keeping the parent's
	// Entries reference to it intact (i.e. do NOT go through the Delete use case which would detach it).
	if err := repo.DeleteByID(child.ID); err != nil {
		t.Fatalf("delete child record: %v", err)
	}

	// sanity: root still references the (now stale) child
	rootStat := statFile(t, uc, root.ID)
	if rootStat.Entries.Len() != 1 {
		t.Fatalf("expected root to still reference the stale child, got %d entries", rootStat.Entries.Len())
	}

	// this triggers applyStandardEntryOrder on root's entries. Before the fix this panicked on
	// optFile.Unwrap() of the stale (None) entry.
	if _, err := uc.MkDir(owner, root.ID, "fresh", MkDirOptions{Mode: 0700}); err != nil {
		t.Fatalf("mkdir fresh (reorder with stale entry): %v", err)
	}

	// the stale reference must still be present exactly once (not purged, not duplicated) and the new dir
	// must have been added.
	rootStat = statFile(t, uc, root.ID)
	counts := map[FID]int{}
	for fid := range rootStat.Entries.All() {
		counts[fid]++
	}

	if counts[child.ID] != 1 {
		t.Fatalf("stale child reference should appear exactly once, got %d", counts[child.ID])
	}

	// exactly two entries: the stale child and the freshly created directory
	if rootStat.Entries.Len() != 2 {
		t.Fatalf("expected 2 entries (stale child + fresh dir), got %d", rootStat.Entries.Len())
	}
}

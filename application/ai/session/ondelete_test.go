// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"errors"
	"slices"
	"testing"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

// OnDelete sees the session and each of its children, e.g. to release their provider files, and an error keeps
// the session, so that the files are never orphaned.
func TestOnDeleteSeesTheSessionAndItsChildren(t *testing.T) {
	var seen []ID
	var fail error
	repo := Repository(json.NewSloppyJSONRepository[Session, ID](mem.NewBlobStore(string(Namespace))))
	uc := NewUseCases(repo, newTestRDB(t), OnDelete(func(id ID) error {
		if fail != nil {
			return fail
		}
		seen = append(seen, id)
		return nil
	}))

	subject := user.SU()
	parent, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model")})
	if err != nil {
		t.Fatal(err)
	}

	child, err := uc.Create(subject, CreateOptions{Model: model.ID("fake-model"), ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}

	fail = errors.New("ledger unavailable")
	if err := uc.Delete(subject, parent.ID); err == nil {
		t.Fatal("expected the deletion to fail")
	}

	if opt, _ := repo.FindByID(parent.ID); opt.IsNone() {
		t.Fatal("a session whose release failed must be kept")
	}

	fail = nil
	if err := uc.Delete(subject, parent.ID); err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(seen, parent.ID) || !slices.Contains(seen, child.ID) {
		t.Fatalf("expected the parent and the child, got %v", seen)
	}
}

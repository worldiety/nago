// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"testing"

	"go.wdy.de/nago/application/permission"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

func TestDeletePublishesDeleted(t *testing.T) {
	repo := &datamem.Repository[User, ID]{}
	if err := repo.Save(User{ID: "u1", Email: "a@example.com"}); err != nil {
		t.Fatal(err)
	}

	bus := &syncBus{}
	del := NewDelete(bus, repo)
	admin := testSubject{id: "admin", perms: []permission.ID{PermDelete}}

	if err := del(testSubject{id: "nobody"}, "u1"); !errors.Is(err, PermissionDeniedErr) {
		t.Fatalf("expected permission denied, got %v", err)
	}

	if err := del(admin, "u1"); err != nil {
		t.Fatal(err)
	}

	// deleting twice is idempotent and announces nothing
	if err := del(admin, "u1"); err != nil {
		t.Fatal(err)
	}

	if len(bus.events) != 1 {
		t.Fatalf("expected exactly one event, got %v", bus.events)
	}

	evt, ok := bus.events[0].(Deleted)
	if !ok || evt.ID != "u1" || evt.Email != "a@example.com" || evt.DeletedAt.IsZero() {
		t.Fatalf("unexpected event %#v", bus.events[0])
	}

	if opt, _ := repo.FindByID("u1"); opt.IsSome() {
		t.Fatal("user must be deleted")
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role_test

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application/permission"
	. "go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
)

// TestSystemRoleCannotBeDeleted is the guarantee an operator relies on: a role a module declared stays put,
// so the feature behind it keeps working.
func TestSystemRoleCannotBeDeleted(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.system"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "System", System: true}); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	if err := uc.Delete(user.SU(), rid); err == nil {
		t.Error("a system role was deleted")
	}

	optRole, err := uc.FindByID(user.SU(), rid)
	if err != nil {
		t.Fatalf("cannot find role: %v", err)
	}

	if optRole.IsNone() {
		t.Error("the system role is gone even though Delete reported an error")
	}
}

// TestOrdinaryRoleStillDeletable makes sure the guard is narrow and does not turn every role into a permanent
// one.
func TestOrdinaryRoleStillDeletable(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.ordinary"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Ordinary"}); err != nil {
		t.Fatalf("cannot upsert role: %v", err)
	}

	if err := uc.Delete(user.SU(), rid); err != nil {
		t.Fatalf("an ordinary role could not be deleted: %v", err)
	}
}

// TestSystemRolePermissionsCannotBeReplaced covers the editing path of the admin UI.
func TestSystemRolePermissionsCannotBeReplaced(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.system.perms"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "System", System: true}); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	if err := uc.UpsertPermissions(user.SU(), rid, []permission.ID{PermFindAll}); err != nil {
		t.Fatalf("the declaring module could not grant its permissions: %v", err)
	}

	if err := uc.UpdatePermissions(user.SU(), rid, []permission.ID{PermDelete}); err == nil {
		t.Error("the permissions of a system role were replaced")
	}

	want := []permission.ID{PermFindAll}
	if got := listPermissions(t, uc, rid); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestUpdateCannotClearTheSystemFlag closes the obvious way around the two guards above: strip the flag with
// an ordinary Update, then delete the role.
func TestUpdateCannotClearTheSystemFlag(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.system.flag"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "System", System: true}); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	if err := uc.Update(user.SU(), Role{ID: rid, Name: "Renamed", System: false}); err != nil {
		t.Fatalf("cannot update role: %v", err)
	}

	optRole, err := uc.FindByID(user.SU(), rid)
	if err != nil {
		t.Fatalf("cannot find role: %v", err)
	}

	updated := optRole.Unwrap()

	if !updated.IsSystem() {
		t.Error("an ordinary update cleared the system flag")
	}

	// Renaming must still work, otherwise an operator cannot adapt the wording to their organisation.
	if updated.Name != "Renamed" {
		t.Errorf("got name %q, want %q", updated.Name, "Renamed")
	}
}

// TestUpdateCannotForgeTheSystemFlag is the mirror image: an ordinary role must not be able to promote itself
// into a protected one, which would make it undeletable through the UI.
func TestUpdateCannotForgeTheSystemFlag(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.forged"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Ordinary"}); err != nil {
		t.Fatalf("cannot upsert role: %v", err)
	}

	if err := uc.Update(user.SU(), Role{ID: rid, Name: "Ordinary", System: true}); err != nil {
		t.Fatalf("cannot update role: %v", err)
	}

	optRole, err := uc.FindByID(user.SU(), rid)
	if err != nil {
		t.Fatalf("cannot find role: %v", err)
	}

	if optRole.Unwrap().IsSystem() {
		t.Error("an ordinary update promoted a role to a system role")
	}
}

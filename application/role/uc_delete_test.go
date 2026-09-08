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
	"go.wdy.de/nago/application/rebac"
	. "go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
)

// newTestUseCases wires the role use cases against in-memory stores, mirroring the static rules that
// Configurator.RoleManagement registers for real (see application/management_role.go): a role may hold any
// declared permission globally, and a role may have users as members.
func newTestUseCases(t *testing.T) (UseCases, *rebac.DB) {
	t.Helper()

	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatalf("cannot create rebac db: %v", err)
	}

	for perm := range permission.All() {
		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   Namespace,
			Relation: rebac.Relation(perm.ID),
			Target:   rebac.Global,
		})
	}

	rdb.RegisterStaticRule(rebac.StaticRule{
		Source:   Namespace,
		Relation: rebac.Member,
		Target:   user.Namespace,
	})

	repo := Repository(json.NewSloppyJSONRepository[Role, ID](mem.NewBlobStore(string(Namespace))))

	return NewUseCases(repo, events.NewEventBus(), rdb), rdb
}

// permissionsOf collects the permissions currently granted to the role directly from the ReBAC database,
// bypassing the use cases so the test observes the stored state rather than what a reader chooses to report.
func permissionsOf(t *testing.T, rdb *rebac.DB, id ID) []permission.ID {
	t.Helper()

	var got []permission.ID
	for pid, err := range ListPermissionsFrom(rdb, id) {
		if err != nil {
			t.Fatalf("cannot list permissions of %q: %v", id, err)
		}
		got = append(got, pid)
	}

	slices.Sort(got)
	return got
}

// TestDeletePurgesGrantedPermissions pins down that deleting a role really removes the permissions it held.
//
// The permission grants are stored as triples with the role as the *source* (see NewUpdatePermissions),
// while Delete used to purge only triples with the role as the *target* - the membership side. The grants
// therefore survived the deletion as orphans.
func TestDeletePurgesGrantedPermissions(t *testing.T) {
	uc, rdb := newTestUseCases(t)

	const rid ID = "test.role.delete"
	granted := []permission.ID{PermFindAll, PermFindByID}

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Test"}); err != nil {
		t.Fatalf("cannot upsert role: %v", err)
	}

	if err := uc.UpdatePermissions(user.SU(), rid, granted); err != nil {
		t.Fatalf("cannot update permissions: %v", err)
	}

	if got := permissionsOf(t, rdb, rid); len(got) != len(granted) {
		t.Fatalf("precondition failed: expected %d permissions before delete, got %v", len(granted), got)
	}

	if err := uc.Delete(user.SU(), rid); err != nil {
		t.Fatalf("cannot delete role: %v", err)
	}

	if got := permissionsOf(t, rdb, rid); len(got) != 0 {
		t.Errorf("deleting the role left %d orphaned permission grants behind: %v", len(got), got)
	}
}

// TestDeleteDoesNotLeakPermissionsIntoARecreatedRole is the consequence that makes the orphaned triples
// dangerous rather than merely untidy: role ids may be chosen by the operator, so a role recreated under a
// previously used id silently inherits the permissions of its namesake.
func TestDeleteDoesNotLeakPermissionsIntoARecreatedRole(t *testing.T) {
	uc, _ := newTestUseCases(t)

	const rid ID = "test.role.recreated"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Privileged"}); err != nil {
		t.Fatalf("cannot upsert role: %v", err)
	}

	if err := uc.UpdatePermissions(user.SU(), rid, []permission.ID{PermDelete}); err != nil {
		t.Fatalf("cannot update permissions: %v", err)
	}

	if err := uc.Delete(user.SU(), rid); err != nil {
		t.Fatalf("cannot delete role: %v", err)
	}

	// The operator reuses the identifier for something entirely different and never grants it anything.
	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Harmless"}); err != nil {
		t.Fatalf("cannot recreate role: %v", err)
	}

	var got []permission.ID
	for pid, err := range uc.ListPermissions(user.SU(), rid) {
		if err != nil {
			t.Fatalf("cannot list permissions: %v", err)
		}
		got = append(got, pid)
	}

	if len(got) != 0 {
		t.Errorf("recreated role inherited %d permissions from its deleted namesake: %v", len(got), got)
	}
}

// TestDeleteRemovesMemberships covers the other half of the same defect. Memberships are stored as
// (role -> member -> user), so the role is the source here as well and the purge missed them too: a user
// stayed a member of a role that no longer existed, and regained it the moment the id was reused.
func TestDeleteRemovesMemberships(t *testing.T) {
	uc, rdb := newTestUseCases(t)

	const rid ID = "test.role.members"
	const uid user.ID = "test-user"

	if _, err := uc.Upsert(user.SU(), Role{ID: rid, Name: "Test"}); err != nil {
		t.Fatalf("cannot upsert role: %v", err)
	}

	membership := rebac.Triple{
		Source:   rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(rid)},
		Relation: rebac.Member,
		Target:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)},
	}

	if err := rdb.Put(membership); err != nil {
		t.Fatalf("cannot add member: %v", err)
	}

	if err := uc.Delete(user.SU(), rid); err != nil {
		t.Fatalf("cannot delete role: %v", err)
	}

	ok, err := rdb.Contains(membership)
	if err != nil {
		t.Fatalf("cannot query membership: %v", err)
	}

	if ok {
		t.Error("deleting the role left its membership triple behind")
	}
}

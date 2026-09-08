// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

func newTestConfigurator(t *testing.T) *Configurator {
	t.Helper()

	return &Configurator{
		dataDir:   t.TempDir(),
		factories: map[proto.RootViewID]func(wnd core.Window) core.View{},
		eventBus:  events.NewEventBus(),
	}
}

func permissionsOfRole(t *testing.T, c *Configurator, id role.ID) []permission.ID {
	t.Helper()

	roles, err := c.RoleManagement()
	if err != nil {
		t.Fatalf("cannot get role management: %v", err)
	}

	var got []permission.ID
	for pid, err := range roles.UseCases.ListPermissions(c.SysUser(), id) {
		if err != nil {
			t.Fatalf("cannot list permissions: %v", err)
		}
		got = append(got, pid)
	}

	slices.Sort(got)
	return got
}

// TestDeclareSystemRole covers the contract a module relies on when it ships its own role.
func TestDeclareSystemRole(t *testing.T) {
	c := newTestConfigurator(t)

	const rid role.ID = "test.system.role"
	declared := []permission.ID{role.PermFindAll, role.PermFindByID}

	if err := c.DeclareSystemRole(role.Role{ID: rid, Name: "Declared", Description: "shipped"}, declared...); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	roles, err := c.RoleManagement()
	if err != nil {
		t.Fatalf("cannot get role management: %v", err)
	}

	optRole, err := roles.UseCases.FindByID(c.SysUser(), rid)
	if err != nil || optRole.IsNone() {
		t.Fatalf("the declared role does not exist: %v", err)
	}

	if !optRole.Unwrap().IsSystem() {
		t.Error("the declared role is not marked as a system role")
	}

	want := slices.Clone(declared)
	slices.Sort(want)

	if got := permissionsOfRole(t, c, rid); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestDeclareSystemRoleKeepsOperatorWording pins down that a restart does not undo a rename. The role name is
// operator-facing configuration, and the admin UI deliberately keeps it editable for system roles.
func TestDeclareSystemRoleKeepsOperatorWording(t *testing.T) {
	c := newTestConfigurator(t)

	const rid role.ID = "test.system.renamed"

	if err := c.DeclareSystemRole(role.Role{ID: rid, Name: "Shipped Name"}); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	roles, err := c.RoleManagement()
	if err != nil {
		t.Fatalf("cannot get role management: %v", err)
	}

	if err := roles.UseCases.Update(c.SysUser(), role.Role{ID: rid, Name: "Operator Name"}); err != nil {
		t.Fatalf("cannot rename role: %v", err)
	}

	// The next application start re-declares the very same role.
	if err := c.DeclareSystemRole(role.Role{ID: rid, Name: "Shipped Name"}); err != nil {
		t.Fatalf("cannot re-declare system role: %v", err)
	}

	optRole, err := roles.UseCases.FindByID(c.SysUser(), rid)
	if err != nil || optRole.IsNone() {
		t.Fatalf("role vanished: %v", err)
	}

	if got := optRole.Unwrap().Name; got != "Operator Name" {
		t.Errorf("got name %q, want %q - re-declaring must not overwrite the operator's wording", got, "Operator Name")
	}

	if !optRole.Unwrap().IsSystem() {
		t.Error("re-declaring lost the system flag")
	}
}

// TestDeclareSystemRoleAddsNewPermissions is the upgrade case: a later release needs one more permission, and
// it has to arrive without the operator doing anything.
func TestDeclareSystemRoleAddsNewPermissions(t *testing.T) {
	c := newTestConfigurator(t)

	const rid role.ID = "test.system.upgraded"

	if err := c.DeclareSystemRole(role.Role{ID: rid, Name: "Upgraded"}, role.PermFindAll); err != nil {
		t.Fatalf("cannot declare system role: %v", err)
	}

	if err := c.DeclareSystemRole(role.Role{ID: rid, Name: "Upgraded"}, role.PermFindAll, role.PermFindByID); err != nil {
		t.Fatalf("cannot re-declare system role: %v", err)
	}

	want := []permission.ID{role.PermFindAll, role.PermFindByID}
	slices.Sort(want)

	if got := permissionsOfRole(t, c, rid); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestDeclareSystemRoleRequiresAnID guards against a caller that forgets the identifier, which would
// otherwise create a new random role on every single start.
func TestDeclareSystemRoleRequiresAnID(t *testing.T) {
	c := newTestConfigurator(t)

	if err := c.DeclareSystemRole(role.Role{Name: "Anonymous"}); err == nil {
		t.Error("a system role without an id was accepted")
	}
}

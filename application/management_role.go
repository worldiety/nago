// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"iter"
	"strings"

	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	uirole "go.wdy.de/nago/application/role/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/layout"
)

// RoleManagement is a nago system(Role Management).
// It provides UseCases for creating, editing, deleting roles.
// Roles are used to grant users bundled permissions.
// They can be created, edited, and deleted via UI or code.
// Roles assignment is managed through UserManagement.
type RoleManagement struct {
	UseCases       role.UseCases
	Pages          uirole.Pages
	roleRepository role.Repository
}

func (c *Configurator) RoleManagement() (RoleManagement, error) {
	if c.roleManagement == nil {
		roleStore, err := c.EntityStore(string(role.Namespace))
		if err != nil {
			return RoleManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		mg, err := c.Migrations()
		if err != nil {
			return RoleManagement{}, fmt.Errorf("cannot get migrations: %w", err)
		}
		rdb, err := c.RDB()
		if err != nil {
			return RoleManagement{}, fmt.Errorf("cannot get rdb: %w", err)
		}

		// we have permissions which are generated and registered any time later at runtime, which must be generally allowed to be assigned
		permission.OnPermissionRegistered(func(permission permission.Permission) {
			rdb.RegisterStaticRule(rebac.StaticRule{
				Source:   role.Namespace,
				Relation: rebac.Relation(permission.ID),
				Target:   rebac.Global,
			})
		})

		for perm := range permission.All() {
			rdb.RegisterStaticRule(rebac.StaticRule{
				Source:   role.Namespace,
				Relation: rebac.Relation(perm.ID),
				Target:   rebac.Global,
			})
		}

		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   role.Namespace,
			Relation: rebac.Member,
			Target:   user.Namespace,
		})

		// implementation note: it is important to first apply all user migrations, otherwise
		// we may risk data loss due to missing fields in current user entities
		if err := mg.Declare(newMigrateRolePermsToReBAC(roleStore, rdb), migration.Options{Immediate: true}); err != nil {
			return RoleManagement{}, fmt.Errorf("cannot declare migration: %w", err)
		}

		roleRepo := json.NewSloppyJSONRepository[role.Role, role.ID](roleStore)

		c.roleManagement = &RoleManagement{
			roleRepository: roleRepo,
			UseCases:       role.NewUseCases(roleRepo, c.EventBus(), rdb),
			Pages: uirole.Pages{
				Roles: "admin/iam/roles",
				Role:  "admin/iam/roles/role",
			},
		}

		rdb.RegisterResources(c.roleManagement.UseCases.Resources)

		c.RootView(c.roleManagement.Pages.Roles, c.DecorateRootView(func(wnd core.Window) core.View {
			return layout.WithBackButton(wnd, uirole.PageRoles(wnd, c.roleManagement.Pages, c.roleManagement.UseCases))
		}))

		c.RootView(c.roleManagement.Pages.Role, c.DecorateRootView(func(wnd core.Window) core.View {
			return layout.WithBackButton(wnd, uirole.PageRole(wnd, c.roleManagement.Pages, c.roleManagement.UseCases))
		}))

		c.AddContextValue(core.ContextValue("nago.roles", form.AnyUseCaseList[role.Role, role.ID](func(subject auth.Subject) iter.Seq2[role.Role, error] {
			return c.roleManagement.UseCases.FindAll(subject)
		})))

		c.AddContextValue(core.ContextValue("nago.roles.find_by_id", c.roleManagement.UseCases.FindByID))
	}

	return *c.roleManagement, nil
}

// DeclareSystemRole ensures that the given role exists, is marked as a system role and holds at least the
// given permissions. It is the way a module ships the authorization its feature needs: instead of documenting
// "grant these three permissions to somebody", the module declares one role and an operator only has to
// assign it.
//
// The declaration is authoritative for what makes the role work - its existence, the System flag and the
// listed permissions - and for nothing else:
//
//   - Name and description are written only when the role is created. Afterwards they belong to the operator,
//     who may adapt the wording to their organisation without a restart undoing it.
//   - Permissions granted in addition are kept, because the declaration states a minimum, not a set
//     (see [role.UpsertPermissions]).
//
// When the role already exists and already holds the permissions, no write is performed at all, so this is
// cheap enough to call unconditionally on every start.
//
// A system role is protected: it can neither be deleted nor have its permissions replaced through the admin
// UI, since both would be undone here on the next start. Assigning members - the entire point of the role -
// is unaffected.
//
// Call it after the permissions in question have been declared, which for package level [permission.Declare]
// variables is always the case. It is idempotent, so calling it repeatedly is harmless.
func (c *Configurator) DeclareSystemRole(r role.Role, permissions ...permission.ID) error {
	if strings.TrimSpace(string(r.ID)) == "" {
		return fmt.Errorf("a system role must have a stable id")
	}

	roles, err := c.RoleManagement()
	if err != nil {
		return fmt.Errorf("cannot get role management: %w", err)
	}

	sys := c.SysUser()

	optExisting, err := roles.UseCases.FindByID(sys, r.ID)
	if err != nil {
		return fmt.Errorf("cannot look up system role %q: %w", r.ID, err)
	}

	if existing := optExisting.UnwrapOr(role.Role{}); optExisting.IsSome() {
		// Keep whatever the operator made of the wording; only re-assert the flag, and only if it is missing.
		if existing.IsSystem() {
			r = existing
		} else {
			r = existing
			r.System = true
			if _, err := roles.UseCases.Upsert(sys, r); err != nil {
				return fmt.Errorf("cannot mark role %q as a system role: %w", r.ID, err)
			}
		}
	} else {
		r.System = true
		if _, err := roles.UseCases.Upsert(sys, r); err != nil {
			return fmt.Errorf("cannot create system role %q: %w", r.ID, err)
		}
	}

	if err := roles.UseCases.UpsertPermissions(sys, r.ID, permissions); err != nil {
		return fmt.Errorf("cannot grant permissions to system role %q: %w", r.ID, err)
	}

	return nil
}

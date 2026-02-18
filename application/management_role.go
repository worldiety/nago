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

	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/role"
	uirole "go.wdy.de/nago/application/role/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
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
		// implementation note: it is important to first apply all user migrations, otherwise
		// we may risk data loss due to missing fields in current user entities
		if err := mg.Declare(newMigrateRolePermsToReBAC(roleStore, rdb), migration.Options{Immediate: true}); err != nil {
			return RoleManagement{}, fmt.Errorf("cannot declare migration: %w", err)
		}

		roleRepo := json.NewSloppyJSONRepository[role.Role, role.ID](roleStore)

		c.roleManagement = &RoleManagement{
			roleRepository: roleRepo,
			UseCases:       role.NewUseCases(roleRepo, c.EventBus(), rdb),
			Pages:          uirole.Pages{Roles: "admin/iam/role"},
		}

		// note the bootstrap case and polymorphic adaption to the auditable
		c.RootView(c.roleManagement.Pages.Roles, c.DecorateRootView(func(wnd core.Window) core.View {

			return uirole.GroupPage(wnd, rcrud.UseCasesFrom(&rcrud.Funcs[role.Role, role.ID]{
				PermFindByID:   role.PermFindByID,
				PermFindAll:    role.PermFindAll,
				PermDeleteByID: role.PermDelete,
				PermCreate:     role.PermCreate,
				PermUpdate:     role.PermUpdate,
				FindByID: func(subject auth.Subject, id role.ID) (std.Option[role.Role], error) {
					return c.roleManagement.UseCases.FindByID(subject, id)
				},
				FindAll: func(subject auth.Subject) iter.Seq2[role.Role, error] {
					return c.roleManagement.UseCases.FindAll(subject)
				},
				DeleteByID: func(subject auth.Subject, id role.ID) error {
					return c.roleManagement.UseCases.Delete(subject, id)
				},
				Create: func(subject auth.Subject, entity role.Role) (role.ID, error) {
					return c.roleManagement.UseCases.Create(subject, entity)
				},
				Update: func(subject auth.Subject, entity role.Role) error {
					return c.roleManagement.UseCases.Update(subject, entity)
				},
				Upsert: func(subject auth.Subject, entity role.Role) (role.ID, error) {
					return c.roleManagement.UseCases.Upsert(subject, entity)
				},
			}))
		}))

		c.AddContextValue(core.ContextValue("nago.roles", form.AnyUseCaseList[role.Role, role.ID](func(subject auth.Subject) iter.Seq2[role.Role, error] {
			return c.roleManagement.UseCases.FindAll(subject)
		})))

		c.AddContextValue(core.ContextValue("nago.roles.find_by_id", c.roleManagement.UseCases.FindByID))
	}

	return *c.roleManagement, nil
}

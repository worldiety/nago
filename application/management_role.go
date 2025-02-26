package application

import (
	"fmt"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/role"
	uirole "go.wdy.de/nago/application/role/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
	"iter"
)

type RoleManagement struct {
	UseCases       role.UseCases
	Pages          uirole.Pages
	roleRepository role.Repository
}

func (c *Configurator) RoleManagement() (RoleManagement, error) {
	if c.roleManagement == nil {
		roleStore, err := c.EntityStore("nago.iam.role")
		if err != nil {
			return RoleManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		roleRepo := json.NewSloppyJSONRepository[role.Role, role.ID](roleStore)

		c.roleManagement = &RoleManagement{
			roleRepository: roleRepo,
			UseCases:       role.NewUseCases(roleRepo),
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

		c.AddSystemService("nago.roles", crud.AnyUseCaseList[role.Role, role.ID](func(subject auth.Subject) iter.Seq2[role.Role, error] {
			return c.roleManagement.UseCases.FindAll(subject)
		}))
	}

	return *c.roleManagement, nil
}

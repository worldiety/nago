package application

import (
	"fmt"
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
			return uirole.GroupPage(wnd, crud.UseCasesFromFuncs(
				func(subject auth.Subject, id role.ID) (std.Option[role.Role], error) {
					return c.roleManagement.UseCases.FindByID(subject, id)
				},
				func(subject auth.Subject) iter.Seq2[role.Role, error] {
					return c.roleManagement.UseCases.FindAll(subject)
				},
				func(subject auth.Subject, id role.ID) error {
					return c.roleManagement.UseCases.Delete(subject, id)
				},
				func(subject auth.Subject, entity role.Role) (role.ID, error) {
					return c.roleManagement.UseCases.Upsert(subject, entity)
				},
			))
		}))
	}

	return *c.roleManagement, nil
}

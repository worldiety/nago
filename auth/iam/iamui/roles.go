package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
)

func Roles(wnd core.Window, modals ui.ModalOwner, service *iam.Service) core.View {
	subject := wnd.Subject()
	return crud.NewView(modals, crud.NewOptions[iam.Role](func(opts *crud.Options[iam.Role]) {
		opts.Title("Rollen")
		opts.ReadAll(service.AllRoles(subject))
		opts.Create(func(role iam.Role) error {
			return service.CreateRole(subject, role)
		})
		opts.Delete(func(role iam.Role) error {
			return service.DeleteRole(subject, role.ID)
		})
		opts.Update(func(role iam.Role) error {
			return service.UpdateRole(subject, role)
		})

		opts.Responsive(wnd)
		opts.Bind(func(bnd *crud.Binding[iam.Role]) {

			crud.Text(bnd, crud.FromPtr("ID", func(model *iam.Role) *auth.RID {
				return &model.ID
			}, crud.RenderHints{
				crud.Overview: crud.Hidden,
				crud.Update:   crud.ReadOnly,
				crud.Create:   crud.Visible,
			}))

			crud.Text(bnd, crud.FromPtr("Name", func(model *iam.Role) *iam.PID {
				return &model.Name
			}))

			crud.Text(bnd, crud.FromPtr("Beschreibung", func(model *iam.Role) *iam.PID {
				return &model.Description
			}))

			crud.OneToMany[iam.Role, iam.Permission, iam.PID](bnd,
				service.AllPermissions(subject),
				func(permission iam.Permission) string {
					return permission.Name()
				},
				crud.FromPtr("Berechtigungen", func(model *iam.Role) *[]iam.PID {
					return &model.Permissions
				}),
			)
		})
	}))
}

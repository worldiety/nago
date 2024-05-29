package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
)

func Roles(subject auth.Subject, modals ui.ModalOwner, service *iam.Service) core.Component {

	return crud.NewView(modals, crud.NewOptions[iam.Role](func(opts *crud.Options[iam.Role]) {
		opts.Title = "Rollen"
		opts.FindAll = service.AllRoles(subject)
		opts.Create = func(role iam.Role) error {
			return service.CreateRole(subject, role)
		}
		opts.OnDelete(func(role iam.Role) error {
			return service.DeleteRole(subject, role.ID)
		})
		opts.OnUpdate(func(role iam.Role) error {
			return service.UpdateRole(subject, role)
		})
		opts.Binding = crud.NewBinding[iam.Role](func(bnd *crud.Binding[iam.Role]) {
			crud.Text(bnd, crud.Field[iam.Role, string]{
				Caption: "ID",
				Stringer: func(role iam.Role) string {
					return string(role.ID)
				},
				IntoModel: func(model iam.Role, value string) (iam.Role, error) {
					model.ID = auth.RID(value)
					return model, nil
				},
				RenderHints: crud.RenderHints{
					crud.Overview: crud.Hidden,
					crud.Update:   crud.ReadOnly,
					crud.Create:   crud.Visible,
				},
			})

			crud.Text(bnd, crud.Field[iam.Role, string]{
				Caption: "Name",
				Stringer: func(role iam.Role) string {
					return role.Name
				},
				IntoModel: func(model iam.Role, value string) (iam.Role, error) {
					model.Name = value
					return model, nil
				},
			})

			crud.OneToMany[iam.Role, iam.Permission, iam.PID](bnd, service.AllPermissions(subject), func(permission iam.Permission) string {
				return permission.Name()
			}, crud.Field[iam.Role, []iam.PID]{
				Caption: "Berechtigungen",
				FromModel: func(role iam.Role) []iam.PID {
					return role.Permissions
				},
				IntoModel: func(model iam.Role, value []iam.PID) (iam.Role, error) {
					model.Permissions = value
					return model, nil
				},
			})
		})
	}))
}

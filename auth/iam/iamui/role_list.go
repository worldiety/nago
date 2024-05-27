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
		opts.Delete(func(role iam.Role) error {
			return service.DeleteRole(subject, role.ID)
		})
		opts.Binding = crud.NewBinding[iam.Role](func(bnd *crud.Binding[iam.Role]) {
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
		})
	}))
}

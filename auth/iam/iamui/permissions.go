package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
)

func Permissions(owner ui.ModalOwner, subject auth.Subject, service *iam.Service) core.Component {
	return crud.NewView(owner, crud.NewOptions[iam.Permission](func(opts *crud.Options[iam.Permission]) {
		opts.Title = "Berechtigungen"
		opts.FindAll = service.AllPermissions(subject)
		opts.Binding = crud.NewBinding(func(bnd *crud.Binding[iam.Permission]) {
			crud.Text(bnd, crud.FromPtr("Name", func(model *iam.Permission) *string {
				tmp := (*model).Name()
				return &tmp
			}))
			crud.Text(bnd, crud.FromPtr("Beschreibung", func(model *iam.Permission) *string {
				tmp := (*model).Desc()
				return &tmp
			}))
		})
	}))
}

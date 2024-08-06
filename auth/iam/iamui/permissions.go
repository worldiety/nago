package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/crud"
)

func Permissions(wnd core.Window, owner uilegacy.ModalOwner, service *iam.Service) core.View {
	subject := wnd.Subject()
	return crud.NewView(owner, crud.NewOptions[iam.Permission](func(opts *crud.Options[iam.Permission]) {
		opts.Title("Berechtigungen")
		opts.ReadAll(service.AllPermissions(subject))
		opts.Responsive(wnd)
		opts.Bind(func(bnd *crud.Binding[iam.Permission]) {
			crud.Text(bnd, crud.FromPtr("ID", func(model *iam.Permission) *string {
				tmp := (*model).Identity()
				return &tmp
			}))
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

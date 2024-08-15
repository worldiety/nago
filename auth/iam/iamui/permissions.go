package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func Permissions(wnd core.Window, service *iam.Service) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[iam.Permission](wnd)
	bnd.Add(
		crud.Text("ID", func(model *iam.Permission) *string {
			tmp := (*model).Identity()
			return &tmp
		}),
		crud.Text("Name", func(model *iam.Permission) *string {
			tmp := (*model).Name()
			return &tmp
		}),
		crud.Text("Beschreibung", func(model *iam.Permission) *string {
			tmp := (*model).Desc()
			return &tmp
		}),
	)

	opts := crud.Options(bnd).
		FindAll(service.AllPermissions(subject))

	return crud.View(opts)

}

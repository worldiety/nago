package uipermission

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func Permissions(wnd core.Window, all permission.FindAll) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[permission.Permission](wnd)
	bnd.Add(
		crud.Text(crud.TextOptions{Label: "EID"}, crud.Ptr(func(model *permission.Permission) *permission.ID {
			tmp := (*model).Identity() // protect against changes through defensive copy, just to remember
			return &tmp
		})),
		crud.Text(crud.TextOptions{Label: "Name"}, crud.Ptr(func(model *permission.Permission) *string {
			tmp := (*model).Name
			return &tmp
		})),
		crud.Text(crud.TextOptions{Label: "Beschreibung"}, crud.Ptr(func(model *permission.Permission) *string {
			tmp := (*model).Description
			return &tmp
		})),
	)

	opts := crud.Options(bnd).
		FindAll(all(subject))

	return crud.View(opts)

}

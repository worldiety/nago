package uipermission

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/presentation/ui/list"
)

type Pages struct {
	Permissions core.NavigationPath
}

func Permissions(wnd core.Window, all permission.FindAll) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[permission.Permission](wnd)
	// we need to keep this binding, otherwise the quick search won't work
	bnd.Add(
		crud.Text(crud.TextOptions{Label: "ID"}, crud.Ptr(func(model *permission.Permission) *permission.ID {
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

	bnd.IntoListEntry(func(entity permission.Permission) list.TEntry {
		return list.Entry().
			Leading(ui.ImageIcon(heroOutline.ShieldCheck)).
			Headline(entity.String()).
			SupportingText(entity.Description)
	})

	opts := crud.Options(bnd).
		Title("Berechtigungen").
		DisableDefaultSorting().
		ViewStyle(crud.ViewStyleListOnly).
		FindAll(all(subject))

	return crud.View(opts).Frame(ui.Frame{}.FullWidth())

}

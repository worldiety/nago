package uirole

import (
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

type Pages struct {
	Roles core.NavigationPath
}

func GroupPage(wnd core.Window, useCases rcrud.UseCases[role.Role, role.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Rollen"}, useCases)(wnd)
}

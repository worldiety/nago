package uirole

import (
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

type Pages struct {
	Roles core.NavigationPath
}

// TODO crud useCases do not distinguish between update and create which causes wrong entity creation if ID is modifiable
func GroupPage(wnd core.Window, useCases crud.UseCases[role.Role, role.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Rollen"}, useCases)(wnd)
}

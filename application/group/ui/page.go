package uigroup

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func GroupPage(wnd core.Window, useCases crud.UseCases[group.Group, group.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Gruppen"}, useCases)(wnd)
}

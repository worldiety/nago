package uiusercircles

import (
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func PageOverview(wnd core.Window, useCases rcrud.UseCases[usercircle.Circle, usercircle.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{}, useCases)(wnd)
}

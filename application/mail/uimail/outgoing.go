package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func OutgoingQueuePage(wnd core.Window, useCases crud.UseCases[mail.Outgoing, mail.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{CanCreate: false}, useCases)(wnd)
}

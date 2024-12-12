package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func OutgoingQueuePage(wnd core.Window, useCases mail.UseCases) core.View {
	cruds := crud.UseCasesFromFuncs(
		useCases.Outgoing.FindByID,
		useCases.Outgoing.FindAll,
		useCases.Outgoing.DeleteByID,
		useCases.Outgoing.Save,
	)
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Warteschlange Ausgang", CreateDisabled: true}, cruds)(wnd)
}

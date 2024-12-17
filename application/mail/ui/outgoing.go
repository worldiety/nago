package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func OutgoingQueuePage(wnd core.Window, useCases mail.UseCases) core.View {
	cruds := rcrud.UseCasesFrom(
		&rcrud.Funcs[mail.Outgoing, mail.ID]{
			PermFindByID:   mail.PermOutgoingFindByID,
			PermFindAll:    mail.PermOutgoingFindAll,
			PermDeleteByID: mail.PermOutgoingDeleteByID,
			PermCreate:     "",
			PermUpdate:     "",
			FindByID:       useCases.Outgoing.FindByID,
			FindAll:        useCases.Outgoing.FindAll,
			DeleteByID:     useCases.Outgoing.DeleteByID,
			Create:         nil,
			Update:         nil,
			Upsert:         nil,
		},
	)
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Warteschlange Ausgang", CreateDisabled: true}, cruds)(wnd)
}

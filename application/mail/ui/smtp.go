package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func SmtpPage(wnd core.Window, useCases mail.UseCases) core.View {
	cruds := rcrud.UseCasesFrom(
		&rcrud.Funcs[mail.Smtp, mail.SmtpID]{
			PermFindByID:   mail.PermSmtpFindByID,
			PermFindAll:    mail.PermSmtpFindAll,
			PermDeleteByID: mail.PermOutgoingDeleteByID,
			PermCreate:     "",
			PermUpdate:     "",
			FindByID:       useCases.Smtp.FindByID,
			FindAll:        useCases.Smtp.FindAll,
			DeleteByID:     useCases.Smtp.DeleteByID,
			Create:         useCases.Smtp.Create,
			Update:         useCases.Smtp.Update,
			Upsert:         nil,
		},
	)
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "SMTP Server"}, cruds)(wnd)
}

package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func SmtpPage(wnd core.Window, useCases mail.UseCases) core.View {
	cruds := crud.UseCasesFromFuncs(
		useCases.Smtp.FindByID,
		useCases.Smtp.FindAll,
		useCases.Smtp.DeleteByID,
		useCases.Smtp.Save,
	)
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "SMTP Server"}, cruds)(wnd)
}

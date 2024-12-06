package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func SmtpPage(wnd core.Window, useCases crud.UseCases[mail.Smtp, mail.SmtpID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "SMTP Server"}, useCases)(wnd)
}

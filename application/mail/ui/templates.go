package uimail

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func TemplatePage(wnd core.Window, useCases rcrud.UseCases[mail.Template, mail.TemplateID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "E-Mail Vorlagen"}, useCases)(wnd)
}

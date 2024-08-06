package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.Component(".", func(wnd core.Window) core.View {
			return uilegacy.NewPage(func(page *uilegacy.Page) {
				page.Body().Set(uilegacy.NewHStack(func(hstack *uilegacy.FlexContainer) {
					hstack.Append(
						uilegacy.NewActionButton("input text box", func() {
							var model string
							xdialog.InputBoxText(page, xdialog.InputBoxOptions[string]{
								Title:  "Eingabe erwartet",
								Prompt: "Dein Name:",
								Value:  &model,
								OnConfirmed: func() {
									xdialog.ShowMessage(page, fmt.Sprintf("Hallo %s", model))
								},
								OnDismissed: func() {
									xdialog.ShowMessage(page, fmt.Sprintf("Das war leider nichts %s", model))
								},
							})
						}),

						uilegacy.NewActionButton("input num box", func() {
							model := 12
							xdialog.InputBoxInt(page, xdialog.InputBoxOptions[int]{
								Prompt: "Dein Alter - no way out:",
								Value:  &model,
								OnConfirmed: func() {
									xdialog.ShowMessage(page, fmt.Sprintf("Alter %d", model))
								},
							})
						}),

						uilegacy.NewActionButton("error", func() {
							xdialog.RequestSupport(wnd, page, fmt.Errorf("internal mysql error, your password is 12345-super-secret-which must never leak into the frontend"))
						}),
					)
				}))
			})

		})
	}).Run()
}

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewVStack(func(viewport *ui.VStack) {
				viewport.Frame().With(func(fr *ora.Frame) {
					fr.Width = ora.Full
					//fr.Height = ora.ViewportHeight
				})

				viewport.Append(ui.MakeText(("HStack alignments")))
				for _, alignment := range ora.Alignments() {
					viewport.Append(ui.NewHStack(func(hstack *ui.HStack) {
						hstack.SetBackgroundColor(ora.ExplicitColor("#8cb4f5"))
						hstack.Frame().Width = ora.Absolute(200)
						hstack.Frame().Height = ora.Absolute(200)
						hstack.SetAlignment(alignment)
						hstack.Append(ui.NewText(func(text *ui.Text) {
							text.SetBackgroundColor("#aa0000")
							text.SetColor("#ffffff")
							text.Value().Set(alignment.String())
						}))
					}))

					viewport.Append(ui.NewDivider(nil))
				}

				viewport.Append(ui.MakeText(("VStack alignments")))
				for _, alignment := range ora.Alignments() {
					viewport.Append(ui.NewVStack(func(vstack *ui.VStack) {
						vstack.SetBackgroundColor(ora.ExplicitColor("#8cb4f5"))
						vstack.Frame().Width = ora.Absolute(200)
						vstack.Frame().Height = ora.Absolute(200)
						vstack.SetAlignment(alignment)
						vstack.Append(ui.NewText(func(text *ui.Text) {
							text.SetBackgroundColor("#aa0000")
							text.SetColor("#ffffff")
							text.Value().Set(alignment.String())
						}))
					}))

					viewport.Append(ui.NewDivider(nil))
				}

			})

		})
	}).Run()
}

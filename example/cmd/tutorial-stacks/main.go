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
			return ui.NewHStack(func(viewport *ui.HStack) {
				viewport.Frame().With(func(fr *ora.Frame) {
					fr.Width = ora.Full
					//fr.Height = ora.ViewportHeight
				})
				viewport.Append(ui.NewHStack(func(cols *ui.HStack) {
					cols.SetGap(ora.Absolute(2))
					cols.Append(
						ui.NewBox(func(box *ui.Box) {
							box.SetBackgroundColor(ora.ExplicitColor("#a8eda1"))
							box.Frame().Width = ora.Absolute(300)
							box.Frame().Height = ora.Absolute(300)
							box.Padding.All(ora.Absolute(4))
							for _, alignment := range ora.Alignments() {
								box.Align(alignment, ui.NewText(func(text *ui.Text) {
									text.SetBackgroundColor("#aa0000")
									text.SetColor("#ffffff")
									text.Padding.All(ora.Absolute(4))
									text.Frame.Width = ora.Absolute(64)
									text.Frame.Height = ora.Absolute(64)
									text.Value().Set(alignment.String())
								}))
							}

						}),
						ui.NewVStack(func(viewport *ui.VStack) {
							viewport.Append(ui.MakeText(("HStack alignments")))
							viewport.SetGap(ora.Absolute(2))

							viewport.Children().AppendFrom(func(yield func(core.Component) bool) {
								for _, alignment := range ora.Alignments() {
									yield(ui.NewHStack(func(hstack *ui.HStack) {
										hstack.SetBackgroundColor(ora.ExplicitColor("#8cb4f5"))
										hstack.Frame().Width = ora.Absolute(200)
										hstack.Frame().Height = ora.Absolute(200)
										hstack.SetAlignment(alignment)
										hstack.Append(ui.NewText(func(text *ui.Text) {
											text.SetBackgroundColor("#aa0000")
											text.SetColor("#ffffff")
											text.Value().Set(alignment.String())
										}))
										hstack.Append(ui.NewText(func(text *ui.Text) {
											text.SetBackgroundColor("#aa0000")
											text.SetColor("#ffffff")
											text.Value().Set(alignment.String())
										}))
									}))

								}
							})

						}),

						ui.NewVStack(func(viewport *ui.VStack) {
							viewport.Append(ui.MakeText(("VStack alignments")))
							viewport.SetGap(ora.Absolute(2))
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
									vstack.Append(ui.NewText(func(text *ui.Text) {
										text.SetBackgroundColor("#aa0000")
										text.SetColor("#ffffff")
										text.Value().Set(alignment.String())
									}))
								}))

							}
						}),
					)
				}))

			})

		})
	}).Run()
}

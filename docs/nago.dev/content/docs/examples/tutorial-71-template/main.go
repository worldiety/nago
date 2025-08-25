// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/example/cmd/tutorial-71-template/typst"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_71")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.PrimaryButton(func() {
					buf, err := template.Apply(user.SU(), typst.Template, template.TypstPDF, template.ExecOptions{Model: typst.ExampleData()})
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.ExportFiles(core.ExportFileBytes("doc.pdf", buf))
				}).Title("Run typst"),
			).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

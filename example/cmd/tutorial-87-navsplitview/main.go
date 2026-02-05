// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/navsplitview"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_87")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("navsplitview demo"),
				navsplitview.TwoColumn(navsplitview.NavLinks{
					"a": ui.VStack(
						ui.Text("content"),
						navsplitview.ListItem(navsplitview.KindDetail, "detail_1", ui.Text("punkt 1")),
						navsplitview.ListItem(navsplitview.KindDetail, "detail_2", ui.Text("punkt 2")),
					).FullWidth(),
					"none": ui.VStack(
						ui.Text("nichts gewählt"),
					),
					"detail_1": ui.Text("detail 1"),
					"detail_2": ui.Text("detail 2"),
				}).
					Default("a", "none").
					FullWidth(),

				ui.Space(ui.L48),
				
				// note the ID which is used to identify the navsplitview and prepended in the links
				// also note that you can use a factory or lazy evaluated view factories
				ui.Text("navsplitview demo 2"),
				navsplitview.TwoColumn(navsplitview.NavFnLinks{
					"_": func(id navsplitview.ViewID) core.View {
						return ui.VStack(
							ui.HStack(ui.Text("A section").Color(ui.ColorIconsMuted)).FullWidth(),
							navsplitview.ListItem(navsplitview.KindDetail, "detail_1", ui.Text("lazy punkt 1")).Prefix("nsv2"),
							ui.HLine(),
							navsplitview.ListItem(navsplitview.KindDetail, "detail_2", ui.Text("lazy punkt 2")).Prefix("nsv2"),
						).FullWidth()
					},
					"none": func(id navsplitview.ViewID) core.View {
						return ui.VStack(
							ui.Text("lazy nichts gewählt"),
						)
					},
					"detail_1": func(id navsplitview.ViewID) core.View {
						return ui.Text("lazy detail 1")
					},
					"detail_2": func(id navsplitview.ViewID) core.View {
						return ui.Text("lazy detail 2")
					},
				}).ID("nsv2").
					Default("_", "none").
					FullWidth(),
			).FullWidth()
		})

	}).Run()
}

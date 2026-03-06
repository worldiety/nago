// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/web/ssr/cfgssr"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_91")
		cfg.Serve(vuejs.Dist())

		var buttons []string
		for i := range 10000 {
			buttons = append(buttons, fmt.Sprintf("List entry %d", i+1))
		}

		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				list.List(
					ui.ForEach(buttons, func(b string) core.View {
						return list.Entry().Headline(b)
					})...,
				),
			).FullWidth().Alignment(ui.Center)
		})

		cfg.RootViewWithDecoration("test/example", func(wnd core.Window) core.View {
			strState := core.AutoState[string](wnd).Init(func() string {
				return "Torben"
			})
			return ui.VStack(
				ui.TextField("firstname", strState.Get()).InputValue(strState),
				ui.Text("hello text"),
				ui.PrimaryButton(func() {}).Title("primary").HRef("https://www.worldiety.de"),
			)
		})

		cfgssr.Enable(cfg, ".")
		cfgssr.Enable(cfg, "test/example")

	}).Run()
}

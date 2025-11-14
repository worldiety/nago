// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	_ "go.wdy.de/nago/application/ai/provider/mistralai"
	_ "go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_79")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			dropped := core.AutoState[string](wnd).Observe(func(newValue string) {
				fmt.Printf("dropped %v\n", newValue)
			})

			return ui.VStack(
				ui.HStack(
					ui.DnDArea(ui.VStack(ui.Text("A")).
						BackgroundColor(ui.ColorSemanticWarn).
						Frame(ui.Frame{}.Size(ui.L64, ui.L64)),
					).ID("a").
						CanDrag(true),

					ui.DnDArea(ui.VStack(ui.Text("B")).
						BackgroundColor(ui.ColorSemanticWarn).
						Frame(ui.Frame{}.Size(ui.L64, ui.L64)),
					).ID("b").
						CanDrag(true),

					ui.DnDArea(ui.VStack(ui.Text("C")).
						BackgroundColor(ui.ColorSemanticWarn).
						Frame(ui.Frame{}.Size(ui.L64, ui.L64)),
					).ID("c").
						CanDrag(true),
				),

				ui.DnDArea(
					ui.VStack(ui.Text("drop zone")).
						BackgroundColor(ui.ColorSemanticError).
						Frame(ui.Frame{}.Size(ui.L120, ui.L120)),
				).ID("b").
					Droppable("a", "c").
					CanDrop(true).
					InputValue(dropped).
					OnDropped(func() {
						// TODO weg
						fmt.Println("dropped", dropped.Get())
					}),
			).Frame(ui.Frame{}.Large())
		})
	}).Run()
}

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
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_58")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			text := core.AutoState[string](wnd).Init(func() string {
				return "hello rich text world!"
			}).Observe(func(newValue string) {
				fmt.Println("received:", newValue)
			})
			return ui.VStack(
				ui.RichTextEditor(text.Get()).InputValue(text).Frame(ui.Frame{Width: ui.L560}),
				ui.PrimaryButton(func() {
					text.Set(text.Get() + "<br>blub")
				}).Title("add blub"),
				ui.HLine(),
				ui.Text("only rich text view"),
				ui.RichText(text.Get()),
			).
				Frame(ui.Frame{}.MatchScreen())

		})
	}).
		Run()
}

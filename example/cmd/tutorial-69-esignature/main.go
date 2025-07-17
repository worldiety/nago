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
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfgsignature "go.wdy.de/nago/application/signature/cfg"
	uisignature "go.wdy.de/nago/application/signature/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/esignature"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_69")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))

		option.Must(cfgsignature.Enable(cfg))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("hello world"),

				// this is just a simple anemic view component
				esignature.Signature().
					TopText("Nago Signature").
					BottomText("Apprentice").
					Body(ui.Text("Torben")),

				// this is the ready-made signature infrastructure
				uisignature.UserSignature(wnd, wnd.Subject().ID(), user.Resource{Name: "nago.iam.user", ID: string(wnd.Subject().ID())}),
			).FullWidth()

		})
	}).Run()
}

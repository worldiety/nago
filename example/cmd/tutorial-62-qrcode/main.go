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
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_62")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("qr code demo"),
				ui.QrCode("https://github.com/worldiety/nago").AccessibilityLabel("github").Frame(ui.Frame{}.Size(ui.L320, ui.L320)),
			).Gap(ui.L32).Alignment(ui.Center).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

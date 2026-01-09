// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgchatbot "go.wdy.de/nago/application/chatbot/cfg"
	uichatbot "go.wdy.de/nago/application/chatbot/ui"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_82")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		messages := option.Must(cfgchatbot.Enable(cfg))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return uichatbot.PageSend(wnd, messages.UseCases)
		})
	}).Run()
}

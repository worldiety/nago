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
	cfgsms "go.wdy.de/nago/application/sms/cfg"
	uisms "go.wdy.de/nago/application/sms/ui"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_81")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		messages := option.Must(cfgsms.Enable(cfg))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return uisms.PageSend(wnd, messages.UseCases)
		})
	}).Run()
}

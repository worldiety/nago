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
	cfgai "go.wdy.de/nago/application/ai/cfg"
	_ "go.wdy.de/nago/application/ai/provider/mistralai"
	_ "go.wdy.de/nago/application/ai/provider/openai"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/drive"
	cfgdrive "go.wdy.de/nago/application/drive/cfg"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_77")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().
			NoFooterContentPadding(".").
			Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))
		drives := option.Must(cfgdrive.Enable(cfg))

		modAi := option.Must(cfgai.Enable(cfg))

		option.Must(drives.UseCases.OpenRoot(user.SU(), drive.OpenRootOptions{
			Create: true,
			Mode:   drive.OtherWrite | drive.OtherRead,
		}))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return uiai.PageChat(wnd, modAi.UseCases)
		})
	}).Run()
}

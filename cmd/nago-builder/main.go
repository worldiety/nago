// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/app/builder/aam/nagogen"
	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/app/builder/environment"
	uienv "go.wdy.de/nago/app/builder/environment/ui"
	"go.wdy.de/nago/application"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagobuilder")
		cfg.Serve(vuejs.Dist())

		appRepo := option.Must(application.JSONRepository[app.App, app.ID](cfg, "nbuilder.app"))
		envRepo := option.Must(application.JSONRepository[environment.Environment, environment.ID](cfg, "nbuilder.environment"))
		evtRepo := option.Must(application.JSONRepository[environment.EventBox, environment.EID](cfg, "nbuilder.event"))
		envUC := environment.NewUseCases(envRepo, appRepo, evtRepo)

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(
			cfg.NewScaffold().
				//	MenuEntry().Title(uienv.StrEnvironment.String()).Forward("drive").Private().
				Decorator(),
		)
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return uienv.PageEnvironments(wnd, envUC)
		})

		ucAam := aam.NewUseCases(envUC.Replay)
		ucGen := nagogen.NewUseCases()
		cfg.RootViewWithDecoration(uienv.PathApp, func(wnd core.Window) core.View {
			return uienv.PageApp(wnd, envUC, ucAam, ucGen)
		})

		cfg.RootViewWithDecoration(uienv.PathNamespace, func(wnd core.Window) core.View {
			return uienv.PageNamespace(wnd, envUC, ucAam)
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}

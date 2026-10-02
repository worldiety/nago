// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// This tutorial sets up the first user of an empty instance. Start it with an empty data directory and the mail
// address of the first user:
//
//	INITIAL_USER_EMAIL=me@example.com go run ./example/cmd/tutorial-114-onboarding
//
// Without an SMTP server, the code is printed to the log instead of being mailed. In production, leave out
// Deliver: the mail scheduler sends the code by an SMTP server of the system group or by the Nago Mail Service.
package main

import (
	"log/slog"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgonboarding "go.wdy.de/nago/application/onboarding/cfg"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

const adminRole role.ID = "de.worldiety.tutorial_114.admin"

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_114")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())

		var all []permission.ID
		for p := range permission.All() {
			all = append(all, p.ID)
		}

		option.MustZero(cfg.DeclareSystemRole(role.Role{ID: adminRole, Name: "Administrator"}, all...))

		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		std.Must(cfgonboarding.Enable(cfg, cfgonboarding.Options{
			Roles: []role.ID{adminRole},
			Deliver: func(to user.Email, code string, validUntil time.Time) error {
				slog.Warn("tutorial: the setup code", "to", to, "code", code, "validUntil", validUntil)
				return nil
			},
		}))

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("Hello " + wnd.Subject().Name()),
			).Frame(ui.Frame{}.MatchScreen())
		}))
	}).Run()
}

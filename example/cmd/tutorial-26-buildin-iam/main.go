// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// #[go.permission.generateTable]
package main

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	cfgusercircle "go.wdy.de/nago/application/usercircle/cfg"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

var myPermission = permission.Declare[SayHello]("de.worldiety.tutorial.say_hello", "Jeden Grüßen", "Diese Erlaubnis muss dem Nutzer zugewiesen werden.")

// SayHello greets everyone who has been authenticated.
type SayHello func(auth auth.Subject) string

func NewSayHello() SayHello {
	return func(auth auth.Subject) string {
		if err := auth.Audit(myPermission); err != nil {
			return fmt.Sprintf("invalid: %v", err)
		}

		return "hello " + auth.Name()
	}
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.SetName("Tutorial")

		std.Must(cfg.Authentication())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.MustZero(cfg.StandardSystems())
		option.Must(cfgusercircle.Enable(cfg))

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		sayHello := NewSayHello()

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text(fmt.Sprintf("%s", sayHello(wnd.Subject()))),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

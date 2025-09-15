// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"golang.org/x/text/language"
)

var (
	StrHelloStringKey = i18n.StringKey("string keys should normally not used")
	StrHelloWorld     = i18n.MustString("mydomain.example.hello_world", i18n.Values{language.English: "hello world", language.German: "Hallo Welt"})
	StrHelloX         = i18n.MustVarString(
		"mydomain.example.hello_x",
		i18n.Values{language.English: "hello {name}", language.German: "Hallo {name}"},
		i18n.LocalizationHint("It is best practice to give the translator some context in which this string is used"),
		i18n.LocalizationVarHint("name", "The subject name to greet, e.g. the firstname"),
	)
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_74")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text(StrHelloStringKey.Get(wnd)),
				ui.Text(StrHelloWorld.Get(wnd)),
				ui.Text(StrHelloX.Get(wnd, i18n.String("name", "Torben"))),
				ui.Text(rstring.ActionSave.Get(wnd)), // there also some predefined standard texts, which may want to use
			).FullWidth()

		})
	}).Run()
}

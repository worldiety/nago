// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"strconv"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/stepper"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_97")
		cfg.Serve(vuejs.Dist())

		num := 0
		cfg.RootView(".", func(wnd core.Window) core.View {
			num++
			return ui.VStack(
				ui.Text(strconv.Itoa(num)),
				stepper.Stepper(
					stepper.Step().Headline("first").SupportingText("the first thing"),
					stepper.Step().Headline("second").SupportingText("the second thing"),
					stepper.Step().Headline("third").SupportingText("the third thing"),
					stepper.Step().Headline("fourth").SupportingText("the fourth thing"),
					stepper.Step().Headline("fifth").SupportingText("the fifth thing"),
				),
			).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

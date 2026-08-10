// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accordion"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed accordion-content.gohtml
var accordionContent string

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_95")
		cfg.Serve(vuejs.Dist())

		count := 8

		cfg.RootView(".", func(wnd core.Window) core.View {
			accordions := make([]core.View, 0)
			for i := range count {
				accordions = append(accordions, accordion.Accordion(
					ui.Text(fmt.Sprintf("Accordion %d", i+1)),
					ui.RichText(fmt.Sprintf("Content %d: %s", i+1, accordionContent)),
					core.StateOf[bool](wnd, fmt.Sprintf("accordion_state_%d", i)),
				).FullWidth())
			}

			accordionsSmall := make([]core.View, 0)
			for i := range count {
				accordionsSmall = append(accordionsSmall, accordion.Accordion(
					ui.Text(fmt.Sprintf("Accordion Small %d", i+1)),
					ui.RichText(fmt.Sprintf("Content %d: %s", i+1, accordionContent)),
					core.StateOf[bool](wnd, fmt.Sprintf("accordion_small_state_%d", i)),
				).Small().FullWidth())
			}

			return ui.VStack(
				ui.VStack(
					accordions...,
				).Alignment(ui.Center).Padding(ui.Padding{}.All(ui.L32)).Frame(ui.Frame{MaxWidth: "800px"}.FullWidth()),
				ui.VStack(
					accordionsSmall...,
				).Alignment(ui.Center).Padding(ui.Padding{}.All(ui.L32)).Frame(ui.Frame{MaxWidth: "420px"}.FullWidth()),
			).Gap(ui.L32).Alignment(ui.Center).Frame(ui.Frame{}.FullWidth())
		})

	}).Run()
}

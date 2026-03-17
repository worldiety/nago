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
	. "go.wdy.de/nago/presentation/ui"
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
					Text(fmt.Sprintf("Accordion %d", i+1)),
					RichText(fmt.Sprintf("Content %d: %s", i+1, accordionContent)),
					core.StateOf[bool](wnd, fmt.Sprintf("accordion_state_%d", i)),
				).FullWidth())
			}

			return HStack(
				VStack(
					accordions...,
				).Alignment(Center).Padding(Padding{}.All(L32)).Frame(Frame{MaxWidth: "800px"}.FullWidth()),
			).Alignment(Center).Frame(Frame{}.FullWidth())
		})

	}).Run()
}

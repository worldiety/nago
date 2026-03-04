// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		buttons := make([]string, 0)
		for i := range 10000 {
			buttons = append(buttons, fmt.Sprintf("List entry %d", i+1))
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				list.List(
					ForEach(buttons, func(b string) core.View {
						return list.Entry().Headline(b)
					})...,
				),
			).FullWidth().Alignment(Center)
		})
	}).Run()
}

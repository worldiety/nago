// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// MyCustomColors is a ColorSet which provides a namespace and the type safe color fields.
type MyCustomColors struct {
	// Colors must be flat in this struct, public and of type color
	MySuperColor Color
}

func (m MyCustomColors) Default(scheme core.ColorScheme) core.ColorSet {
	if scheme == core.Light {
		return MyCustomColors{MySuperColor: "#cd29ff"}
	}

	return MyCustomColors{MySuperColor: "#12ffc8"}
}

func (m MyCustomColors) Namespace() core.NamespaceName {
	return "myCustomColor"
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.ColorSet(core.Light, MyCustomColors{
			MySuperColor: "#ff0000",
		})

		cfg.RootView(".", func(wnd core.Window) core.View {
			colors := core.Colors[MyCustomColors](wnd)
			oraColors := core.Colors[Colors](wnd)
			return VStack(
				Text("hello world").Color(oraColors.I0).BackgroundColor(oraColors.M0),
				FilledButton(colors.MySuperColor, nil).Title("my super button"),
			).Gap(L16).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}

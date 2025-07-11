// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed hummel.jpg
var hummelData application.StaticBytes

//go:embed gras.jpg
var grasData application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		hummelUri := cfg.Resource(hummelData)
		grasUri := cfg.Resource(grasData)

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				Image().
					URI(grasUri).
					Frame(Frame{}.Size("", L320)),

				CircleImage(hummelUri).
					AccessibilityLabel("Hummel an Lavendel").
					Padding(Padding{Top: L160.Negate()}),

				VStack(
					Text("Hummel").
						Font(Title),

					HStack(
						Text("WZO Terrasse"),
						Spacer(),
						Text("Oldenburg"),
					).Font(Font{Size: L12}).
						Frame(Frame{}.FullWidth()),

					HLine(),
					Text("Es gibt auch").Font(Title),
					Text("Andere Viecher"),
				).Alignment(Leading).
					Frame(Frame{Width: L320}),
			).
				Frame(Frame{Height: ViewportHeight, Width: Full})

		})

	}).Run()
}

func CircleImage(data core.URI) DecoredView {
	return Image().
		URI(data).
		Border(Border{}.
			Shadow(L8).
			Color("#ffffff").
			Width(L4).
			Circle()).
		Frame(Frame{}.Size(L320, L320))

}

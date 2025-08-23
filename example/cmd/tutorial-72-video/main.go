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
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/video"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed palm.mp4
var palm application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_72")
		cfg.Serve(vuejs.Dist())

		palmUri := cfg.Resource(palm)

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				video.Video(palmUri).
					AutoPlay(true).
					Loop(true).
					Controls(true),
			).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

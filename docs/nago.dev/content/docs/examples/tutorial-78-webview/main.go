// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	_ "go.wdy.de/nago/application/ai/provider/mistralai"
	_ "go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/webview"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_78")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, "+
					"no sea takimata sanctus est Lorem ipsum dolor sit amet.").
					TextAlignment(ui.TextAlignJustify).
					Hyphens(ui.HyphensAuto),

				webview.WebView().
					Src("https://www.youtube.com/embed/x6zAJ_CQnMo?si=vBOiOd-f2m9zkAT9").
					Allow("accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share").
					ReferrerPolicy("strict-origin-when-cross-origin").
					Title("YouTube video player").
					Frame(ui.Frame{Width: "560px", Height: "315px"}),
			).Frame(ui.Frame{}.Large())
		})
	}).Run()
}

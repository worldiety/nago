// Copyright (c) 2026 worldiety GmbH
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
	"go.wdy.de/nago/web/vuejs"
)

//go:embed signature.svg
var signature string

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_107")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			stateInitFunc := func() ui.Signature {
				return ui.Signature{
					SVG: signature,
				}
			}

			stateDefault := core.StateOf[ui.Signature](wnd, "stateDefault").Init(stateInitFunc)
			stateOptional := core.StateOf[ui.Signature](wnd, "stateOptional").Init(stateInitFunc)
			stateSupport := core.StateOf[ui.Signature](wnd, "stateSupport").Init(stateInitFunc)
			stateError := core.StateOf[ui.Signature](wnd, "stateError").Init(stateInitFunc)
			stateDisabled := core.StateOf[ui.Signature](wnd, "stateDisabled").Init(stateInitFunc)

			return ui.VStack(
				ui.ThemeSwitcher(
					ui.PrimaryButton(nil).Title("Toggle theme"),
				),
				ui.SignatureField("Default", stateDefault),
				ui.SignatureField("Default (optional)", stateOptional).Optional(true),
				ui.SignatureField("Mit Support", stateSupport).SupportingText("Ich bin ein Support-Text"),
				ui.SignatureField("Mit Fehler", stateError).ErrorText("Ich bin ein Fehler-Text"),
				ui.SignatureField("Disabled", stateDisabled).Disabled(true),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

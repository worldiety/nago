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
	icons2 "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/footer"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed ORA_logo.svg
var appIcon application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// update the global app icon
		cfg.AppIcon(cfg.Resource(appIcon))

		cfg.RootView(".", func(wnd core.Window) core.View {

			return Scaffold(ScaffoldAlignmentTop).
				Body(VStack(
					// this only causes the side effect of setting the current page title
					WindowTitle("Scaffold Example"),
					Text("Page body"),
				)).
				Footer(footer.Footer().
					Logo(Image().Embed(appIcon)).
					Impress("https://www.worldiety.de/impressum").
					PrivacyPolicy("https://www.worldiety.de/datenschutz").
					GeneralTermsAndConditions("https://www.worldiety.de/loesungen/software-entwicklung").
					ProviderName("© worldiety GmbH"),
				).
				Logo(Image().Embed(appIcon).Frame(Frame{}.Size("auto", L64))).
				Breakpoint(1000).
				Menu(
					ScaffoldMenuEntry{
						Title: "Without icon",
						Action: func() {
							fmt.Println("clicked 'Without icon'")
						},
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.User).Frame(Frame{}.Size(L20, L20)),
						Title:          "With icon",
						MarkAsActiveAt: ".",
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.ChatBubbleLeft).Frame(Frame{}.Size(L20, L20)),
						Title:          "With icon and long title",
						MarkAsActiveAt: ".",
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.QuestionMarkCircle).Frame(Frame{}.Size(L20, L20)),
						Title:          "With sub menu",
						MarkAsActiveAt: ".",
						Menu: []ScaffoldMenuEntry{
							{
								Title: "sub a",
							},
							{
								Title: "sub b",
							},
						},
					},
					ScaffoldMenuEntry{
						Icon:           Image().Embed(icons.Window).Frame(Frame{}.Size(L20, L20)),
						IconActive:     Image().Embed(icons2.Grid).Frame(Frame{}.Size(L20, L20)),
						Title:          "With active icon",
						MarkAsActiveAt: ".",
					},
					ForwardScaffoldMenuEntry(wnd, icons.User, "Forward entry", "/"),
				)
		})
	}).Run()
}

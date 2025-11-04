// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_65")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				HStack(
					VStack(
						Text("Display Large\nzweite Zeile\ndritte Zeile").Font(DisplayLarge),
						Text("Headline Large\nzweite Zeile\ndritte Zeile").Font(HeadlineLarge),
						Text("Title Large\nzweite Zeile\ndritte Zeile").Font(TitleLarge),
						Text("Body Large\nzweite Zeile\ndritte Zeile").Font(BodyLarge),
						Text("Label Large\nzweite Zeile\ndritte Zeile").Font(LabelLarge),
					).Gap(L64).Alignment(Leading),
					VStack(
						Text("Display Medium\nzweite Zeile\ndritte Zeile").Font(DisplayMedium),
						Text("Headline Medium\nzweite Zeile\ndritte Zeile").Font(HeadlineMedium),
						Text("Title Medium\nzweite Zeile\ndritte Zeile").Font(TitleMedium),
						Text("Body Medium\nzweite Zeile\ndritte Zeile").Font(BodyMedium),
						Text("Label Medium\nzweite Zeile\ndritte Zeile").Font(LabelMedium),
					).Gap(L64).Alignment(Leading),
					VStack(
						Text("Display Small\nzweite Zeile\ndritte Zeile").Font(DisplaySmall),
						Text("Headline Small\nzweite Zeile\ndritte Zeile").Font(HeadlineSmall),
						Text("Title Small\nzweite Zeile\ndritte Zeile").Font(TitleSmall),
						Text("Body Small\nzweite Zeile\ndritte Zeile").Font(BodySmall),
						Text("Label Small\nzweite Zeile\ndritte Zeile").Font(LabelSmall),
					).Gap(L64).Alignment(Leading),
				).Gap(L64).Alignment(TopLeading).Padding(Padding{}.All(L64)),
				HStack(
					VStack(
						Text("Mono Large\nzweite Zeile\ndritte Zeile").Font(MonoLarge),
						Text("Mono Bold Large\nzweite Zeile\ndritte Zeile").Font(MonoBoldLarge),
						Text("Mono Italic Large\nzweite Zeile\ndritte Zeile").Font(MonoItalicLarge),
					).Gap(L64).Alignment(Leading),
					VStack(
						Text("Mono Medium\nzweite Zeile\ndritte Zeile").Font(MonoMedium),
						Text("Mono Bold Medium\nzweite Zeile\ndritte Zeile").Font(MonoBoldMedium),
						Text("Mono Italic Medium\nzweite Zeile\ndritte Zeile").Font(MonoItalicMedium),
					).Gap(L64).Alignment(Leading),
					VStack(
						Text("Mono Small\nzweite Zeile\ndritte Zeile").Font(MonoSmall),
						Text("Mono Bold Small\nzweite Zeile\ndritte Zeile").Font(MonoBoldSmall),
						Text("Mono Italic Small\nzweite Zeile\ndritte Zeile").Font(MonoItalicSmall),
					).Gap(L64).Alignment(Leading),
				).Gap(L64).Alignment(TopLeading).Padding(Padding{}.All(L64)),
			).Alignment(Leading)
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}

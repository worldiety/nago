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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {

			return HStack(
				defaultButtons(),
				VLine(),
				customButtons(wnd),
			).Gap(L16).Alignment(Top).Frame(Frame{}.FullWidth())

		})
	}).Run()
}

func defaultButtons() core.View {
	return VStack(
		Text("Standard Buttons").Font(Title),

		PrimaryButton(func() {
			fmt.Println("clicked the real primary")
		}).Title("primary button"),

		PrimaryButton(nil).
			Title("primary with pre icon").
			PreIcon(icons.SpeakerWave),

		PrimaryButton(nil).
			Title("primary with post icon").
			PostIcon(icons.SpeakerWave),

		PrimaryButton(nil).
			PreIcon(icons.SpeakerWave),

		SecondaryButton(nil).Title("secondary button"),
		TertiaryButton(nil).Title("tertiary button"),
		PasswordField("Test", "Test"),
		PasswordField("Test", "Test"),
		Toggle(false),

		PrimaryButton(nil).
			Title("link button").
			HRef("https://www.worldiety.de").
			Target("_self"),
	).Gap(L16).
		Padding(Padding{}.All(L16))
}

func customButtons(wnd core.Window) core.View {
	colors := core.Colors[Colors](wnd) // grab our active ColorSet
	return VStack(
		Text("Custom Buttons").Font(Title),

		// we cannot use the variable "I0" because the function needs to calculate active and hover derivates itself
		FilledButton(colors.I0, func() {
			fmt.Println("clicked a custom filled button")
		}).Title("fake primary button"),

		FilledButton(colors.I0, nil).
			PreIcon(icons.SpeakerWave).
			Title("fake primary button"),

		FilledButton(colors.I0, nil).
			PostIcon(icons.SpeakerWave).
			Title("fake primary button"),

		FilledButton(colors.I0, nil).
			PreIcon(icons.SpeakerWave),

		// hardcoded button, which does not react to color changes
		FilledButton("#EF8A97", nil).TextColor("#ffffff").Title("arbitrary color"),
	).Gap(L16).
		Padding(Padding{}.All(L16)) // graphical glitch: use some padding, custom buttons are 2dp larger due to emulated focus border, otherwise it gets clipped
}

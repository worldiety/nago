// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"math/rand"
	"strconv"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_110")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(true).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			isPresented := core.AutoState[bool](wnd)

			stateCount := core.StateOf[int](wnd, "stateCount")
			msgPart := "I am a notification banner alert thingy."

			msg := msgPart
			if rand.Intn(100) < 50 {
				msg = msg + " " + msgPart
			}
			if rand.Intn(100) < 50 {
				msg = msg + " " + msgPart
			}
			if rand.Intn(100) < 50 {
				msg = msg + " " + msgPart
			}

			showBanner := func(intent alert.Intent) {
				stateCount.Set(stateCount.Get() + 1)
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   "Beep boop " + strconv.Itoa(stateCount.Get()),
					Message: msg,
					Intent:  intent,
				})
			}

			return ui.VStack(
				ui.ThemeSwitcher(
					ui.PrimaryButton(nil).Title("Toggle theme"),
				),

				ui.VStack(
					ui.SecondaryButton(func() {
						isPresented.Set(true)
					}).Title("Show Dialog"),

					ui.If(isPresented.Get(), ui.Modal(
						ui.Dialog(
							ui.VStack(
								ui.PrimaryButton(func() { showBanner(alert.IntentOk) }).Title("Show alert OK"),
								ui.PrimaryButton(func() { showBanner(alert.IntentSuccess) }).Title("Show alert Success"),
								ui.PrimaryButton(func() { showBanner(alert.IntentWarning) }).Title("Show alert Warning"),
								ui.PrimaryButton(func() { showBanner(alert.IntentError) }).Title("Show alert Error"),
							).Gap(ui.L16),
						),
					).OnDismissRequest(func() {
						isPresented.Set(false)
					})),
				),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		}))
	}).Run()
}

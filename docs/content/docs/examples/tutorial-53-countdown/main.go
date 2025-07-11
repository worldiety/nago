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
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_53")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Login(false).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			done := core.AutoState[bool](wnd)
			nr := core.AutoState[int](wnd)
			fmt.Println("render was called")

			return ui.VStack(
				ui.Text("hello world!"),
				ui.PrimaryButton(func() {
					nr.Set(nr.Get() + 1)
					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   "ok message",
						Message: fmt.Sprintf("Message no %d", nr.Get()),
						Intent:  alert.IntentOk,
					})
				}).Title("Spawn"),
				ui.If(done.Get(), ui.Text("timer is done")),
				ui.CountDown(time.Second*10).Action(func() {
					done.Set(true)
				}).Style(ui.CountDownStyleClock).
					Done(done.Get()).
					Frame(ui.Frame{}.FullWidth()),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())

		}))

	}).Run()
}

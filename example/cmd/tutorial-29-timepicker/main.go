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
	"go.wdy.de/nago/presentation/ui/timepicker"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			duration := core.AutoState[time.Duration](wnd).Init(func() time.Duration {
				return time.Minute * 61
			})
			return VStack(
				timepicker.Picker("Dauer", duration).
					SupportingText("Wähle eine tolle Zeit").
					Format(timepicker.DecomposedFormat).
					Days(true).
					Hours(true).
					Minutes(true).
					Seconds(true),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}

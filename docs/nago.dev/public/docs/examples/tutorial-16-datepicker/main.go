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
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			date := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 1, Month: 6, Year: 2024}
			})

			start := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 2, Month: 7, Year: 2024}
			})

			end := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 20, Month: 7, Year: 2024}
			})

			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", Text(fmt.Sprintf("Deine Eingabe: %v, start=%v end=%v", date, start, end)), showAlert, alert.Ok()),
				SingleDatePicker("Single date", date.Get(), date),
				SingleDatePicker("Singel date (double month view)", date.Get(), date).DoubleMode(true),
				RangeDatePicker("Date range", start.Get(), start, end.Get(), end),
				RangeDatePicker("Date range (double month view)", start.Get(), start, end.Get(), end).DoubleMode(true),

				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}

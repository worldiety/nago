// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
)

func init() {
	app.Register("time-picker", func(wnd core.Window) core.View {
		alarm := core.AutoState[time.Duration](wnd).Init(func() time.Duration {
			return 6*time.Hour + 45*time.Minute
		})
		timeout := core.AutoState[time.Duration](wnd).Init(func() time.Duration {
			return 1*time.Hour + 30*time.Minute
		})

		return VStack(
			timepicker.Picker("Alarm", alarm).
				Format(timepicker.ClockFormat).
				Frame(Frame{}.FullWidth()),
			timepicker.Picker("Session timeout", timeout).
				Format(timepicker.DecomposedFormat).
				Days(false).Seconds(false).
				SupportingText("Hours and minutes").
				Frame(Frame{}.FullWidth()),
		).Gap(L16).Frame(Frame{Width: L320})
	})
}

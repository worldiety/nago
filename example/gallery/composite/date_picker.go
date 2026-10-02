// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("date-picker", func(wnd core.Window) core.View {
		birthday := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
			return xtime.Date{Day: 14, Month: 3, Year: 1990}
		})
		start := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
			return xtime.Date{Day: 3, Month: 8, Year: 2026}
		})
		end := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
			return xtime.Date{Day: 14, Month: 8, Year: 2026}
		})

		return VStack(
			SingleDatePicker("Birthday", birthday.Get(), birthday),
			RangeDatePicker("Vacation", start.Get(), start, end.Get(), end).
				SupportingText("Pick the first and the last day"),
		).Alignment(Leading).Gap(L16).Frame(Frame{Width: L320})
	})
}

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
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timeframe"
)

func init() {
	app.Register("time-frame-picker", func(wnd core.Window) core.View {
		meeting := core.AutoState[xtime.TimeFrame](wnd).Init(func() xtime.TimeFrame {
			start := time.Date(2026, 3, 12, 9, 30, 0, 0, time.UTC)
			return xtime.TimeFrame{
				StartTime: xtime.UnixMilliseconds(start.UnixMilli()),
				EndTime:   xtime.UnixMilliseconds(start.Add(90 * time.Minute).UnixMilli()),
				Timezone:  "UTC",
			}
		})

		return timeframe.Picker("Meeting", meeting).
			Title("Schedule the meeting").
			SupportingText("Date, start and end time").
			Frame(Frame{Width: L400})
	})
}

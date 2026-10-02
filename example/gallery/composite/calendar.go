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
	"go.wdy.de/nago/presentation/ui/calendar"
)

func init() {
	app.Register("calendar", func(wnd core.Window) core.View {
		day := func(month time.Month, d int) calendar.Instant {
			return calendar.Instant{At: time.Date(2026, month, d, 0, 0, 0, 0, time.UTC)}
		}

		vacation := calendar.Category{Label: "Vacation", Color: "#2BCA73"}
		training := calendar.Category{Label: "Training", Color: "#1B8C98"}

		// Year uses German column labels, so set English ones
		year := calendar.Year(2026)
		year.Columns = nil
		for m := time.January; m <= time.December; m++ {
			year.Columns = append(year.Columns, calendar.Column{Label: m.String()[:3]})
		}

		// the timeline styles ignore Frame, so the parent defines the width
		return VStack(calendar.Calendar(
			calendar.Event{From: day(2, 2), To: day(3, 6), Label: "Ski trip", Lane: calendar.Lane{Label: "Anna"}, Categories: []calendar.Category{vacation}},
			calendar.Event{From: day(5, 4), To: day(6, 12), Label: "Go workshop", Lane: calendar.Lane{Label: "Anna"}, Categories: []calendar.Category{training}},
			calendar.Event{From: day(7, 6), To: day(8, 28), Label: "Summer break", Lane: calendar.Lane{Label: "Ben"}, Categories: []calendar.Category{vacation}},
			calendar.Event{From: day(9, 21), To: day(10, 30), Label: "UX bootcamp", Lane: calendar.Lane{Label: "Ben"}, Categories: []calendar.Category{training}},
			calendar.Event{From: day(11, 23), To: day(12, 31), Label: "Holidays", Lane: calendar.Lane{Label: "Clara"}, Categories: []calendar.Category{vacation}},
		).ViewPort(year).FullWidth()).Frame(Frame{Width: L1200})
	})
}

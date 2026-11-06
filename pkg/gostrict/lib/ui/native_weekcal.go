// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	nui "go.wdy.de/nago/presentation/ui"
	nweekcal "go.wdy.de/nago/presentation/ui/weekcal"
)

func weekCalendarDays(days []WeekCalendarDay) []nweekcal.Day {
	res := make([]nweekcal.Day, 0, len(days))
	for _, d := range days {
		res = append(res, nweekcal.Day{Title: d.Title, Subtitle: d.Subtitle, Today: d.Today, Muted: d.Muted})
	}
	return res
}

func weekCalendarEvents(events []WeekCalendarEvent) []nweekcal.Event {
	res := make([]nweekcal.Event, 0, len(events))
	for _, e := range events {
		res = append(res, nweekcal.Event{
			ID: e.ID, Day: int(e.Day), From: int(e.From), To: int(e.To), Title: e.Title, Text: e.Text,
			Color: nui.Color(e.Color), Fixed: e.Fixed,
		})
	}
	return res
}

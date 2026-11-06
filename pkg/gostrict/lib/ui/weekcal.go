// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nweekcal "go.wdy.de/nago/presentation/ui/weekcal"
)

// WeekCalendarDay is the head of a column of a week calendar, like weekcal.Day of nago.
type WeekCalendarDay struct {
	Title string
	// Subtitle is shown below the title, with up to two lines, e.g. a holiday or an absence.
	Subtitle string
	// Today marks the current day.
	Today bool
	// Muted marks a day, which is not a working day, e.g. a weekend or a holiday.
	Muted bool
}

// WeekCalendarEvent is a block of a week calendar, like weekcal.Event of nago. Day is the index of its day, From and
// To are minutes since midnight of the day, 0 <= From < To <= 1440. The ID is not empty and unique within the
// calendar. A fixed event can only be clicked.
type WeekCalendarEvent struct {
	ID       string
	Day      int32
	From, To int32
	Title    string
	Text     string
	Color    Color
	Fixed    bool
}

// WeekCalendarSpan is an interval, which the user spanned in the free time of a week calendar, like weekcal.Span
// of nago.
type WeekCalendarSpan struct {
	Day, From, To int32
}

// WeekCalendarMove is a move of an event of a week calendar or a change of its duration, like weekcal.Move of nago.
// OldDay, OldFrom and OldTo are the interval, which the user saw, when the gesture began: if the event has been
// changed elsewhere in the meantime, it differs from the stored one.
type WeekCalendarMove struct {
	ID                     string
	Day, From, To          int32
	OldDay, OldFrom, OldTo int32
}

// TWeekCalendar is a week calendar, see WeekCalendar.
type TWeekCalendar struct{ ref js.Ref }

func (TWeekCalendar) isView() {}

// WeekCalendar returns a calendar with 1 to 7 days as columns over a time axis, like weekcal.WeekCalendar of nago.
// The id must be unique within the window. By default, it shows 06:00 to 22:00 in a raster of 15 minutes.
//
//gostrict:host
func WeekCalendar(id string, days ...WeekCalendarDay) TWeekCalendar {
	return TWeekCalendar{ref: wrap(nweekcal.WeekCalendar(id, weekCalendarDays(days)...))}
}

// Events sets the events of the calendar.
//
//gostrict:host
func (c TWeekCalendar) Events(events ...WeekCalendarEvent) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Events(weekCalendarEvents(events)...))}
}

// Range sets the visible minutes, 0 <= from < to <= 1440 in the raster of Step, by default 360 to 1320.
//
//gostrict:host
func (c TWeekCalendar) Range(from, to int32) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Range(int(from), int(to)))}
}

// Step sets the raster in minutes: 5, 10, 15, 20, 30 or 60, by default 15.
//
//gostrict:host
func (c TWeekCalendar) Step(minutes int32) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Step(int(minutes)))}
}

// ClickSpan sets the minutes, which a click into free time spans, a multiple of Step, by default 60.
//
//gostrict:host
func (c TWeekCalendar) ClickSpan(minutes int32) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).ClickSpan(int(minutes)))}
}

// Now shows the line of the current time at the day and the minute.
//
//gostrict:host
func (c TWeekCalendar) Now(day, minute int32) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Now(int(day), int(minute)))}
}

// HourHeight sets the height of an hour in dp, 24 to 240, by default 48.
//
//gostrict:host
func (c TWeekCalendar) HourHeight(dp float64) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).HourHeight(dp))}
}

// Disabled shows the calendar without any action.
//
//gostrict:host
func (c TWeekCalendar) Disabled(disabled bool) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Disabled(disabled))}
}

// Frame sets the size, by default the full width and the height of the range.
//
//gostrict:host
func (c TWeekCalendar) Frame(frame Frame) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).Frame(frame.nago()))}
}

// OnClick calls fn with the id of an event, which the user clicked.
//
//gostrict:host
//gostrict:calls later
func (c TWeekCalendar) OnClick(fn func(id string)) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).OnClick(fn))}
}

// OnSelect calls fn with an interval, which the user spanned in free time. An error is shown in the calendar.
//
//gostrict:host
//gostrict:calls later
func (c TWeekCalendar) OnSelect(fn func(s WeekCalendarSpan) error) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).OnSelect(func(s nweekcal.Span) error {
		return fn(WeekCalendarSpan{Day: int32(s.Day), From: int32(s.From), To: int32(s.To)})
	}))}
}

// OnMove calls fn with a move of an event, also to another day, or a change of its duration. An error is shown in
// the calendar.
//
//gostrict:host
//gostrict:calls later
func (c TWeekCalendar) OnMove(fn func(m WeekCalendarMove) error) TWeekCalendar {
	return TWeekCalendar{ref: wrap(get[nweekcal.TWeekCalendar](c.ref).OnMove(func(m nweekcal.Move) error {
		return fn(WeekCalendarMove{
			ID: m.ID, Day: int32(m.Day), From: int32(m.From), To: int32(m.To),
			OldDay: int32(m.OldDay), OldFrom: int32(m.OldFrom), OldTo: int32(m.OldTo),
		})
	}))}
}

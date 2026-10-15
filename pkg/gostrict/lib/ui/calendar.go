// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"time"

	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/calendar"
)

// CalendarStyle is the display style of a calendar, like calendar.Style of nago. The calendar itself is TCalendar.
type CalendarStyle int32

const (
	// CalendarTimelineYear shows the year as a timeline, like calendar.TimelineYear of nago.
	CalendarTimelineYear CalendarStyle = 0
	// CalendarTimelineDay shows the day as a timeline, like calendar.TimelineDay of nago.
	CalendarTimelineDay CalendarStyle = 1
	// CalendarStartTimeSequence shows the events as a sequence of their start times, like calendar.StartTimeSequence of nago.
	CalendarStartTimeSequence CalendarStyle = 2
)

// CalendarSeqPillTimeHint tells, how the start time of an event in a start time sequence is shown, like
// calendar.SeqPillTimeHint of nago.
type CalendarSeqPillTimeHint int32

const (
	// CalendarPillTimeYYYYMMDD shows the date, like calendar.PillTimeYYYYMMDD of nago.
	CalendarPillTimeYYYYMMDD CalendarSeqPillTimeHint = 0
	// CalendarPillTimeHHMM shows the clock time, like calendar.PillTimeHHMM of nago.
	CalendarPillTimeHHMM CalendarSeqPillTimeHint = 1
	// CalendarPillTimeNone shows neither, like calendar.PillTimeNone of nago.
	CalendarPillTimeNone CalendarSeqPillTimeHint = 2
)

// CalendarColors defines the colors of a calendar, like calendar.Colors of nago.
type CalendarColors struct {
	Header               Color
	LaneBackground       Color
	PrePostBackground    Color
	PrePostForeground    Color
	EventBackground      Color
	EventHoverBackground Color
	Text                 Color
	EventText            Color
	Separator            Color
}

// CalendarCategory classifies an event by a label and a color, like calendar.Category of nago.
type CalendarCategory struct {
	Color Color
	Label string
}

// CalendarLane is a track or row of a calendar, which groups events, like calendar.Lane of nago.
type CalendarLane struct {
	Label string
}

// CalendarColumn is a labeled column of a calendar timeline, like calendar.Column of nago.
type CalendarColumn struct {
	Label string
}

// CalendarPillar is a high-level theme, which groups related events, like calendar.Pillar of nago.
type CalendarPillar struct {
	Label  string
	Icon   SVG
	Width  Length
	Height Length
}

// CalendarChip shows additional information of an event, like calendar.Chip of nago.
type CalendarChip struct {
	Label       string
	Icon        SVG
	FillColor   Color
	StrokeColor Color
	BgColor     Color
	TextColor   Color
	Alignment   Alignment
	FullWidth   bool
}

// CalendarOffset is a time span relative to an instant of an event, like calendar.Offset of nago. The duration is in
// milliseconds.
type CalendarOffset struct {
	Duration int32
	Icon     SVG
	Label    string
}

// CalendarPercent is a relative position between 0 and 100 within a calendar, like calendar.Percent of nago.
type CalendarPercent float64

// String rounds and formats the percent to at most 2 decimal places, like calendar.Percent.String of nago.
//
//gostrict:host
func (p CalendarPercent) String() string {
	return calendar.Percent(p).String()
}

// Length converts the percent into a length, like calendar.Percent.Length of nago.
//
//gostrict:host
func (p CalendarPercent) Length() Length {
	return Length(calendar.Percent(p).Length())
}

// CalendarDefaultColors returns the standard colors of a calendar, like calendar.DefaultColors of nago.
//
//gostrict:host
func CalendarDefaultColors() CalendarColors {
	return calendarColorsFrom(calendar.DefaultColors())
}

// TCalendarEvent is an event of a calendar, like calendar.Event of nago. The builder methods set its fields.
type TCalendarEvent struct{ ref js.Ref }

// TCalendarViewPort is the time span and the columns, which a calendar shows, like calendar.ViewPort of nago.
type TCalendarViewPort struct{ ref js.Ref }

// TCalendar shows events in a timeline or as a sequence of start times, like calendar.TCalendar of nago.
type TCalendar struct{ ref js.Ref }

func (TCalendar) isView() {}

// CalendarEvent returns an event with the label, which starts at from and ends at to, both inclusive, like
// calendar.Event of nago.
//
//gostrict:host
func CalendarEvent(label string, from, to Time) TCalendarEvent {
	return TCalendarEvent{ref: wrap(calendar.Event{Label: label, From: calendar.Instant{At: from.nago()}, To: calendar.Instant{At: to.nago()}})}
}

// Action sets the function, which is called, when the user clicks the event.
//
//gostrict:host
func (e TCalendarEvent) Action(fn func()) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Action = fn })
}

// FromOffset sets the offset of the start, e.g. the travel time by bus.
//
//gostrict:host
func (e TCalendarEvent) FromOffset(offset CalendarOffset) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.From.Offset = offset.nago() })
}

// ToOffset sets the offset of the end, e.g. the travel time by train.
//
//gostrict:host
func (e TCalendarEvent) ToOffset(offset CalendarOffset) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.To.Offset = offset.nago() })
}

// Categories sets the categories, each of which is shown as its own color bar.
//
//gostrict:host
func (e TCalendarEvent) Categories(categories ...CalendarCategory) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Categories = calendarCategories(categories) })
}

// Pillar sets the pillar, which groups the event.
//
//gostrict:host
func (e TCalendarEvent) Pillar(pillar CalendarPillar) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Pillar = pillar.nago() })
}

// Lane sets the lane, which groups the event.
//
//gostrict:host
func (e TCalendarEvent) Lane(lane CalendarLane) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Lane = lane.nago() })
}

// Chips sets the chips, which show additional information.
//
//gostrict:host
func (e TCalendarEvent) Chips(chips ...CalendarChip) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Chips = calendarChips(chips) })
}

// Organiser sets the name of the organiser.
//
//gostrict:host
func (e TCalendarEvent) Organiser(organiser string) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Organiser = organiser })
}

// Location sets the location.
//
//gostrict:host
func (e TCalendarEvent) Location(location string) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.Location = location })
}

// Cancelled marks the event as cancelled by its host, like calendar.Event.IsCancelled of nago.
//
//gostrict:host
func (e TCalendarEvent) Cancelled(cancelled bool) TCalendarEvent {
	return e.with(func(v *calendar.Event) { v.IsCancelled = cancelled })
}

// Render sets the function, which returns the view of the event instead of the default one.
//
//gostrict:host
func (e TCalendarEvent) Render(fn func(style CalendarStyle) View) TCalendarEvent {
	return e.with(func(v *calendar.Event) {
		v.Render = func(s calendar.Style) core.View { return viewOf(fn(CalendarStyle(s))) }
	})
}

// CalendarYear returns the view port of a year with a column for each month, like calendar.Year of nago.
//
//gostrict:host
func CalendarYear(year int32) TCalendarViewPort {
	return TCalendarViewPort{ref: wrap(calendar.Year(int(year)))}
}

// CalendarDay returns the view port of a day, like calendar.Day of nago. The month starts at 1.
//
//gostrict:host
func CalendarDay(year, month, day int32) TCalendarViewPort {
	return TCalendarViewPort{ref: wrap(calendar.Day(int(year), time.Month(month), int(day)))}
}

// From sets the start of the view port.
//
//gostrict:host
func (p TCalendarViewPort) From(t Time) TCalendarViewPort {
	return p.with(func(v *calendar.ViewPort) { v.From = t.nago() })
}

// To sets the end of the view port.
//
//gostrict:host
func (p TCalendarViewPort) To(t Time) TCalendarViewPort {
	return p.with(func(v *calendar.ViewPort) { v.To = t.nago() })
}

// Columns sets the labeled columns of the timeline.
//
//gostrict:host
func (p TCalendarViewPort) Columns(columns ...CalendarColumn) TCalendarViewPort {
	return p.with(func(v *calendar.ViewPort) { v.Columns = calendarColumns(columns) })
}

// LaneWidth sets the width of the label of a lane in percent.
//
//gostrict:host
func (p TCalendarViewPort) LaneWidth(width CalendarPercent) TCalendarViewPort {
	return p.with(func(v *calendar.ViewPort) { v.LaneWidth = calendar.Percent(width) })
}

// TimeStyle sets, how the start time of an event is shown in a start time sequence.
//
//gostrict:host
func (p TCalendarViewPort) TimeStyle(style CalendarSeqPillTimeHint) TCalendarViewPort {
	return p.with(func(v *calendar.ViewPort) { v.TimeStyle = calendar.SeqPillTimeHint(style) })
}

// Percent returns the relative position of the time within the view port, like calendar.ViewPort.Percent of nago.
//
//gostrict:host
func (p TCalendarViewPort) Percent(t Time) CalendarPercent {
	return CalendarPercent(get[calendar.ViewPort](p.ref).Percent(t.nago()))
}

// Calendar returns a calendar of the events, which shows the current year as a timeline at first, like
// calendar.Calendar of nago.
//
//gostrict:host
func Calendar(events ...TCalendarEvent) TCalendar {
	return TCalendar{ref: wrap(calendar.Calendar(calendarEvents(events)...))}
}

// Style sets the style.
//
//gostrict:host
func (c TCalendar) Style(style CalendarStyle) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).Style(calendar.Style(style)))}
}

// Append adds events.
//
//gostrict:host
func (c TCalendar) Append(events ...TCalendarEvent) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).Append(calendarEvents(events)...))}
}

// Frame sets the size.
//
//gostrict:host
func (c TCalendar) Frame(frame Frame) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).Frame(frame.nago()))}
}

// ViewPort sets the view port.
//
//gostrict:host
func (c TCalendar) ViewPort(vp TCalendarViewPort) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).ViewPort(get[calendar.ViewPort](vp.ref)))}
}

// Colors sets the colors.
//
//gostrict:host
func (c TCalendar) Colors(colors CalendarColors) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).Colors(colors.nago()))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TCalendar) FullWidth() TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).FullWidth())}
}

// MaxCategories limits the number of the color bars of an event.
//
//gostrict:host
func (c TCalendar) MaxCategories(n int32) TCalendar {
	return TCalendar{ref: wrap(get[calendar.TCalendar](c.ref).MaxCategories(int(n)))}
}

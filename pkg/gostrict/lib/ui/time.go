// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"time"
)

// Time is a point in time as milliseconds since the Unix epoch, like xtime.UnixMilliseconds of nago. A float64 holds
// it exactly up to 2^53. The calendar fields use the local time zone of the server.
type Time float64

// Duration is a time span in milliseconds.
type Duration float64

const (
	Millisecond Duration = 1
	Second      Duration = 1000
	Minute      Duration = 60000
	Hour        Duration = 3600000
)

// Date is a day in a month of a year without a time zone, like xtime.Date of nago. The month starts at 1.
type Date struct {
	Day   int32
	Month int32
	Year  int32
}

// TimeFrame is an interval between two points in time, like xtime.TimeFrame of nago. The time zone is an
// identifier like Europe/Berlin or empty for the local one.
type TimeFrame struct {
	StartTime Time
	EndTime   Time
	Timezone  string
}

// TimeOf returns the time of the calendar fields in the local time zone of the server, like time.Date.
//
//gostrict:host
func TimeOf(year, month, day, hour, minute, second int32) Time {
	return timeFrom(time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.Local))
}

// Now returns the current time.
//
//gostrict:host
func Now() Time {
	return Time(xtime.Now())
}

// Add returns the time moved by the duration.
//
//gostrict:host
func (t Time) Add(d Duration) Time {
	return Time(float64(t) + float64(d))
}

// Date returns the day of the time in the local time zone.
//
//gostrict:host
func (t Time) Date() Date {
	return dateFrom(xtime.UnixMilliseconds(t).Date(time.Local))
}

// Year returns the year in the local time zone.
//
//gostrict:host
func (t Time) Year() int32 {
	return int32(t.nago().Year())
}

// Month returns the month in the local time zone, starting at 1.
//
//gostrict:host
func (t Time) Month() int32 {
	return int32(t.nago().Month())
}

// Day returns the day of the month in the local time zone, starting at 1.
//
//gostrict:host
func (t Time) Day() int32 {
	return int32(t.nago().Day())
}

// Hour returns the hour in the local time zone.
//
//gostrict:host
func (t Time) Hour() int32 {
	return int32(t.nago().Hour())
}

// Minute returns the minute in the local time zone.
//
//gostrict:host
func (t Time) Minute() int32 {
	return int32(t.nago().Minute())
}

// Format returns the time in the local time zone as text, which follows the layout of Go, e.g. 2006-01-02 15:04.
//
//gostrict:host
func (t Time) Format(layout string) string {
	return t.nago().Format(layout)
}

// Time returns the start of the day in the local time zone, like xtime.Date.Time of nago.
//
//gostrict:host
func (d Date) Time() Time {
	return timeFrom(d.nago().Time(time.Local))
}

// String returns the date as year.month.day, like xtime.Date.String of nago.
//
//gostrict:host
func (d Date) String() string {
	return d.nago().String()
}

// IsZero reports, whether no day is set.
//
//gostrict:host
func (d Date) IsZero() bool {
	return d.nago().IsZero()
}

// DateState is a day of the window, which it keeps between two renders, e.g. of a date picker.
type DateState struct{ ref js.Ref }

// TimeFrameState is a time frame of the window, e.g. of a timeframe picker.
type TimeFrameState struct{ ref js.Ref }

// DurationState is a duration of the window, e.g. of a time picker.
type DurationState struct{ ref js.Ref }

// DateStateOf returns the state with the id of the window, which has no day at first.
//
//gostrict:host
func DateStateOf(wnd Window, id string) DateState {
	return DateState{ref: wrap(core.StateOf[xtime.Date](wnd.nago(), id))}
}

// Get returns the day.
//
//gostrict:host
func (s DateState) Get() Date {
	return dateFrom(s.nago().Get())
}

// Set sets the day and renders the window again.
//
//gostrict:host
func (s DateState) Set(v Date) {
	s.nago().Set(v.nago())
}

// Init sets the day of fn, if the state is new.
//
//gostrict:host
func (s DateState) Init(fn func() Date) DateState {
	return DateState{ref: wrap(s.nago().Init(func() xtime.Date { return fn().nago() }))}
}

// Observe calls fn, when the user changes the day.
//
//gostrict:host
func (s DateState) Observe(fn func(v Date)) DateState {
	return DateState{ref: wrap(s.nago().Observe(func(v xtime.Date) { fn(dateFrom(v)) }))}
}

// TimeFrameStateOf returns the state with the id of the window, which has no time frame at first.
//
//gostrict:host
func TimeFrameStateOf(wnd Window, id string) TimeFrameState {
	return TimeFrameState{ref: wrap(core.StateOf[xtime.TimeFrame](wnd.nago(), id))}
}

// Get returns the time frame.
//
//gostrict:host
func (s TimeFrameState) Get() TimeFrame {
	return timeFrameFrom(s.nago().Get())
}

// Set sets the time frame and renders the window again.
//
//gostrict:host
func (s TimeFrameState) Set(v TimeFrame) {
	s.nago().Set(v.nago())
}

// Init sets the time frame of fn, if the state is new.
//
//gostrict:host
func (s TimeFrameState) Init(fn func() TimeFrame) TimeFrameState {
	return TimeFrameState{ref: wrap(s.nago().Init(func() xtime.TimeFrame { return fn().nago() }))}
}

// Observe calls fn, when the user changes the time frame.
//
//gostrict:host
func (s TimeFrameState) Observe(fn func(v TimeFrame)) TimeFrameState {
	return TimeFrameState{ref: wrap(s.nago().Observe(func(v xtime.TimeFrame) { fn(timeFrameFrom(v)) }))}
}

// DurationStateOf returns the state with the id of the window, which is 0 at first.
//
//gostrict:host
func DurationStateOf(wnd Window, id string) DurationState {
	return DurationState{ref: wrap(core.StateOf[time.Duration](wnd.nago(), id))}
}

// Get returns the duration.
//
//gostrict:host
func (s DurationState) Get() Duration {
	return durationFrom(s.nago().Get())
}

// Set sets the duration and renders the window again.
//
//gostrict:host
func (s DurationState) Set(v Duration) {
	s.nago().Set(v.nago())
}

// Init sets the duration of fn, if the state is new.
//
//gostrict:host
func (s DurationState) Init(fn func() Duration) DurationState {
	return DurationState{ref: wrap(s.nago().Init(func() time.Duration { return fn().nago() }))}
}

// Observe calls fn, when the user changes the duration.
//
//gostrict:host
func (s DurationState) Observe(fn func(v Duration)) DurationState {
	return DurationState{ref: wrap(s.nago().Observe(func(v time.Duration) { fn(durationFrom(v)) }))}
}

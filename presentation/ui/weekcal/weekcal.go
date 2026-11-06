// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package weekcal provides a week calendar, which shows up to seven days as columns over a time axis with the events
// of the days as blocks. The user spans an interval in free time, moves a block, also to another day, changes its
// duration at its bottom edge and clicks a block, with the mouse, a pen, a finger (after holding it for half a
// second) or the keyboard.
//
// Times are minutes since midnight of a day, a wall clock without a zone: a day has always 1440 minutes. A day is
// the index of a column, starting at 0. An event, which crosses midnight, must be split into one event per day, each
// with its own id.
package weekcal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"

	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// MinutesPerDay is the length of a day: the wall clock knows no change of the daylight saving time.
const MinutesPerDay = 1440

// MaxDays is the largest number of days of a calendar.
const MaxDays = 7

// Steps are the allowed rasters in minutes.
var Steps = []int{5, 10, 15, 20, 30, 60}

// Day is the head of a column.
type Day struct {
	Title string
	// Subtitle is shown below the title, with up to two lines, e.g. a holiday or an absence.
	Subtitle string
	// Today marks the current day.
	Today bool
	// Muted marks a day, which is not a working day, e.g. a weekend or a holiday.
	Muted bool
}

// Event is a block of the calendar.
type Event struct {
	// ID identifies the event within the calendar. It is not empty and unique.
	ID string
	// Day is the index of the day of the event.
	Day int
	// From and To are minutes since midnight of the day, 0 <= From < To <= 1440.
	From, To int
	Title    string
	Text     string
	Color    ui.Color
	// Fixed events can only be clicked.
	Fixed bool
}

// Span is an interval, which the user spanned in free time.
type Span struct {
	Day, From, To int
}

// Move is a move of an event or a change of its duration by the user. OldDay, OldFrom and OldTo are the interval,
// which the user saw, when the gesture began, so that an application can recognize an event, which has been changed
// elsewhere in the meantime.
type Move struct {
	ID                     string
	Day, From, To          int
	OldDay, OldFrom, OldTo int
}

// TWeekCalendar is a basic component (WeekCalendar).
//
// Every action of the user is an event and passed to [TWeekCalendar.OnClick], [TWeekCalendar.OnSelect] or
// [TWeekCalendar.OnMove], which may invoke any use case. The calendar does not trust the client: an action is passed
// only, if it is valid for the calendar of the last render of the window, e.g. a moved event must still be shown at
// the interval, from which the user moved it. The calendar renders again afterward and shows whatever the
// application returns then. While an action is pending, the client keeps its preview and takes no new gesture.
type TWeekCalendar struct {
	id         string
	stateID    string
	days       []Day
	events     []Event
	from, to   int
	step       int
	clickSpan  int
	now        bool
	nowDay     int
	nowMinute  int
	hourHeight float64
	disabled   bool
	frame      ui.Frame
	onClick    func(id string)
	onSelect   func(s Span) error
	onMove     func(m Move) error
}

// WeekCalendar creates a calendar with 1 to 7 days, see [MaxDays]. The id must be unique within the window, because
// the calendar keeps its actions in states derived from it. By default, the calendar shows 06:00 to 22:00 in a
// raster of 15 minutes, with 48dp per hour, the full width and the height of its range.
func WeekCalendar(id string, days ...Day) TWeekCalendar {
	return TWeekCalendar{
		id:         id,
		days:       days,
		from:       360,
		to:         1320,
		step:       15,
		clickSpan:  60,
		hourHeight: 48,
		frame:      ui.Frame{Width: ui.Full},
	}
}

// StateID sets the id, from which the states of the calendar derive, if it shall differ from its id, e.g. with a
// prefix of the caller. The id is still shown to the client.
func (c TWeekCalendar) StateID(id string) TWeekCalendar {
	c.stateID = id
	return c
}

// Events sets the events. An event, which lies partly outside of the range, is cut at the edge and marked; events
// entirely outside of it are counted per day at the top or the bottom.
func (c TWeekCalendar) Events(events ...Event) TWeekCalendar {
	c.events = events
	return c
}

// Range sets the visible minutes, 0 <= from < to <= 1440, both in the raster of [TWeekCalendar.Step]. The default
// is 360 to 1320.
func (c TWeekCalendar) Range(from, to int) TWeekCalendar {
	c.from, c.to = from, to
	return c
}

// Step sets the raster in minutes, one of [Steps]. The default is 15.
func (c TWeekCalendar) Step(minutes int) TWeekCalendar {
	c.step = minutes
	return c
}

// ClickSpan sets the duration in minutes, which a click into free time or the enter key on the cursor spans, a
// multiple of [TWeekCalendar.Step]. It is cut at the end of the range. The default is 60.
func (c TWeekCalendar) ClickSpan(minutes int) TWeekCalendar {
	c.clickSpan = minutes
	return c
}

// Now shows the line of the current time in the column of the day at the minute. By default, there is none.
func (c TWeekCalendar) Now(day, minute int) TWeekCalendar {
	c.now, c.nowDay, c.nowMinute = true, day, minute
	return c
}

// HourHeight sets the height of an hour in dp, 24 to 240. The default is 48.
func (c TWeekCalendar) HourHeight(dp float64) TWeekCalendar {
	c.hourHeight = dp
	return c
}

// Disabled shows the calendar without any action.
func (c TWeekCalendar) Disabled(disabled bool) TWeekCalendar {
	c.disabled = disabled
	return c
}

// Frame sets the size of the calendar. By default, it has the full width and the height of its range; with a
// smaller height, it scrolls vertically. If a day gets less than 96dp, the calendar scrolls horizontally, and the
// time axis stays.
func (c TWeekCalendar) Frame(frame ui.Frame) TWeekCalendar {
	c.frame = frame
	return c
}

// OnClick is invoked with the id of an event, which the user clicked. A gesture on an event, which ends without
// a change, is a click. Without it, a click on an event does nothing.
func (c TWeekCalendar) OnClick(fn func(id string)) TWeekCalendar {
	c.onClick = fn
	return c
}

// OnSelect is invoked with an interval, which the user spanned in free time; a click into free time spans
// [TWeekCalendar.ClickSpan]. Without it, the free time cannot be used. An error is shown in the calendar until the
// next action.
func (c TWeekCalendar) OnSelect(fn func(s Span) error) TWeekCalendar {
	c.onSelect = fn
	return c
}

// OnMove is invoked with a move of an event, also to another day, or with a change of its duration. Without it, the
// events cannot be moved. An error is shown in the calendar until the next action; if the application changes
// nothing, the event jumps back to its former place.
func (c TWeekCalendar) OnMove(fn func(m Move) error) TWeekCalendar {
	c.onMove = fn
	return c
}

// Validate returns the first violation of the rules of the days, the events and the settings, or nil: 1 to 7 days;
// each event with an id, which is not empty and unique, a day of the calendar and 0 <= From < To <= 1440; a step of
// [Steps]; a range 0 <= from < to <= 1440 in the raster; a click span, which is a positive multiple of the step; an
// hour height from 24 to 240; the current time at a day of the calendar and a minute from 0 to 1440.
func (c TWeekCalendar) Validate() error {
	if len(c.days) == 0 || len(c.days) > MaxDays {
		return fmt.Errorf("a week calendar has 1 to %d days, not %d", MaxDays, len(c.days))
	}

	if !slices.Contains(Steps, c.step) {
		return fmt.Errorf("the step is one of 5, 10, 15, 20, 30 or 60 minutes, not %d", c.step)
	}

	if c.from < 0 || c.from >= c.to || c.to > MinutesPerDay {
		return fmt.Errorf("the range is 0 <= from < to <= 1440, not %d to %d", c.from, c.to)
	}

	if c.from%c.step != 0 || c.to%c.step != 0 {
		return fmt.Errorf("the range %d to %d is not in the raster of %d minutes", c.from, c.to, c.step)
	}

	if c.clickSpan <= 0 || c.clickSpan%c.step != 0 {
		return fmt.Errorf("the click span is a positive multiple of the step of %d minutes, not %d", c.step, c.clickSpan)
	}

	if math.IsNaN(c.hourHeight) || c.hourHeight < 24 || c.hourHeight > 240 {
		return fmt.Errorf("the hour height is 24 to 240 dp, not %v", c.hourHeight)
	}

	if c.now && (c.nowDay < 0 || c.nowDay >= len(c.days) || c.nowMinute < 0 || c.nowMinute > MinutesPerDay) {
		return fmt.Errorf("the current time is at a day from 0 to %d and a minute from 0 to 1440, not at day %d and minute %d", len(c.days)-1, c.nowDay, c.nowMinute)
	}

	ids := make(map[string]struct{}, len(c.events))
	for i, e := range c.events {
		if e.ID == "" {
			return fmt.Errorf("the event at index %d has an empty id", i)
		}

		if _, ok := ids[e.ID]; ok {
			return fmt.Errorf("the id %q of an event is not unique", e.ID)
		}
		ids[e.ID] = struct{}{}

		if e.Day < 0 || e.Day >= len(c.days) {
			return fmt.Errorf("the event %q has the day %d, but the calendar has the days 0 to %d", e.ID, e.Day, len(c.days)-1)
		}

		if e.From < 0 || e.From >= e.To || e.To > MinutesPerDay {
			return fmt.Errorf("the event %q is not 0 <= from < to <= 1440, but %d to %d", e.ID, e.From, e.To)
		}
	}

	return nil
}

func (c TWeekCalendar) Render(ctx core.RenderContext) core.RenderNode {
	if err := c.Validate(); err != nil {
		// a programming error of the application, which must not show a broken calendar
		slog.Error("invalid week calendar", "id", c.id, "err", err)
		return ui.Text(fmt.Sprintf("week calendar %q: %v", c.id, err)).Render(ctx)
	}

	wnd := ctx.Window()
	sid := c.stateID
	if sid == "" {
		sid = c.id
	}

	prefix := "nago.weekcal." + sid
	ack := core.StateOf[int64](wnd, prefix+".ack")
	msg := core.StateOf[string](wnd, prefix+".message")

	var input proto.Ptr
	if c.onClick != nil || c.onSelect != nil || c.onMove != nil {
		input = core.StateOf[string](wnd, prefix+".input").Observe(func(raw string) {
			c.handle(raw, ack, msg)
		}).Ptr()
	}

	days := make(proto.WeekCalendarDays, 0, len(c.days))
	for _, d := range c.days {
		days = append(days, proto.WeekCalendarDay{
			Title:    proto.Str(d.Title),
			Subtitle: proto.Str(d.Subtitle),
			Today:    proto.Bool(d.Today),
			Muted:    proto.Bool(d.Muted),
		})
	}

	events := make(proto.WeekCalendarEvents, 0, len(c.events))
	for _, e := range c.events {
		events = append(events, proto.WeekCalendarEvent{
			Id:    proto.Str(e.ID),
			Day:   proto.Int(e.Day),
			From:  proto.Int(e.From),
			To:    proto.Int(e.To),
			Title: proto.Str(e.Title),
			Text:  proto.Str(e.Text),
			Color: proto.Color(e.Color),
			Fixed: proto.Bool(e.Fixed),
		})
	}

	return &proto.WeekCalendar{
		Id:         proto.Str(c.id),
		Days:       days,
		Events:     events,
		From:       proto.Int(c.from),
		To:         proto.Int(c.to),
		Step:       proto.Int(c.step),
		ClickSpan:  proto.Int(c.clickSpan),
		ShowNow:    proto.Bool(c.now),
		NowDay:     proto.Int(c.nowDay),
		NowMinute:  proto.Int(c.nowMinute),
		HourHeight: proto.Float(c.hourHeight),
		Input:      input,
		Ack:        proto.Int(ack.Get()),
		Message:    proto.Str(msg.Get()),
		CanClick:   proto.Bool(c.onClick != nil),
		CanSelect:  proto.Bool(c.onSelect != nil),
		CanMove:    proto.Bool(c.onMove != nil),
		Disabled:   proto.Bool(c.disabled),
		Frame:      frameOf(c.frame),
	}
}

// interval is the interval of an event, from which the client began a gesture.
type interval struct {
	Day  int `json:"day"`
	From int `json:"from"`
	To   int `json:"to"`
}

// action is written by the client into the input state.
type action struct {
	Seq  int64     `json:"seq"`
	Kind string    `json:"kind"`
	ID   string    `json:"id"`
	Day  int       `json:"day"`
	From int       `json:"from"`
	To   int       `json:"to"`
	Base *interval `json:"base"`
}

// The kinds of an action.
const (
	kindClick  = "click"
	kindSelect = "select"
	kindMove   = "move"
	kindResize = "resize"
)

// handle checks an action of the client against the calendar of the last render and passes it to its callback. The
// sequence number tells the client, that the action has been processed, also if it has been rejected. The message
// of a former callback lasts until the next action.
func (c TWeekCalendar) handle(raw string, ack *core.State[int64], msg *core.State[string]) {
	var a action
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		return
	}

	ack.Set(a.Seq)
	msg.Set("")

	var err error
	switch c.check(a) {
	case kindClick:
		c.onClick(a.ID)
	case kindSelect:
		err = c.onSelect(Span{Day: a.Day, From: a.From, To: a.To})
	case kindMove, kindResize:
		err = c.onMove(Move{ID: a.ID, Day: a.Day, From: a.From, To: a.To, OldDay: a.Base.Day, OldFrom: a.Base.From, OldTo: a.Base.To})
	}

	if err != nil {
		msg.Set(errorText(err))
	}
}

// check returns the kind of the action, if it is valid for the calendar, and "" otherwise.
func (c TWeekCalendar) check(a action) string {
	if c.disabled || a.Day < 0 || a.Day >= len(c.days) {
		return ""
	}

	switch a.Kind {
	case kindClick:
		if c.onClick == nil {
			return ""
		}

		if _, ok := c.event(a.ID); !ok {
			return ""
		}
	case kindSelect:
		if c.onSelect == nil || !c.inRaster(a.From) || !c.inRaster(a.To) || a.From >= a.To || !c.inRange(a.From, a.To) {
			return ""
		}
	case kindMove, kindResize:
		if c.onMove == nil || a.Base == nil {
			return ""
		}

		e, ok := c.event(a.ID)
		if !ok || e.Fixed || !c.inRange(e.From, e.To) || *a.Base != (interval{Day: e.Day, From: e.From, To: e.To}) {
			return ""
		}

		if a.Kind == kindMove {
			if a.To-a.From != e.To-e.From || (a.From-e.From)%c.step != 0 || !c.inRange(a.From, a.To) {
				return ""
			}
		} else if a.Day != e.Day || a.From != e.From || !c.inRaster(a.To) || a.To <= a.From || a.To > c.to {
			return ""
		}
	default:
		return ""
	}

	return a.Kind
}

func (c TWeekCalendar) event(id string) (Event, bool) {
	for _, e := range c.events {
		if e.ID == id {
			return e, true
		}
	}

	return Event{}, false
}

func (c TWeekCalendar) inRaster(minute int) bool {
	return minute%c.step == 0
}

func (c TWeekCalendar) inRange(from, to int) bool {
	return from >= c.from && to <= c.to
}

// errorText is the text of an error of a callback, which the calendar shows.
func errorText(err error) string {
	var localized std.LocalizedError
	if errors.As(err, &localized) {
		if localized.Description() != "" {
			return localized.Description()
		}
		return localized.Title()
	}

	return err.Error()
}

func frameOf(f ui.Frame) proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
		FlexProperties: proto.FlexProperties{
			Grow:          proto.Bool(f.FlexGrow),
			PreventShrink: proto.Bool(f.FlexPreventShrink),
		},
	}
}

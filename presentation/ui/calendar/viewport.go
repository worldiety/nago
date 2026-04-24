// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"math"
	"strconv"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/ui"
)

// Column represents a single labeled column in the timeline (e.g. a month).
type Column struct {
	Label string
}

type viewPortStyle int

const (
	vpYear viewPortStyle = iota
	vpMonth
)

// ViewPort defines the visible time range of the calendar, including start/end,
// displayed columns, lane width, and a label for the lane.
type ViewPort struct {
	From      time.Time
	To        time.Time
	Columns   []Column
	LaneWidth Percent
	LaneLabel func(bnd i18n.Bundler) string
	style     viewPortStyle
	TimeStyle SeqPillTimeHint
}

// Year creates a ViewPort for a given year, spanning from January to December
// and dividing the range into 12 monthly columns.
func Year(year int) ViewPort {
	return ViewPort{
		From: time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local),
		To:   time.Date(year, time.December, 31, 24, 0, 0, -1, time.Local),
		Columns: []Column{
			{Label: "Januar"}, {Label: "Februar"}, {Label: "März"}, {Label: "April"}, {Label: "Mai"}, {Label: "Juni"}, {Label: "Juli"}, {Label: "August"}, {Label: "September"}, {Label: "October"}, {Label: "November"}, {Label: "Dezember"},
		},
		LaneWidth: 16,
		LaneLabel: func(bnd i18n.Bundler) string {
			return strconv.Itoa(year)
		},
		style:     vpYear,
		TimeStyle: PillTimeYYYYMMDD,
	}
}

func Day(year int, month time.Month, day int) ViewPort {
	from := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return ViewPort{
		From: from,
		To:   time.Date(year, month, day, 24, 0, 0, -1, time.Local),
		Columns: []Column{
			{Label: "Day"},
		},
		LaneWidth: 16,
		LaneLabel: func(bnd i18n.Bundler) string {
			var dayLabel string
			switch from.Weekday() {
			case time.Sunday:
				dayLabel = rstring.LabelSunday.Get(bnd)
			case time.Monday:
				dayLabel = rstring.LabelMonday.Get(bnd)
			case time.Tuesday:
				dayLabel = rstring.LabelTuesday.Get(bnd)
			case time.Wednesday:
				dayLabel = rstring.LabelWednesday.Get(bnd)
			case time.Thursday:
				dayLabel = rstring.LabelThursday.Get(bnd)
			case time.Friday:
				dayLabel = rstring.LabelFriday.Get(bnd)
			case time.Saturday:
				dayLabel = rstring.LabelSaturday.Get(bnd)
			}

			return dayLabel + ", " + strconv.Itoa(day) + "." + strconv.Itoa(int(month)) + "." + strconv.Itoa(year)
		},
		style:     vpMonth,
		TimeStyle: PillTimeHHMM,
	}
}

// Percent returns the relative position of a given time within the ViewPort,
// expressed as a percentage (0–100).
func (i ViewPort) Percent(t time.Time) Percent {
	start := float64(i.From.Unix())
	end := float64(i.To.Unix())
	now := float64(t.Unix())
	p := (now - start) / (end - start) * 100
	p = math.Round(p*100) / 100
	return Percent(p)
}

// Percent is between 0 and 100
type Percent float64

// String rounds and formats the percent to at most 2 decimal places.
func (p Percent) String() string {
	return strconv.FormatFloat(float64(p), 'f', -1, 64) + "%"
}

// Length converts the Percent value into a ui.Length,
// which can be directly used for sizing in UI components.
func (p Percent) Length() ui.Length {
	return ui.Length(p.String())
}

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

	"go.wdy.de/nago/presentation/ui"
)

// Column represents a single labeled column in the timeline (e.g. a month).
type Column struct {
	Label string
}

// ViewPort defines the visible time range of the calendar, including start/end,
// displayed columns, lane width, and a label for the lane.
type ViewPort struct {
	From      time.Time
	To        time.Time
	Columns   []Column
	LaneWidth Percent
	LaneLabel string
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
		LaneLabel: strconv.Itoa(year),
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

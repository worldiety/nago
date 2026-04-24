// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"time"

	"go.wdy.de/nago/application/color"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Category represents a classification for events, with a label and a color for visualization.
type Category struct {
	Color ui.Color // The color associated with this category.
	Label string   // The text label describing the category.
}

// Instant represents a point in time with an optional offset (pre-/post-duration).
type Instant struct {
	At     time.Time // The base timestamp of the instant.
	Offset Offset    // Additional offset information (e.g., travel time).
}

// Offset represents a time span relative to an Instant, often visualized with an icon and label.
type Offset struct {
	Duration time.Duration // Offset duration relative to the instant, may indicate travel or preparation.
	Icon     core.SVG      // Icon representing this offset (e.g., bus, train).
	Label    string        // Descriptive label for the offset.
}

// Lane represents a track or row in the calendar/timeline to group events (e.g., separate resources).
type Lane struct {
	Label string // The label displayed for the lane.
}

// Event represents a calendar entry or scheduled activity with metadata.
type Event struct {
	From        Instant // From is inclusive and Offset is e.g. the travel time by bus
	To          Instant // To is inclusive and Offset is e.g. the travel time by train
	Label       string  // Title of the event
	Action      func()  // Action if clicked on the event
	Category    Category
	Lane        Lane
	Chips       []Chip
	Organiser   string
	Location    string
	IsCancelled bool                  // True if event is cancelled by the host
	Render      func(Style) core.View // custom render func, may be nil to render the default way
}

// Chip can be used to display additional event information,
// such as the current number of participants or whether
// sign-up for the waiting list is available.
// Currently supported alignments are TopTrailing and BottomLeading.
type Chip struct {
	Label       string // Text within the chip
	Icon        core.SVG
	FillColor   color.Color // If set, the icon is filled with this color
	StrokeColor color.Color // If set, the icon stroke is drawn with this color
	BgColor     color.Color
	TextColor   color.Color
	Alignment   ui.Alignment
	FullWidth   bool
}

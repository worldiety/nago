// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"time"
)

// Occupation represents how much capacity of a resource is used within an event (e.g., seats in a bus).
type Occupation struct {
	Current int // The current number of occupied units.
	Max     int // The maximum capacity available.
}

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
	From       Instant    // Start time of the event (inclusive), may include travel/prep offset.
	To         Instant    // End time of the event (inclusive), may include travel/post offset.
	Label      string     // Title or description of the event.
	Action     func()     // Callback executed when the event is clicked.
	Category   Category   // Category classification (e.g., work, private, travel).
	Lane       Lane       // Lane (row/track) the event belongs to.
	Occupation Occupation // Information about resource usage (current/max capacity).
}

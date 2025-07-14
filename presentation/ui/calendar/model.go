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

type Occupation struct {
	Current int
	Max     int
}

type Category struct {
	Color ui.Color
	Label string
}

type Instant struct {
	At     time.Time
	Offset Offset
}

type Offset struct {
	Duration time.Duration // offset from At which may be context-dependent a pre- or post-duration.
	Icon     core.SVG
	Label    string
}

type Lane struct {
	Label string
}

type Event struct {
	From       Instant // From is inclusive and Offset is e.g. the travel time by bus
	To         Instant // To is inclusive and Offset is e.g. the travel time by train
	Label      string  // Title of the event
	Action     func()  // Action if clicked on the event
	Category   Category
	Lane       Lane
	Occupation Occupation
}

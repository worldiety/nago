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
	"time"

	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// CountDownStyle is the visual style of a countdown, see CountDown.
type CountDownStyle int32

const (
	// CountDownStyleClock shows the remaining time as a clock.
	CountDownStyleClock CountDownStyle = 0
	// CountDownStyleProgress shows the remaining time as a progress indicator.
	CountDownStyleProgress CountDownStyle = 1
)

// TCountDown is a timer, which counts down from a duration, see CountDown.
type TCountDown struct{ ref js.Ref }

func (TCountDown) isView() {}

// CountDown returns a countdown, which counts down from the duration in milliseconds. A negative duration counts from
// zero. By default, days, hours, minutes and seconds are shown.
//
//gostrict:host
func CountDown(millis int32) TCountDown {
	return TCountDown{ref: wrap(nui.CountDown(time.Duration(millis) * time.Millisecond))}
}

// Style sets the visual style of the countdown.
//
//gostrict:host
func (c TCountDown) Style(style CountDownStyle) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Style(nui.CountDownStyle(style)))}
}

// Action calls action, when the countdown ends.
//
//gostrict:host
func (c TCountDown) Action(action func()) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Action(action))}
}

// Days shows or hides the days.
//
//gostrict:host
func (c TCountDown) Days(show bool) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Days(show))}
}

// Hours shows or hides the hours.
//
//gostrict:host
func (c TCountDown) Hours(show bool) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Hours(show))}
}

// Minutes shows or hides the minutes.
//
//gostrict:host
func (c TCountDown) Minutes(show bool) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Minutes(show))}
}

// Seconds shows or hides the seconds.
//
//gostrict:host
func (c TCountDown) Seconds(show bool) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Seconds(show))}
}

// Frame sets the size and the position of the countdown.
//
//gostrict:host
func (c TCountDown) Frame(frame Frame) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Frame(frame.nago()))}
}

// TextColor sets the color of the countdown text.
//
//gostrict:host
func (c TCountDown) TextColor(color Color) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).TextColor(nui.Color(color)))}
}

// SeparatorColor sets the color of the separators, e.g. of the colons.
//
//gostrict:host
func (c TCountDown) SeparatorColor(color Color) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).SeparatorColor(nui.Color(color)))}
}

// Done marks the countdown as finished, which overrides its active state.
//
//gostrict:host
func (c TCountDown) Done(done bool) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).Done(done))}
}

// ProgressBackground sets the background color of the progress indicator.
//
//gostrict:host
func (c TCountDown) ProgressBackground(background Color) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).ProgressBackground(nui.Color(background)))}
}

// ProgressColor sets the foreground color of the progress indicator.
//
//gostrict:host
func (c TCountDown) ProgressColor(foreground Color) TCountDown {
	return TCountDown{ref: wrap(get[nui.TCountDown](c.ref).ProgressColor(nui.Color(foreground)))}
}

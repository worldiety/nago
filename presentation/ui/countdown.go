// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"time"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type CountDownStyle int

const (
	CountDownStyleClock CountDownStyle = iota
	CountDownStyleProgress
)

// TCountDown is a composite component (Countdown).
// It displays a timer counting down from a specified duration,
// optionally showing days, hours, minutes, and seconds. The component
// supports custom colors, styling, progress indicators, and an action
// callback to be executed when the countdown completes.
type TCountDown struct {
	children           []core.View    // child views for rendering countdown parts
	action             func()         // callback executed when the countdown ends
	duration           time.Duration  // total duration of the countdown
	frame              Frame          // layout frame for size and positioning
	showDays           bool           // whether to display days
	showHours          bool           // whether to display hours
	showMinutes        bool           // whether to display minutes
	showSeconds        bool           // whether to display seconds
	textColor          Color          // color of the countdown text
	separatorColor     Color          // color of the separators (e.g., colon)
	style              CountDownStyle // visual style of the countdown
	done               bool           // true if countdown has finished
	progressBackground Color          // background color of the progress indicator
	progressForeground Color          // foreground color of the progress indicator
}

// CountDown creates a new countdown timer initialized with the given duration.
// By default, days, hours, minutes, and seconds are all displayed.
func CountDown(duration time.Duration) TCountDown {
	return TCountDown{
		duration:    duration,
		showDays:    true,
		showHours:   true,
		showMinutes: true,
		showSeconds: true,
	}
}

// Style sets the visual style of the countdown (e.g., text-only or with progress).
func (c TCountDown) Style(style CountDownStyle) TCountDown {
	c.style = style
	return c
}

// Action sets the callback function to be executed when the countdown ends.
func (c TCountDown) Action(action func()) TCountDown {
	c.action = action
	return c
}

// Days toggles whether the countdown displays days.
func (c TCountDown) Days(show bool) TCountDown {
	c.showDays = show
	return c
}

// Hours toggles whether the countdown displays hours.
func (c TCountDown) Hours(show bool) TCountDown {
	c.showHours = show
	return c
}

// Minutes toggles whether the countdown displays minutes.
func (c TCountDown) Minutes(show bool) TCountDown {
	c.showMinutes = show
	return c
}

// Seconds toggles whether the countdown displays seconds.
func (c TCountDown) Seconds(show bool) TCountDown {
	c.showSeconds = show
	return c
}

// Frame sets the layout frame of the countdown, including size and positioning.
func (c TCountDown) Frame(frame Frame) TCountDown {
	c.frame = frame
	return c
}

// TextColor sets the color of the countdown text.
func (c TCountDown) TextColor(color Color) TCountDown {
	c.textColor = color
	return c
}

// SeparatorColor sets the color of separators (e.g., colons) in the countdown display.
func (c TCountDown) SeparatorColor(color Color) TCountDown {
	c.separatorColor = color
	return c
}

// Done marks the countdown as finished, overriding its active state.
func (c TCountDown) Done(done bool) TCountDown {
	c.done = done
	return c
}

// ProgressBackground sets the background color of the countdown's progress indicator.
func (c TCountDown) ProgressBackground(background Color) TCountDown {
	c.progressBackground = background
	return c
}

// ProgressColor sets the foreground color of the countdown's progress indicator.
func (c TCountDown) ProgressColor(foreground Color) TCountDown {
	c.progressForeground = foreground
	return c
}

// Render builds and returns the protocol representation of the countdown.
func (c TCountDown) Render(ctx core.RenderContext) core.RenderNode {
	if c.separatorColor == "" {
		c.separatorColor = ColorLine
	}

	if c.textColor == "" {
		c.textColor = ColorText
	}

	return &proto.CountDown{
		Action:             ctx.MountCallback(c.action),
		Frame:              c.frame.ora(),
		Duration:           proto.DurationSec(c.duration.Seconds()),
		ShowDays:           proto.Bool(c.showDays),
		ShowHours:          proto.Bool(c.showHours),
		ShowMinutes:        proto.Bool(c.showMinutes),
		ShowSeconds:        proto.Bool(c.showSeconds),
		TextColor:          proto.Color(c.textColor),
		SeparatorColor:     proto.Color(c.separatorColor),
		Style:              proto.CountDownStyle(c.style),
		Done:               proto.Bool(c.done),
		ProgressBackground: proto.Color(c.progressBackground),
		ProgressColor:      proto.Color(c.progressForeground),
	}
}

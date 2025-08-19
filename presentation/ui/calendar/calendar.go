// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"time"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Style int

const (
	TimelineYear Style = iota
	TimelineDay
	StartTimeSequence
)

// TCalendar is a composite component (Calendar).
// This component displays calendar data in different styles (e.g., monthly or weekly view)
// and supports rendering events within a defined viewport.
// It allows customization of frame, style, and colors to adapt to different use cases.
type TCalendar struct {
	style  Style
	events []Event
	frame  ui.Frame
	vp     ViewPort
	colors Colors
}

// Calendar creates a new TCalendar initialized with the current year, a yearly timeline style, and default colors.
func Calendar(events ...Event) TCalendar {
	now := time.Now()
	return TCalendar{
		events: events,
		style:  TimelineYear,
		vp:     Year(now.Year()),
		colors: DefaultColors(),
	}
}

// Style sets the display style (e.g., timeline view) for the calendar.
func (c TCalendar) Style(style Style) TCalendar {
	c.style = style
	return c
}

// Append adds one or more events to the existing calendar events.
func (c TCalendar) Append(events ...Event) TCalendar {
	c.events = append(c.events, events...)
	return c
}

// Frame defines the layout frame (size, width, height) for the calendar component.
func (c TCalendar) Frame(frame ui.Frame) TCalendar {
	c.frame = frame
	return c
}

// ViewPort sets the visible time range (e.g., year, month) of the calendar.
func (c TCalendar) ViewPort(vp ViewPort) TCalendar {
	c.vp = vp
	return c
}

// Colors customizes the color scheme used for rendering the calendar and its events.
func (c TCalendar) Colors(colors Colors) TCalendar {
	c.colors = colors
	return c
}

// Render renders the calendar component based on the selected style and configuration.
func (c TCalendar) Render(ctx core.RenderContext) core.RenderNode {
	switch c.style {
	case StartTimeSequence:
		return renderStartTimeSequence(c, ctx)
	default:
		return renderTimelineYear(c, ctx)
	}
}

// Colors defines the color scheme used for different calendar elements such as headers, events, and backgrounds.
type Colors struct {
	Header               ui.Color
	LaneBackground       ui.Color
	PrePostBackground    ui.Color
	PrePostForeground    ui.Color
	EventBackground      ui.Color
	EventHoverBackground ui.Color
	Text                 ui.Color
	EventText            ui.Color
	Separator            ui.Color
}

// DefaultColors returns the standard color scheme for calendars with predefined UI system colors.
func DefaultColors() Colors {
	return Colors{
		Header:               ui.ColorCardTop,
		LaneBackground:       ui.ColorCardBody,
		EventBackground:      ui.ColorIconsMuted,
		PrePostBackground:    ui.M8,
		PrePostForeground:    ui.ColorBackground,
		Text:                 ui.M8,
		EventText:            ui.ColorBackground,
		Separator:            ui.ColorLine,
		EventHoverBackground: ui.ColorInteractive,
	}
}

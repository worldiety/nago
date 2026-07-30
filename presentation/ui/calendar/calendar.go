// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package calendar

import (
	"fmt"
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
	style         Style
	events        []Event
	frame         ui.Frame
	vp            ViewPort
	colors        Colors
	maxCategories int
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

func (c TCalendar) FullWidth() TCalendar {
	c.frame = c.frame.FullWidth()
	return c
}

// MaxCategories limits how many category color bars are rendered per event.
// A value <= 0 means unlimited (all categories are shown).
func (c TCalendar) MaxCategories(n int) TCalendar {
	c.maxCategories = n
	return c
}

// limitCategories truncates the given categories to at most max entries.
// A max value <= 0 means no limit.
func limitCategories(cats []Category, max int) []Category {
	if max > 0 && len(cats) > max {
		return cats[:max]
	}
	return cats
}

// categoryBar renders the categories of an event as a single vertical bar that
// is split into equally sized segments stacked on top of each other. With two
// categories the bar is split in half, with three into thirds and so on. The
// number of visible categories is limited by max (<= 0 means unlimited).
func categoryBar(cats []Category, max int) core.View {
	cats = limitCategories(cats, max)
	n := len(cats)

	segHeight := ui.Full
	if n > 0 {
		segHeight = ui.Length(fmt.Sprintf("%f%%", 100/float64(n)))
	}

	return ui.VStack(
		ui.ForEach(cats, func(cat Category) core.View {
			return ui.HStack().
				BackgroundColor(cat.Color).
				Frame(ui.Frame{Width: ui.Full, Height: segHeight}).
				AccessibilityLabel(cat.Label)
		})...,
	).Frame(ui.Frame{MinWidth: ui.L12, Height: ui.Full})
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

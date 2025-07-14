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

type Style int

const (
	TimelineYear Style = iota
	TimelineDay
	StartTimeSequence
)

type TCalendar struct {
	style  Style
	events []Event
	frame  ui.Frame
	vp     ViewPort
	colors Colors
}

func Calendar(events ...Event) TCalendar {
	now := time.Now()
	return TCalendar{
		events: events,
		style:  TimelineYear,
		vp:     Year(now.Year()),
		colors: DefaultColors(),
	}
}

func (c TCalendar) Style(style Style) TCalendar {
	c.style = style
	return c
}

func (c TCalendar) Append(events ...Event) TCalendar {
	c.events = append(c.events, events...)
	return c
}

func (c TCalendar) Frame(frame ui.Frame) TCalendar {
	c.frame = frame
	return c
}

func (c TCalendar) ViewPort(vp ViewPort) TCalendar {
	c.vp = vp
	return c
}

func (c TCalendar) Colors(colors Colors) TCalendar {
	c.colors = colors
	return c
}

func (c TCalendar) Render(ctx core.RenderContext) core.RenderNode {
	switch c.style {
	default:
		return renderTimelineYear(c, ctx)
	}
}

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

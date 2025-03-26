// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"time"
)

type TCountDown struct {
	children       []core.View
	action         func()
	duration       time.Duration
	frame          Frame
	showDays       bool
	showHours      bool
	showMinutes    bool
	showSeconds    bool
	textColor      Color
	separatorColor Color
}

func CountDown(duration time.Duration) TCountDown {
	return TCountDown{
		duration:    duration,
		showDays:    true,
		showHours:   true,
		showMinutes: true,
		showSeconds: true,
	}
}

func (c TCountDown) Action(action func()) TCountDown {
	c.action = action
	return c
}

func (c TCountDown) Days(show bool) TCountDown {
	c.showDays = show
	return c
}

func (c TCountDown) Hours(show bool) TCountDown {
	c.showHours = show
	return c
}

func (c TCountDown) Minutes(show bool) TCountDown {
	c.showMinutes = show
	return c
}

func (c TCountDown) Seconds(show bool) TCountDown {
	c.showSeconds = show
	return c
}

func (c TCountDown) Frame(frame Frame) TCountDown {
	c.frame = frame
	return c
}

func (c TCountDown) TextColor(color Color) TCountDown {
	c.textColor = color
	return c
}

func (c TCountDown) SeparatorColor(color Color) TCountDown {
	c.separatorColor = color
	return c
}

func (c TCountDown) Render(ctx core.RenderContext) core.RenderNode {
	if c.separatorColor == "" {
		c.separatorColor = ColorLine
	}

	if c.textColor == "" {
		c.textColor = ColorText
	}

	return &proto.CountDown{
		Action:         ctx.MountCallback(c.action),
		Frame:          c.frame.ora(),
		Duration:       proto.DurationSec(c.duration.Seconds()),
		ShowDays:       proto.Bool(c.showDays),
		ShowHours:      proto.Bool(c.showHours),
		ShowMinutes:    proto.Bool(c.showMinutes),
		ShowSeconds:    proto.Bool(c.showSeconds),
		TextColor:      proto.Color(c.textColor),
		SeparatorColor: proto.Color(c.separatorColor),
	}
}

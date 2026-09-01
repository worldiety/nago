// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type TFieldset struct {
	children []core.View
	title    string
	frame    ui.Frame
}

// Fieldset creates a new fieldset containing the given child views.
func Fieldset(children ...core.View) TFieldset {
	return TFieldset{
		children: children,
	}
}

// Title sets the title of the fieldset, which can be used for labeling grouped input fields.
func (c TFieldset) Title(title string) TFieldset {
	c.title = title
	return c
}

// Frame sets the layout frame of the fieldset, including size and positioning.
func (c TFieldset) Frame(frame ui.Frame) TFieldset {
	c.frame = frame
	return c
}

func (c TFieldset) Render(ctx core.RenderContext) core.RenderNode {
	children := make([]proto.Component, 0)
	for _, child := range c.children {
		children = append(children, child.Render(ctx))
	}

	return &proto.Fieldset{
		Children: children,
		Title:    proto.Str(c.title),
		Frame: proto.Frame{
			MinWidth:  proto.Length(c.frame.MinWidth),
			MaxWidth:  proto.Length(c.frame.MaxWidth),
			MinHeight: proto.Length(c.frame.MinHeight),
			MaxHeight: proto.Length(c.frame.MaxHeight),
			Width:     proto.Length(c.frame.Width),
			Height:    proto.Length(c.frame.Height),
			FlexProperties: proto.FlexProperties{
				Grow:          proto.Bool(c.frame.FlexGrow),
				PreventShrink: proto.Bool(c.frame.FlexPreventShrink),
			},
		},
	}
}

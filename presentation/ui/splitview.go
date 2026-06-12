// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type SplitViewOrientation int

const (
	SplitViewOrientationHorizontal SplitViewOrientation = SplitViewOrientation(proto.Horizontal)
	SplitViewOrientationVertical   SplitViewOrientation = SplitViewOrientation(proto.Vertical)
)

type TSplitView struct {
	value       float64              // current split ratio (0.0 to 1.0)
	inputValue  *core.State[float64] // optional external state for controlled component behavior
	orientation SplitViewOrientation // orientation of the split view
	frame       Frame                // layout frame
	contentA    core.View            // left/top content
	contentB    core.View            // right/bottom content
	minRatio    float64              // sets a lower limit for the ratio
	maxRatio    float64              // sets an upper limit for the ratio
}

func SplitView(contentA core.View, contentB core.View) TSplitView {
	return TSplitView{
		value:       0.5,
		orientation: SplitViewOrientationHorizontal,
		contentA:    contentA,
		contentB:    contentB,
	}
}

func (c TSplitView) Value(value float64) TSplitView {
	c.value = value
	return c
}

func (c TSplitView) InputValue(state *core.State[float64]) TSplitView {
	c.inputValue = state
	return c
}

func (c TSplitView) Orientation(orientation SplitViewOrientation) TSplitView {
	c.orientation = orientation
	return c
}

func (c TSplitView) Frame(frame Frame) TSplitView {
	c.frame = frame
	return c
}

func (c TSplitView) ContentA(contentA core.View) TSplitView {
	c.contentA = contentA
	return c
}

func (c TSplitView) ContentB(contentB core.View) TSplitView {
	c.contentB = contentB
	return c
}

func (c TSplitView) MinRatio(minRatio float64) TSplitView {
	c.minRatio = minRatio
	return c
}

func (c TSplitView) MaxRatio(maxRatio float64) TSplitView {
	c.maxRatio = maxRatio
	return c
}

func (c TSplitView) Render(ctx core.RenderContext) core.RenderNode {
	value := c.value
	if c.inputValue != nil {
		value = c.inputValue.Get()
	}

	return &proto.SplitView{
		InputValue:  c.inputValue.Ptr(),
		Value:       proto.Float(value),
		ContentA:    c.contentA.Render(ctx),
		ContentB:    c.contentB.Render(ctx),
		Frame:       c.frame.ora(),
		Orientation: proto.Orientation(c.orientation),
		MinRatio:    proto.Float(c.minRatio),
		MaxRatio:    proto.Float(c.maxRatio),
	}
}

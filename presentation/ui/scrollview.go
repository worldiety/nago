package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ScrollViewAxis int

func (a ScrollViewAxis) ora() ora.ScrollViewAxis {
	if a == ScrollViewAxisHorizontal {
		return "h"
	}

	return "v"
}

const (
	ScrollViewAxisVertical ScrollViewAxis = iota
	ScrollViewAxisHorizontal
)

type TScrollView struct {
	content core.View
	axis    ScrollViewAxis
	frame   Frame
}

// A ScrollView can either be horizontal or vertical. By default, it is vertical.
func ScrollView(content core.View) TScrollView {
	return TScrollView{
		content: content,
		axis:    ScrollViewAxisVertical,
	}
}

func (c TScrollView) Axis(axis ScrollViewAxis) TScrollView {
	c.axis = axis
	return c
}

func (c TScrollView) Frame(frame Frame) TScrollView {
	c.frame = frame
	return c
}

func (c TScrollView) Render(ctx core.RenderContext) core.RenderNode {
	return ora.ScrollView{
		Type:    ora.ScrollViewT,
		Content: render(ctx, c.content),
		Axis:    c.axis.ora(),
		Frame:   c.frame.ora(),
	}
}

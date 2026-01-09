// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
)

type StackLayout int

const (
	StackLayoutAuto StackLayout = iota
	StackLayoutVertical
	StackLayoutHorizontal
	StackLayoutWrap
)

const StackLayoutFlex = StackLayoutWrap

// TStack is a layout component(Stack).
// It is responsive and can switch between [HStack] and [VStack] during rendering.
type TStack struct {
	children        []core.View
	alignment       Alignment
	backgroundColor Color
	frame           Frame
	gap             Length
	padding         Padding
	border          Border
	layout          StackLayout
	noClip          bool

	adapt func(wnd core.Window, stack TStack) TStack
}

// Stack is a responsive variant which decides between [VStack] and [HStack].
func Stack(children ...core.View) TStack {
	return TStack{
		children: children,
	}
}

func (c TStack) NoClip(b bool) TStack {
	c.noClip = b
	return c
}

func (c TStack) Append(children ...core.View) TStack {
	c.children = append(c.children, children...)
	return c
}

func (c TStack) BackgroundColor(color Color) TStack {
	c.backgroundColor = color
	return c
}

func (c TStack) Alignment(alignment Alignment) TStack {
	c.alignment = alignment
	return c
}

func (c TStack) Responsive(fn func(wnd core.Window, stack TStack) TStack) TStack {
	c.adapt = fn
	return c
}

func (c TStack) Frame(frame Frame) TStack {
	c.frame = frame
	return c
}

func (c TStack) Gap(gap Length) TStack {
	c.gap = gap
	return c
}

func (c TStack) Padding(padding Padding) TStack {
	c.padding = padding
	return c
}

func (c TStack) Border(border Border) TStack {
	c.border = border
	return c
}

func (c TStack) Layout(layout StackLayout) TStack {
	c.layout = layout
	return c
}

func (c TStack) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	if c.adapt != nil {
		c = c.adapt(ctx.Window(), c)
	}

	layout := c.layout
	if layout == StackLayoutAuto {
		if wnd.Info().SizeClass < core.SizeClassMedium {
			layout = StackLayoutVertical
		} else {
			layout = StackLayoutHorizontal
		}
	}

	switch layout {
	case StackLayoutVertical:
		return VStack(c.children...).
			BackgroundColor(c.backgroundColor).
			Gap(c.gap).
			NoClip(c.noClip).
			Alignment(c.alignment).
			Frame(c.frame).
			Padding(c.padding).
			Border(c.border).
			Render(ctx)
	default:
		wrap := c.layout == StackLayoutWrap
		return HStack(c.children...).
			BackgroundColor(c.backgroundColor).
			Gap(c.gap).
			NoClip(c.noClip).
			Alignment(c.alignment).
			Wrap(wrap).
			Frame(c.frame).
			Padding(c.padding).
			Border(c.border).
			Render(ctx)

	}
}

func (c TStack) FullWidth() TStack {
	c.frame.Width = Full
	return c
}

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
)

// TFixedSpacer is a utility component(Fixed Spacer).
type TFixedSpacer struct {
	width  Length
	height Length
}

// FixedSpacer returns an empty view with the given dimensions.
func FixedSpacer(width, height Length) TFixedSpacer {
	return TFixedSpacer{width: width, height: height}
}

func (t TFixedSpacer) Render(ctx core.RenderContext) core.RenderNode {
	/*return VStack(
		// double wrap, to trick the CSS flexbox (mis) behavior
		VStack().Frame(Frame{Width: width, Height: height}),
	)*/

	// trying to simplify the above: is this even more correct?
	return VStack().Frame(Frame{MinWidth: t.width, MaxWidth: t.width, MinHeight: t.height, MaxHeight: t.height}).Render(ctx)
}

// TSpace is a layout component(Space).
type TSpace struct {
	size Length
}

func Space(size Length) TSpace {
	return TSpace{size: size}
}

func (t TSpace) Render(ctx core.RenderContext) core.RenderNode {
	return FixedSpacer(t.size, t.size).Render(ctx)
}

// TSpacer is a layout component(Spacer).
type TSpacer struct {
	backgroundColor proto.Color
	frame           proto.Frame
	border          proto.Border
}

// Spacer is used in VStack or HStack to grow and shrink as required.
func Spacer() TSpacer {
	return TSpacer{}
}

func (c TSpacer) BackgroundColor(backgroundColor Color) TSpacer {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TSpacer) Frame(frame Frame) TSpacer {
	c.frame = frame.ora()
	return c
}

func (c TSpacer) Border(border Border) {
	c.border = border.ora()
}

func (c TSpacer) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Spacer{
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Border:          c.border,
	}
}

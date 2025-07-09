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

type TVStack struct {
	children               []core.View
	alignment              proto.Alignment
	backgroundColor        proto.Color
	textColor              proto.Color
	hoveredBackgroundColor proto.Color
	pressedBackgroundColor proto.Color
	focusedBackgroundColor proto.Color
	frame                  Frame
	gap                    proto.Length
	padding                proto.Padding
	border                 proto.Border
	hoveredBorder          proto.Border
	focusedBorder          proto.Border
	pressedBorder          proto.Border
	stylePreset            proto.StylePreset

	invisible bool
	font      proto.Font
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string
	action             func()
	position           Position
	id                 string
	noClip             bool
}

// VStack is a container, in which the given children will be layout in a column according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func VStack(children ...core.View) TVStack {
	c := TVStack{
		children: children,
	}
	return c
}

func (c TVStack) Append(children ...core.View) TVStack {
	c.children = append(c.children, children...)
	return c
}

func (c TVStack) Gap(gap Length) TVStack {
	c.gap = gap.ora()
	return c
}

func (c TVStack) Position(position Position) TVStack {
	c.position = position
	return c
}

func (c TVStack) StylePreset(preset StylePreset) TVStack {
	c.stylePreset = preset.ora()
	return c
}

func (c TVStack) TextColor(textColor Color) TVStack {
	c.textColor = textColor.ora()
	return c
}

func (c TVStack) BackgroundColor(backgroundColor Color) TVStack {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TVStack) HoveredBackgroundColor(backgroundColor Color) TVStack {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

func (c TVStack) PressedBackgroundColor(backgroundColor Color) TVStack {
	c.pressedBackgroundColor = backgroundColor.ora()
	return c
}

func (c TVStack) FocusedBackgroundColor(backgroundColor proto.Color) TVStack {
	c.focusedBackgroundColor = backgroundColor
	return c
}

func (c TVStack) Action(f func()) TVStack {
	c.action = f
	return c
}

func (c TVStack) NoClip(b bool) TVStack {
	c.noClip = b
	return c
}

func (c TVStack) Alignment(alignment Alignment) TVStack {
	c.alignment = alignment.ora()
	return c
}

func (c TVStack) Font(font Font) TVStack {
	c.font = font.ora()
	return c
}

func (c TVStack) Frame(f Frame) DecoredView {
	c.frame = f
	return c
}

func (c TVStack) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TVStack) FullWidth() TVStack {
	c.frame.Width = "100%"
	return c
}

func (c TVStack) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TVStack) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TVStack) HoveredBorder(border Border) TVStack {
	c.hoveredBorder = border.ora()
	return c
}

func (c TVStack) PressedBorder(border Border) TVStack {
	c.pressedBorder = border.ora()
	return c
}

func (c TVStack) FocusedBorder(border Border) TVStack {
	c.focusedBorder = border.ora()
	return c
}

func (c TVStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TVStack) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TVStack) ID(id string) TVStack {
	c.id = id
	return c
}

func (c TVStack) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.VStack{
		Children:           renderComponents(ctx, c.children),
		Frame:              c.frame.ora(),
		Alignment:          c.alignment,
		BackgroundColor:    c.backgroundColor,
		Gap:                c.gap,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Font:               c.font,
		StylePreset:        c.stylePreset,
		TextColor:          c.textColor,

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		Action:                 ctx.MountCallback(c.action),
		Position:               c.position.ora(),
		Id:                     proto.Str(c.id),
		NoClip:                 proto.Bool(c.noClip),
	}
}

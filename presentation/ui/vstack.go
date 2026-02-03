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

// TVStack is a layout component (VStack).
// VStack is a vertical layout container that arranges its child views in a column.
// It supports alignment, spacing, background styling, borders, and interaction states.
// The VStack can be interactive if an action is defined and responds to hover, press,
// and focus states with visual feedback.
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
	transformation         Transformation

	invisible bool       // controls visibility
	font      proto.Font // font applied to text children
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string
	action             func()
	position           Position
	id                 string
	noClip             bool
	animation          Animation
	opacity            float64
	background         *Background
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

// Append adds additional child views to the VStack.
func (c TVStack) Append(children ...core.View) TVStack {
	c.children = append(c.children, children...)
	return c
}

// Gap sets the spacing between child views.
func (c TVStack) Gap(gap Length) TVStack {
	c.gap = gap.ora()
	return c
}

// Position sets the positioning of the VStack.
func (c TVStack) Position(position Position) TVStack {
	c.position = position
	return c
}

// StylePreset applies a predefined style preset.
func (c TVStack) StylePreset(preset StylePreset) TVStack {
	c.stylePreset = preset.ora()
	return c
}

// TextColor sets the default text color for the VStack.
func (c TVStack) TextColor(textColor Color) TVStack {
	c.textColor = textColor.ora()
	return c
}

// Opacity sets the visibility of this component. The range is [0..1] where 0 means fully transparent and 1 means
// fully visible. This also affects all contained children.
func (c TVStack) Opacity(opacity float64) TVStack {
	c.opacity = 1 - opacity
	return c
}

// BackgroundColor sets the background color.
func (c TVStack) BackgroundColor(backgroundColor Color) TVStack {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// HoveredBackgroundColor sets the background color when hovered.
func (c TVStack) HoveredBackgroundColor(backgroundColor Color) TVStack {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

// PressedBackgroundColor sets the background color when pressed.
func (c TVStack) PressedBackgroundColor(backgroundColor Color) TVStack {
	c.pressedBackgroundColor = backgroundColor.ora()
	return c
}

// FocusedBackgroundColor sets the background color when focused.
func (c TVStack) FocusedBackgroundColor(backgroundColor proto.Color) TVStack {
	c.focusedBackgroundColor = backgroundColor
	return c
}

// Action assigns an action handler, making the VStack interactive.
func (c TVStack) Action(f func()) TVStack {
	c.action = f
	return c
}

// NoClip disables clipping of child content when true.
func (c TVStack) NoClip(b bool) TVStack {
	c.noClip = b
	return c
}

// Alignment sets the alignment of child views within the column.
func (c TVStack) Alignment(alignment Alignment) TVStack {
	c.alignment = alignment.ora()
	return c
}

// Font sets the default font for text children.
func (c TVStack) Font(font Font) TVStack {
	c.font = font.ora()
	return c
}

// Frame sets the layout frame of the VStack.
func (c TVStack) Frame(f Frame) DecoredView {
	c.frame = f
	return c
}

func (c TVStack) With(fn func(stack TVStack) TVStack) TVStack {
	return fn(c)
}

// WithFrame modifies the current frame using the provided function.
func (c TVStack) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth sets the VStack to span 100% of the available width.
func (c TVStack) FullWidth() TVStack {
	c.frame.Width = "100%"
	return c
}

// Padding sets the inner padding of the VStack.
func (c TVStack) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TVStack) WithPadding(padding Padding) TVStack {
	c.padding = padding.ora()
	return c
}

// Border sets the default border styling.
func (c TVStack) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// HoveredBorder sets the border styling when hovered.
func (c TVStack) HoveredBorder(border Border) TVStack {
	c.hoveredBorder = border.ora()
	return c
}

// PressedBorder sets the border styling when pressed.
func (c TVStack) PressedBorder(border Border) TVStack {
	c.pressedBorder = border.ora()
	return c
}

// FocusedBorder sets the border styling when focused.
func (c TVStack) FocusedBorder(border Border) TVStack {
	c.focusedBorder = border.ora()
	return c
}

// Visible controls the visibility of the VStack.
func (c TVStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the accessibility label for screen readers.
func (c TVStack) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// ID assigns a unique identifier to the VStack.
func (c TVStack) ID(id string) TVStack {
	c.id = id
	return c
}

func (c TVStack) Animation(animation Animation) TVStack {
	c.animation = animation
	return c
}

func (c TVStack) Transformation(transformation Transformation) TVStack {
	c.transformation = transformation
	return c
}

func (c TVStack) Background(bg Background) TVStack {
	c.background = &bg
	return c
}

// Render builds and returns the protocol representation of the VStack.
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
		Animation:          proto.Animation(c.animation),
		Transformation:     c.transformation.ora(),

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
		Opacity:                clampOpacity(c.opacity),
		Background:             c.background.proto(),
	}
}

func clampOpacity(o float64) proto.Uint {
	var opacity proto.Uint
	if o != 0 {
		// we transmit it inverse, 0=fully visible and 1=fully transparent
		opacity = proto.Uint(o * 100)
		opacity = max(0, opacity)
		opacity = min(100, opacity)
	}

	return opacity
}

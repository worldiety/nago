// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"runtime/debug"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// THStack is a layout component(HStack).
// HStack is a horizontal layout container that arranges its child views in a row.
// It supports alignment, spacing, background styling, borders, and interaction states.
// The HStack is interactive if an action is defined and can respond to hover, press,
// and focus states with visual feedback.
type THStack struct {
	children               []core.View
	alignment              proto.Alignment
	backgroundColor        proto.Color
	hoveredBackgroundColor proto.Color
	pressedBackgroundColor proto.Color
	focusedBackgroundColor proto.Color
	frame                  Frame
	gap                    proto.Length
	padding                proto.Padding
	font                   proto.Font
	border                 proto.Border
	hoveredBorder          proto.Border
	focusedBorder          proto.Border
	pressedBorder          proto.Border
	accessibilityLabel     string
	invisible              bool
	action                 func()
	stylePreset            proto.StylePreset
	originTrace            string
	wrap                   bool
	disabled               bool
	position               Position
	id                     string
	noClip                 bool
	textColor              Color
}

// HStack is a container, in which the given children will be layout in a row according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func HStack(children ...core.View) THStack {
	c := THStack{
		children: children,
	}

	if core.Debug {
		c.originTrace = strings.Split(string(debug.Stack()), "\n")[6]
	}

	return c
}

// Append adds one or more child views to the horizontal stack.
func (c THStack) Append(children ...core.View) THStack {
	c.children = append(c.children, children...)
	return c
}

// Enabled has only an effect if StylePreset is applied, otherwise it is ignored.
func (c THStack) Enabled(enabled bool) THStack {
	c.disabled = !enabled
	return c
}

// Padding sets the inner spacing around the stack's children.
func (c THStack) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

// Gap sets the spacing between child views in the horizontal stack.
func (c THStack) Gap(gap Length) THStack {
	c.gap = gap.ora()
	return c
}

// Position sets the position of the horizontal stack within its parent layout.
func (c THStack) Position(position Position) THStack {
	c.position = position
	return c
}

// Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width,
// it must create multiple rows to place its children. Note, that the text layout behavior is unspecified
// (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own
// custom size, so this must not use a grid-like layouting.
func (c THStack) Wrap(wrap bool) THStack {
	c.wrap = wrap
	return c
}

// BackgroundColor sets the background color of the horizontal stack.
func (c THStack) BackgroundColor(backgroundColor Color) THStack {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// HoveredBackgroundColor sets the background color of the stack
// when the user hovers over it.
func (c THStack) HoveredBackgroundColor(backgroundColor Color) THStack {
	c.hoveredBackgroundColor = backgroundColor.ora()
	return c
}

// PressedBackgroundColor sets the background color of the stack
// when it is pressed or clicked.
func (c THStack) PressedBackgroundColor(backgroundColor Color) THStack {
	c.pressedBackgroundColor = backgroundColor.ora()
	return c
}

// FocusedBackgroundColor sets the background color of the stack
// when it is focused (e.g., via keyboard navigation).
func (c THStack) FocusedBackgroundColor(backgroundColor Color) THStack {
	c.focusedBackgroundColor = proto.Color(backgroundColor)
	return c
}

// Alignment sets how the stack's children are aligned vertically
// within the horizontal row.
func (c THStack) Alignment(alignment Alignment) THStack {
	c.alignment = alignment.ora()
	return c
}

// Frame sets the layout frame of the horizontal stack, including size and positioning.
func (c THStack) Frame(fr Frame) DecoredView {
	c.frame = fr
	return c
}

// WithFrame applies a transformation function to the stack's frame
// and returns the updated component.
func (c THStack) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// With applies a transformation function to the stack itself and returns the result.
// Useful for chaining configuration in a functional style.
func (c THStack) With(fn func(stack THStack) THStack) THStack {
	return fn(c)
}

// FullWidth sets the stack to span the full available width.
func (c THStack) FullWidth() THStack {
	c.frame.Width = "100%"
	return c
}

// Font sets the font style applied to text content inside the stack.
func (c THStack) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

// StylePreset applies a predefined style preset to the stack, controlling its appearance.
func (c THStack) StylePreset(preset StylePreset) THStack {
	c.stylePreset = preset.ora()
	return c
}

// Border sets the border styling of the stack.
func (c THStack) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// HoveredBorder sets the border styling when the stack is hovered.
func (c THStack) HoveredBorder(border Border) THStack {
	c.hoveredBorder = border.ora()
	return c
}

// PressedBorder sets the border styling when the stack is pressed or clicked.
func (c THStack) PressedBorder(border Border) THStack {
	c.pressedBorder = border.ora()
	return c
}

// FocusedBorder sets the border styling when the stack is focused.
func (c THStack) FocusedBorder(border Border) THStack {
	c.focusedBorder = border.ora()
	return c
}

// Visible controls the visibility of the stack; setting false hides it.
func (c THStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the label used by screen readers for accessibility.
func (c THStack) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// TextColor sets the color of text content inside the stack.
func (c THStack) TextColor(textColor Color) THStack {
	c.textColor = textColor
	return c
}

// Action sets the callback function to be invoked when the stack is clicked or tapped.
func (c THStack) Action(f func()) THStack {
	c.action = f
	return c
}

// ID assigns a unique identifier to the stack, useful for testing or referencing.
func (c THStack) ID(id string) THStack {
	c.id = id
	return c
}

// NoClip toggles whether the stack clips its children.
// By default, stacks clip their children; setting true disables clipping.
func (c THStack) NoClip(b bool) THStack {
	c.noClip = b
	return c
}

// Render builds and returns the protocol representation of the horizontal stack.
func (c THStack) Render(ctx core.RenderContext) core.RenderNode {
	ptr := ctx.MountCallback(c.action)
	if core.Debug {
		fmt.Printf("hstack got %d @%s\n", ptr, c.originTrace)
	}
	return &proto.HStack{
		Children:           renderComponents(ctx, c.children),
		Gap:                c.gap,
		Frame:              c.frame.ora(),
		Alignment:          c.alignment,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		Border:             c.border,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Font:               c.font,

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		Action:                 ptr,
		Wrap:                   proto.Bool(c.wrap),
		Disabled:               proto.Bool(c.disabled),
		Position:               c.position.ora(),
		TextColor:              proto.Color(c.textColor),

		StylePreset: c.stylePreset,
		Id:          proto.Str(c.id),
		NoClip:      proto.Bool(c.noClip),
	}
}

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

type StackLayout int

const (
	StackLayoutAuto StackLayout = iota
	StackLayoutVertical
	StackLayoutHorizontal
	StackLayoutWrap
)

const StackLayoutFlex = StackLayoutWrap

type THStack = TStack

type TVStack = TStack

// TStack is a layout component(Stack).
// It is responsive and can switch between [HStack] and [VStack] during rendering.
type TStack struct {
	children               []core.View
	alignment              Alignment
	backgroundColor        Color
	hoveredBackgroundColor Color
	pressedBackgroundColor Color
	focusedBackgroundColor Color
	frame                  Frame
	gap                    Length
	padding                Padding
	border                 Border
	hoveredBorder          Border
	focusedBorder          Border
	pressedBorder          Border
	layout                 StackLayout
	noClip                 bool
	font                   Font
	accessibilityLabel     string
	invisible              bool
	action                 func()
	stylePreset            StylePreset
	url                    core.URI
	target                 string
	wrap                   bool
	disabled               bool
	position               Position
	id                     string
	opacity                float64
	background             *Background
	textColor              Color
	animation              Animation
	transformation         Transformation
	originTrace            string
	adapt                  func(wnd core.Window, stack TStack) TStack
}

// Stack is a responsive variant which decides between [VStack] and [HStack].
func Stack(children ...core.View) TStack {
	return TStack{
		children: children,
	}
}

// HStack is a container, in which the given children will be layout in a row according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func HStack(children ...core.View) TStack {
	c := TStack{
		children: children,
		layout:   StackLayoutHorizontal,
	}

	if core.Debug {
		c.originTrace = strings.Split(string(debug.Stack()), "\n")[6]
	}

	return c
}

// VStack is a container, in which the given children will be layout in a column according to the applied
// alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding.
func VStack(children ...core.View) TStack {
	c := TStack{
		children: children,
		layout:   StackLayoutVertical,
	}

	if core.Debug {
		c.originTrace = strings.Split(string(debug.Stack()), "\n")[6]
	}

	return c
}

func (c TStack) NoClip(b bool) TStack {
	c.noClip = b
	return c
}

// Font sets the font style applied to text content inside the stack.
func (c TStack) Font(font Font) TStack {
	c.font = font
	return c
}

// AccessibilityLabel sets the label used by screen readers for accessibility.
func (c TStack) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// Visible controls the visibility of the stack; setting false hides it.
func (c TStack) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// Action sets the callback function to be invoked when the stack is clicked or tapped.
func (c TStack) Action(f func()) TStack {
	c.action = f
	return c
}

// StylePreset applies a predefined style preset to the stack, controlling its appearance.
func (c TStack) StylePreset(preset StylePreset) TStack {
	c.stylePreset = preset
	return c
}

// HRef sets the URL that the button navigates to when clicked if no action is specified.
// If both URL and Action are set, the URL takes precedence.
// This avoids another render cycle if the only goal is to navigate to a different page.
// It also avoids issues with browser which block async browser interactions like Safari.
// In fact, the [core.Navigation.Open] does not work properly on Safari.
// See also [TButton.Target].
func (c TStack) HRef(url core.URI) TStack {
	c.url = url
	return c
}

// Target sets the name of the browsing context, like _self, _blank, _ parent, _top.
func (c TStack) Target(target string) TStack {
	c.target = target
	return c
}

// Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width,
// it must create multiple rows to place its children. Note, that the text layout behavior is unspecified
// (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own
// custom size, so this must not use a grid-like layouting.
func (c TStack) Wrap(wrap bool) TStack {
	c.wrap = wrap
	return c
}

// Enabled has only an effect if StylePreset is applied, otherwise it is ignored.
func (c TStack) Enabled(enabled bool) TStack {
	c.disabled = !enabled
	return c
}

// Position sets the position of the horizontal stack within its parent layout.
func (c TStack) Position(position Position) TStack {
	c.position = position
	return c
}

// ID assigns a unique identifier to the stack, useful for testing or referencing.
func (c TStack) ID(id string) TStack {
	c.id = id
	return c
}

// Opacity sets the visibility of this component. The range is [0..1] where 0 means fully transparent and 1 means
// fully visible. This also affects all contained children.
func (c TStack) Opacity(opacity float64) TStack {
	c.opacity = 1 - opacity
	return c
}

func (c TStack) Background(bg Background) TStack {
	c.background = &bg
	return c
}

// TextColor sets the color of text content inside the stack.
func (c TStack) TextColor(textColor Color) TStack {
	c.textColor = textColor
	return c
}

func (c TStack) Animation(animation Animation) TStack {
	c.animation = animation
	return c
}

func (c TStack) Transformation(transformation Transformation) TStack {
	c.transformation = transformation
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

// HoveredBackgroundColor sets the background color of the stack
// when the user hovers over it.
func (c TStack) HoveredBackgroundColor(backgroundColor Color) TStack {
	c.hoveredBackgroundColor = backgroundColor
	return c
}

// PressedBackgroundColor sets the background color of the stack
// when it is pressed or clicked.
func (c TStack) PressedBackgroundColor(backgroundColor Color) TStack {
	c.pressedBackgroundColor = backgroundColor
	return c
}

// FocusedBackgroundColor sets the background color of the stack
// when it is focused (e.g., via keyboard navigation).
func (c TStack) FocusedBackgroundColor(backgroundColor Color) TStack {
	c.focusedBackgroundColor = backgroundColor
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

func (c TStack) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the stack's frame
// and returns the updated component.
func (c TStack) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// With applies a transformation function to the stack itself and returns the result.
// Useful for chaining configuration in a functional style.
func (c TStack) With(fn func(stack TStack) TStack) TStack {
	return fn(c)
}

func (c TStack) Gap(gap Length) TStack {
	c.gap = gap
	return c
}

func (c TStack) Padding(padding Padding) DecoredView {
	c.padding = padding
	return c
}

func (c TStack) WithPadding(padding Padding) TStack {
	c.padding = padding
	return c
}

func (c TStack) Border(border Border) DecoredView {
	c.border = border
	return c
}

// HoveredBorder sets the border styling when the stack is hovered.
func (c TStack) HoveredBorder(border Border) TStack {
	c.hoveredBorder = border
	return c
}

// PressedBorder sets the border styling when the stack is pressed or clicked.
func (c TStack) PressedBorder(border Border) TStack {
	c.pressedBorder = border
	return c
}

// FocusedBorder sets the border styling when the stack is focused.
func (c TStack) FocusedBorder(border Border) TStack {
	c.focusedBorder = border
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

	var orientation proto.Orientation
	if layout == StackLayoutVertical {
		orientation = proto.Vertical
	} else {
		orientation = proto.Horizontal
	}

	ptr := ctx.MountCallback(c.action)
	if core.Debug {
		fmt.Printf("stack got %d @%s\n", ptr, c.originTrace)
	}

	return &proto.Stack{
		Children:           renderComponents(ctx, c.children),
		Gap:                proto.Length(c.gap),
		Frame:              c.frame.ora(),
		Alignment:          proto.Alignment(c.alignment),
		BackgroundColor:    proto.Color(c.backgroundColor),
		Padding:            c.padding.ora(),
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Border:             c.border.ora(),
		Font:               c.font.ora(),
		Action:             ptr,
		BackgroundColorStates: proto.ColorStates{
			Hover:   proto.Color(c.hoveredBackgroundColor),
			Focus:   proto.Color(c.focusedBackgroundColor),
			Pressed: proto.Color(c.pressedBackgroundColor),
		},
		HoveredBorder:  c.hoveredBorder.ora(),
		PressedBorder:  c.pressedBorder.ora(),
		FocusedBorder:  c.focusedBorder.ora(),
		Wrap:           proto.Bool(c.wrap),
		StylePreset:    proto.StylePreset(c.stylePreset),
		Position:       c.position.ora(),
		Disabled:       proto.Bool(c.disabled),
		Invisible:      proto.Bool(c.invisible),
		Id:             proto.Str(c.id),
		TextColor:      proto.Color(c.textColor),
		NoClip:         proto.Bool(c.noClip),
		Animation:      proto.Animation(c.animation),
		Transformation: c.transformation.ora(),
		Opacity:        clampOpacity(c.opacity),
		Background:     c.background.proto(),
		Url:            proto.URI(c.url),
		Target:         proto.Str(c.target),
		Orientation:    orientation,
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

func (c TStack) FullWidth() TStack {
	c.frame.Width = Full
	return c
}

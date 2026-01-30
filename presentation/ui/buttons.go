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

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TButton is a basic component(Button).
// A basic clickable UI component used to trigger actions or events. There are three different kinds of Buttons:
// PrimaryButton, SecondaryButton & TertiaryButton.
type TButton struct {
	title              string
	accessibilityLabel string
	preIcon            proto.SVG
	postIcon           proto.SVG
	frame              Frame
	preset             ButtonStyle
	font               Font
	action             func()
	trace              string
	invisible          bool
	disabled           bool
	id                 string
}

// Button creates a new Button with the given style preset.
func Button(style ButtonStyle, action func()) TButton {
	return initButton(action, style)
}

// PrimaryButton uses an internal preset to represent a primary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func PrimaryButton(action func()) TButton {
	return initButton(action, StyleButtonPrimary)
}

// SecondaryButton uses an internal preset to represent a secondary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func SecondaryButton(action func()) TButton {
	return initButton(action, StyleButtonSecondary)
}

// TertiaryButton uses an internal preset to represent a tertiary button. See also FilledButton for a custom-colored
// Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
func TertiaryButton(action func()) TButton {
	return initButton(action, StyleButtonTertiary)
}

func initButton(action func(), preset StylePreset) TButton {
	btn := TButton{action: action, preset: preset}
	if core.Debug {
		btn.trace = string(debug.Stack())
	}
	return btn
}

// Enabled toggles whether the button is interactive.
// This has an effect only if a StylePreset is applied; otherwise it is ignored.
func (c TButton) Enabled(b bool) TButton {
	c.disabled = !b
	return c
}

// Title sets the text label displayed on the button.
func (c TButton) Title(text string) TButton {
	c.title = text
	return c
}

// Visible controls the visibility of the button; setting false hides it.
func (c TButton) Visible(b bool) TButton {
	c.invisible = !b
	return c
}

func (c TButton) Disabled(b bool) TButton {
	c.disabled = b
	return c
}

// Font sets the font style for the button's text label.
func (c TButton) Font(font Font) TButton {
	c.font = font
	return c
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
func (c TButton) AccessibilityLabel(label string) TButton {
	c.accessibilityLabel = label
	return c
}

// PreIcon sets the icon displayed before the text label.
func (c TButton) PreIcon(svg core.SVG) TButton {
	c.preIcon = proto.SVG(svg)
	return c
}

// PostIcon sets the icon displayed after the text label.
func (c TButton) PostIcon(svg core.SVG) TButton {
	c.postIcon = proto.SVG(svg)
	return c
}

// Preset applies a style preset to the button, controlling its appearance and behavior.
func (c TButton) Preset(preset StylePreset) TButton {
	c.preset = preset
	return c
}

// Frame sets the layout frame of the button, including size and positioning.
func (c TButton) Frame(frame Frame) TButton {
	c.frame = frame
	return c
}

// ID assigns a unique identifier to the button, useful for testing or referencing.
func (c TButton) ID(id string) TButton {
	c.id = id
	return c
}

// Render builds and returns the visual representation of the button.
func (c TButton) Render(context core.RenderContext) proto.Component {
	alabel := c.accessibilityLabel
	if alabel == "" {
		alabel = c.title
	}

	if core.Debug {
		if alabel == "" {
			panic(fmt.Errorf("the ora guidelines forbid buttons without accessibility label, allocated here: %s", c.trace))
		}
	}

	return HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", Text(c.title).Font(c.font)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Gap(L4).
		ID(c.id).
		Enabled(!c.disabled).
		Action(c.action).
		StylePreset(c.preset).
		Frame(c.frame).
		Visible(!c.invisible).
		AccessibilityLabel(alabel). // this is redundant and requires the text twice, however we are "just" an hstack
		Render(context)
}

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

// TFilledButton is a basic component (Filled Button).
// It is rendered with a solid background color, optional text and icons,
// and can execute an action when pressed. Styling includes colors,
// spacing, and interactive states like hover, press, and focus.
type TFilledButton struct {
	fillColor Color
	textColor Color
	title     string
	preIcon   proto.SVG
	postIcon  proto.SVG
	frame     Frame
	action    func()
}

// FilledButton creates a new filled button with the given background color
// and action callback.
func FilledButton(fillColor Color, action func()) TFilledButton {
	return TFilledButton{fillColor: fillColor, action: action}
}

// Title sets the text label displayed on the button.
func (c TFilledButton) Title(text string) TFilledButton {
	c.title = text
	return c
}

// PreIcon sets the icon displayed before the text label.
func (c TFilledButton) PreIcon(svg core.SVG) TFilledButton {
	c.preIcon = proto.SVG(svg)
	return c
}

// PostIcon sets the icon displayed after the text label.
func (c TFilledButton) PostIcon(svg core.SVG) TFilledButton {
	c.postIcon = proto.SVG(svg)
	return c
}

// TextColor sets the text color of the button label.
func (c TFilledButton) TextColor(color Color) TFilledButton {
	c.textColor = color
	return c
}

// Frame sets the layout frame of the button, including size and positioning.
func (c TFilledButton) Frame(frame Frame) TFilledButton {
	c.frame = frame
	return c
}

// Render builds and returns the visual representation of the filled button.
// It arranges optional icons and text inside a horizontal stack, applies
// interactive styles, and ensures consistent sizing for icon-only buttons.
func (c TFilledButton) Render(context core.RenderContext) proto.Component {
	decView := customButton(c.fillColor, HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", btnTitle(c.title, c.textColor)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Action(c.action))

	if (c.title == "" && len(c.preIcon) != 0) || (c.title == "" && len(c.postIcon) != 0) {
		decView = decView.Frame(Frame{Width: L40, Height: L40}).Padding(Padding{}.Horizontal("0dp"))
	}

	var zero Frame
	if c.frame != zero {
		decView = decView.Frame(c.frame)
	}

	return decView.Render(context)
}

// btnTitle creates a styled text element for the button label,
// using a medium-weight font and optional custom color.
func btnTitle(text string, color Color) TText {
	t := Text(text).Font(Font{Size: L14, Weight: 500})
	if color != "" {
		t = t.Color(color)
	}

	return t
}

// customButton applies default styling and interactive states to a button.
// It sets hover, pressed, and focused effects, as well as padding, background
// color, and a circular border to ensure consistent sizing.
func customButton(fillColor Color, hstack THStack) DecoredView {
	return hstack.
		HoveredBackgroundColor(fillColor.WithTransparency(25)).
		PressedBackgroundColor(fillColor.WithTransparency(35)).
		PressedBorder(Border{}.Circle().Color("#00000000").Width(L2)).
		FocusedBorder(Border{}.Circle().Color("#ffffff").Width(L2)).
		Gap(L4).
		BackgroundColor(fillColor).
		Frame(Frame{Height: "2.375rem"}).
		Padding(Padding{}.Horizontal("1.125rem")).
		// add invisible default border, to avoid dimension changes,
		// note, that we need to fix that with frame and padding above
		Border(Border{}.Circle().Color("#00000000").Width(L2))
}

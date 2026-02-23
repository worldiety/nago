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

// TDialog is a overlay component (Dialog).
// It represents a modal or popup window with optional title, body, and footer sections.
// The dialog can be aligned, styled with padding and frames, and may support or disable
// box-based layouts for its content.
type TDialog struct {
	uri              core.URI     // unique identifier for the dialog instance
	dlg              proto.VStack // internal protocol structure for layout
	preBody          core.View    // optional view rendered before the body
	body             core.View    // main content of the dialog
	footer           core.View    // optional footer content (e.g., buttons)
	title            core.View    // optional title displayed at the top
	titleX           core.View    // optional title element aligned differently (e.g., actions)
	alignment        Alignment    // alignment of the dialog content
	modalPadding     Padding      // padding around dialog content
	frame            Frame        // layout frame for sizing and positioning
	disableBoxLayout bool         // when true, disables box layout handling
}

// Dialog creates a new dialog with the given body content and default frame settings.
// By default, the dialog has a fixed width of 400dp, max width set to full, and max height
// limited to the viewport minus 12rem.
func Dialog(body core.View) TDialog {
	return TDialog{
		frame: Frame{Width: L400, MaxWidth: Full, MaxHeight: "calc(100dvh - 12rem)"},
		body:  body,
	}
}

// Title sets the title view of the dialog, displayed at the top.
func (c TDialog) Title(title core.View) TDialog {
	c.title = title
	return c
}

// PreBody sets an optional view that will be rendered before the main body.
func (c TDialog) PreBody(v core.View) TDialog {
	c.preBody = v
	return c
}

// TitleX sets an additional title view, often used for actions or
// secondary title elements aligned differently from the main title.
func (c TDialog) TitleX(x core.View) TDialog {
	c.titleX = x
	return c
}

// Footer sets the footer view of the dialog, typically used for action buttons.
func (c TDialog) Footer(footer core.View) TDialog {
	c.footer = footer
	return c
}

// Alignment sets the alignment of the dialog content.
func (c TDialog) Alignment(alignment Alignment) TDialog {
	c.alignment = alignment
	return c
}

// ModalPadding sets the padding around the dialog content.
func (c TDialog) ModalPadding(padding Padding) TDialog {
	c.modalPadding = padding
	return c
}

// Frame sets the layout frame of the dialog, including size and positioning.
func (c TDialog) Frame(frame Frame) TDialog {
	c.frame = frame
	return c
}

// WithFrame applies a transformation function to the dialog's frame
// and returns the updated component.
func (c TDialog) WithFrame(fn func(Frame) Frame) TDialog {
	c.frame = fn(c.frame)
	return c
}

// DisableBoxLayout disables box layout handling for the dialog when set to true.
func (c TDialog) DisableBoxLayout(b bool) TDialog {
	c.disableBoxLayout = b
	return c
}

// Render builds and returns the protocol representation of the dialog.
func (c TDialog) Render(ctx core.RenderContext) proto.Component {
	colors := core.Colors[Colors](ctx.Window())

	stack := VStack(
		If(c.title != nil, HStack(c.title, Spacer(), c.titleX).Alignment(Leading).BackgroundColor(ColorCardTop).Frame(Frame{}.FullWidth()).Padding(Padding{Left: L20, Right: L20, Top: L12, Bottom: L12})),
		VStack(
			c.preBody,
			c.body,
		).
			Frame(Frame{Width: Full, MaxHeight: c.frame.MaxHeight}). // TODO strange mix of layout...
			Padding(Padding{Left: L20, Top: L16, Right: L20, Bottom: L20}),

		// footer must spawn without padding
		If(c.footer != nil, HStack(c.footer).
			Alignment(Trailing).
			BackgroundColor(ColorCardFooter).
			Padding(Padding{}.Horizontal(L16)).
			Frame(Frame{Width: Full, Height: L60, MinHeight: L60}),
		),
	).
		BackgroundColor(ColorCardBody).
		Border(Border{}.Radius(L20).Elevate(4)).
		Frame(Frame{Height: c.frame.Height, MinWidth: c.frame.MinWidth, Width: c.frame.Width, MaxWidth: "calc(100% - 2rem)"}) // TODO ... this looks wrong

	if c.disableBoxLayout {
		return stack.Render(ctx)
	}

	dlg := BoxAlign(c.alignment, stack).
		DisableOutsidePointerEvents(true).
		BackgroundColor(colors.M5.WithTransparency(40)).Padding(c.modalPadding)

	return dlg.Render(ctx)
}

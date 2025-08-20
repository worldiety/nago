// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package esignature

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TSignature represents a composite component (Signature).
// This component arranges a top label, a main signature body view,
// and a bottom label inside a frame.
type TSignature struct {
	frame      ui.Frame  // frame defining size and layout
	topView    core.View // optional text displayed above the signature
	sigView    core.View // main signature view or component
	bottomView core.View // optional text displayed below the signature
}

// Signature creates a new, empty TSignature.
func Signature() TSignature {
	return TSignature{}
}

// Frame sets the frame of the signature container.
func (c TSignature) Frame(frame ui.Frame) TSignature {
	c.frame = frame
	return c
}

// TopText sets the text displayed above the signature body.
func (c TSignature) TopText(text string) TSignature {
	c.topView = ui.Text(text).Font(ui.BodySmall).Padding(ui.Padding{Top: "-0.5rem"})
	return c
}

// BottomText sets the text displayed below the signature body.
func (c TSignature) BottomText(text string) TSignature {
	c.bottomView = ui.Text(text).Font(ui.BodySmall).Padding(ui.Padding{Bottom: "-0.75rem"})
	return c
}

// Body sets the main body view of the signature.
func (c TSignature) Body(v core.View) TSignature {
	c.sigView = v
	return c
}

// Render builds and returns the RenderNode for the TSignature.
// It arranges the signature in three parts:
// - Top: a rounded corner line with optional top text
// - Body: a vertical line with the main signature view
// - Bottom: a rounded corner line with optional bottom text
// The layout is wrapped in a vertical stack, framed, padded, and aligned to the leading edge.
func (c TSignature) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		// top left line
		ui.HStack(
			ui.VStack().Border(ui.Border{
				TopLeftRadius: ui.L16,
				LeftWidth:     ui.L1,
				TopWidth:      ui.L1,
			}.Color(ui.ColorInputBorder)).Frame(ui.Frame{MinWidth: ui.L48, MinHeight: ui.L16}),
			c.topView,
		).Alignment(ui.Top).NoClip(true).Gap(ui.L4),

		// left vertical line + content
		ui.HStack(
			ui.VStack().Border(ui.Border{
				LeftWidth: ui.L1,
			}.Color(ui.ColorInputBorder)).Frame(ui.Frame{MinWidth: ui.L1, MinHeight: ui.L32}),

			ui.VStack(
				c.sigView,
			).Alignment(ui.Leading),
		).Gap(ui.L16).Alignment(ui.Stretch),

		// bottom left line
		ui.HStack(
			ui.VStack().Border(ui.Border{
				BottomLeftRadius: ui.L16,
				LeftWidth:        ui.L1,
				BottomWidth:      ui.L1,
			}.Color(ui.ColorInputBorder)).Frame(ui.Frame{MinWidth: ui.L48, MinHeight: ui.L16}),
			c.bottomView,
		).NoClip(true).Gap(ui.L4),
	).Alignment(ui.Leading).
		Frame(c.frame).
		Padding(ui.Padding{Top: ui.L16, Bottom: ui.L16}).
		Render(ctx)
}

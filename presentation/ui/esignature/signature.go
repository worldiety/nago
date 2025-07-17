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

type TSignature struct {
	frame      ui.Frame
	topView    core.View
	sigView    core.View
	bottomView core.View
}

func Signature() TSignature {
	return TSignature{}
}

func (c TSignature) Frame(frame ui.Frame) TSignature {
	c.frame = frame
	return c
}

func (c TSignature) TopText(text string) TSignature {
	c.topView = ui.Text(text).Font(ui.BodySmall).Padding(ui.Padding{Top: "-0.5rem"})
	return c
}

func (c TSignature) BottomText(text string) TSignature {
	c.bottomView = ui.Text(text).Font(ui.BodySmall).Padding(ui.Padding{Bottom: "-0.75rem"})
	return c
}

func (c TSignature) Body(v core.View) TSignature {
	c.sigView = v
	return c
}

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

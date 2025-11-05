// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package accordion

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type TAccordion struct {
	header core.View
	body   core.View
	frame  ui.Frame
	open   *core.State[bool]
}

func Accordion(header, body core.View, open *core.State[bool]) TAccordion {
	return TAccordion{
		header: header,
		body:   body,
		open:   open,
	}
}

func (t TAccordion) Frame(frame ui.Frame) TAccordion {
	t.frame = frame
	return t
}

func (t TAccordion) FullWidth() TAccordion {
	t.frame.Width = ui.Full
	return t
}

func (t TAccordion) Render(ctx core.RenderContext) core.RenderNode {
	// TODO we should create this as proto primitive to avoid render-roundtrips and allow better SEO support, e.g. also with SSR

	ico := icons.ChevronDown
	var tf ui.Transformation
	if t.open.Get() {
		//	ico = icons.ChevronUp
		tf.RotateZ = 180
	}

	return ui.VStack(
		ui.HStack(
			t.header,
			ui.Spacer(),
			ui.VStack(ui.ImageIcon(ico)).Animation(ui.AnimateTransition).Transformation(tf),
		).Action(func() {
			t.open.Set(!t.open.Get())
		}).FullWidth(),
		ui.If(t.open.Get(), t.body),
	).Frame(t.frame).Render(ctx)
}

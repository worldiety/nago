// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package accordion

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type TAccordion struct {
	header     core.View
	body       core.View
	frame      ui.Frame
	value      bool
	inputValue *core.State[bool]
}

// Accordion initializes a new TAccordion element
// The function takes a header view, body view and open state
func Accordion(header, body core.View, open *core.State[bool]) TAccordion {
	return TAccordion{
		header:     header,
		body:       body,
		value:      open.Get(),
		inputValue: open,
	}
}

// Frame sets the accordions frame to control the accordions bounds
func (t TAccordion) Frame(frame ui.Frame) TAccordion {
	t.frame = frame
	return t
}

// InputValue binds the accordion to an external boolean state,
// allowing it to be controlled from outside the component.
func (t TAccordion) InputValue(input *core.State[bool]) TAccordion {
	t.inputValue = input
	return t
}

// FullWidth sets the accordion's frame to full width
func (t TAccordion) FullWidth() TAccordion {
	t.frame.Width = ui.Full
	return t
}

func (t TAccordion) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Accordion{
		Header:  t.header.Render(ctx),
		Content: t.body.Render(ctx),
		Frame: proto.Frame{
			MinWidth:  proto.Length(t.frame.MinWidth),
			MaxWidth:  proto.Length(t.frame.MaxWidth),
			MinHeight: proto.Length(t.frame.MinHeight),
			MaxHeight: proto.Length(t.frame.MaxHeight),
			Width:     proto.Length(t.frame.Width),
			Height:    proto.Length(t.frame.Height),
		},
		InputValue: t.inputValue.Ptr(),
		Value:      proto.Bool(t.value),
	}
}

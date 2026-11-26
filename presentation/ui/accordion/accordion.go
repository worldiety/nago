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
	small      bool
	separator  bool
	value      bool
	inputValue *core.State[bool]
}

// Accordion initializes a new TAccordion element
// The function takes a header view, body view and an optional open state. If open is nil, the open state
// is only kept in the frontend, so toggling the accordion does not require a server roundtrip. Use
// [TAccordion.Open] to define the initial state in that case.
func Accordion(header, body core.View, open *core.State[bool]) TAccordion {
	t := TAccordion{
		header:     header,
		body:       body,
		inputValue: open,
		separator:  true,
	}

	if open != nil {
		t.value = open.Get()
	}

	return t
}

// Open sets whether the accordion is open (expanded) or closed (collapsed). Without a bound state, this
// is only the initial state, which can be toggled afterwards by the user within the frontend.
func (t TAccordion) Open(isOpen bool) TAccordion {
	t.value = isOpen
	return t
}

// Frame sets the accordions frame to control the accordions bounds
func (t TAccordion) Frame(frame ui.Frame) TAccordion {
	t.frame = frame
	return t
}

// Small sets the small flag to enable a smaller visual appearance of the accordion
func (t TAccordion) Small() TAccordion {
	t.small = true
	return t
}

// HideSeparator sets a flag to hide the bottom border (separator) of the accordion
func (t TAccordion) HideSeparator() TAccordion {
	t.separator = false
	return t
}

// InputValue binds the accordion to an external boolean state,
// allowing it to be controlled from outside the component. Passing nil keeps the open state
// solely in the frontend.
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
		Small:      proto.Bool(t.small),
		Separator:  proto.Bool(t.separator),
		InputValue: t.inputValue.Ptr(),
		Value:      proto.Bool(t.value),
	}
}

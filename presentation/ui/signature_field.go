// Copyright (c) 2026 worldiety GmbH
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

// TSignatureField is an advanced component to input a signature.
// This component provides a field with optional label, support, and error text.
type TSignatureField struct {
	label          string                 // label displayed above the field
	value          Signature              // static value (used if InputValue is not set)
	inputValue     *core.State[Signature] // bound state for controlled input
	supportingText string                 // helper text shown below the field
	errorText      string                 // error message shown below the field
	disabled       bool                   // disables user interaction
	frame          Frame                  // layout constraints
	style          TextFieldStyle         // visual style of the field (outlined, filled, etc.)
}

type Signature struct {
	SVG string // SVG representation of the signature
}

// SignatureField creates a new signature field with the given label and initial value.
// By default, it is single-line and uncontrolled until InputValue is set.
func SignatureField(label string, state *core.State[Signature]) TSignatureField {
	c := TSignatureField{
		label:      label,
		inputValue: state,
	}

	if state != nil {
		c.value = state.Get()
	}

	return c
}

// Label sets the label text of the field.
func (c TSignatureField) Label(label string) TSignatureField {
	c.label = label
	return c
}

// Value sets a static Signature value for the field. This is only used if no InputValue state is provided.
func (c TSignatureField) Value(value Signature) TSignatureField {
	c.value = value
	return c
}

// InputValue binds the field to a reactive state.
// This enables controlled input behavior where the state is updated as the user types.
func (c TSignatureField) InputValue(input *core.State[Signature]) TSignatureField {
	c.inputValue = input
	return c
}

// SupportingText sets helper text for the field.
// This text is displayed below the input and is typically used to provide hints or guidance.
func (c TSignatureField) SupportingText(text string) TSignatureField {
	c.supportingText = text
	return c
}

// ErrorText sets an error message for the field.
// When provided, this text is shown below the input in place of supporting text,
// usually styled to indicate an error state.
func (c TSignatureField) ErrorText(text string) TSignatureField {
	c.errorText = text
	return c
}

// Disabled disables or enables the field.
// When disabled, the user cannot interact with the field.
func (c TSignatureField) Disabled(disabled bool) TSignatureField {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame of the field (size, width, height, etc.).
func (c TSignatureField) Frame(frame Frame) TSignatureField {
	c.frame = frame
	return c
}

// Style sets the visual style of the field (e.g., outlined, filled, etc.).
func (c TSignatureField) Style(style TextFieldStyle) TSignatureField {
	c.style = style
	return c
}

// Render builds and returns the protocol representation of the field.
func (c TSignatureField) Render(_ core.RenderContext) core.RenderNode {
	value := c.value
	if c.inputValue != nil {
		value = c.inputValue.Get()
	}

	return &proto.SignatureField{
		Value: proto.Signature{
			SVG: proto.Str(value.SVG),
		},
		Frame:          c.frame.ora(),
		InputValue:     c.inputValue.Ptr(),
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Disabled:       proto.Bool(c.disabled),
	}
}

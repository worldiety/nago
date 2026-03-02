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

type SelectOption struct {
	Value    string
	Label    string
	Disabled bool
}

// TSelect is a basic component (Select).
// It allows users to select one option of a given set of options,
type TSelect struct {
	label          string               // label displayed above or inside the field
	value          string               // current value
	inputValue     *core.State[string]  // optional external binding for controlled state
	options        []SelectOption       // list of options to choose from
	supportingText string               // helper text shown below the field
	errorText      string               // error message shown below the field
	disabled       bool                 // when true, interaction is disabled
	leading        core.View            // optional leading element (e.g., icon)
	style          proto.TextFieldStyle // visual style of the field (outlined, filled, etc.)
	frame          Frame                // layout constraints
	id             string               // unique identifier for the select
}

// Select represents a user interface element which lets the user select one option from a list.
func Select(options []SelectOption, value string) TSelect {
	c := TSelect{
		options: options,
		value:   value,
	}

	return c
}

// Label sets the label displayed above or inside the select.
func (c TSelect) Label(label string) TSelect {
	c.label = label
	return c
}

// Value sets the initial value of the select.
func (c TSelect) Value(value string) TSelect {
	c.value = value
	return c
}

// InputValue binds the select to an external value state,
// allowing it to be controlled from outside the component.
func (c TSelect) InputValue(input *core.State[string]) TSelect {
	c.inputValue = input
	return c
}

// Options sets the list of options available for selection.
func (c TSelect) Options(options []SelectOption) TSelect {
	c.options = options
	return c
}

// SupportingText sets the supporting text displayed below the select.
func (c TSelect) SupportingText(text string) TSelect {
	c.supportingText = text
	return c
}

// ErrorText sets the error text displayed below the select.
func (c TSelect) ErrorText(text string) TSelect {
	c.errorText = text
	return c
}

// Disabled enables or disables user interaction with the select.
func (c TSelect) Disabled(disabled bool) TSelect {
	c.disabled = disabled
	return c
}

// Leading sets a leading view for the select.
// This view is displayed at the start of the select, e.g., an icon.
func (c TSelect) Leading(v core.View) TSelect {
	c.leading = v
	return c
}

// Style sets the visual style of the select.
func (c TSelect) Style(s TextFieldStyle) TSelect {
	c.style = s.ora()
	return c
}

// Frame sets the layout frame of the field (size, width, height, etc.).
func (c TSelect) Frame(frame Frame) TSelect {
	c.frame = frame
	return c
}

// ID assigns a unique identifier to the select, useful for testing or referencing.
func (c TSelect) ID(id string) TSelect {
	c.id = id
	return c
}

// Render builds and returns the protocol representation of the select.
func (c TSelect) Render(ctx core.RenderContext) core.RenderNode {
	options := make([]proto.SelectOption, 0, len(c.options))
	for _, option := range c.options {
		options = append(options, proto.SelectOption{
			Value:    proto.Str(option.Value),
			Disabled: proto.Bool(option.Disabled),
			Label:    proto.Str(option.Label),
		})
	}

	return &proto.Select{
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Value:          proto.Str(c.value),
		Frame:          c.frame.ora(),
		InputValue:     c.inputValue.Ptr(),
		Style:          c.style,
		Leading:        render(ctx, c.leading),
		Disabled:       proto.Bool(c.disabled),
		Id:             proto.Str(c.id),
		Options:        options,
	}
}

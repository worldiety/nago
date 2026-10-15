// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TToggleField is a toggle with a label, see ToggleField.
type TToggleField struct{ ref js.Ref }

func (TToggleField) isView() {}

// ToggleField aggregates a toggle together with form field typical labels, hints and error texts.
//
//gostrict:host
func ToggleField(label string, value bool) TToggleField {
	return TToggleField{ref: wrap(nui.ToggleField(label, value))}
}

// Label sets the label, which is shown next to the toggle.
//
//gostrict:host
func (c TToggleField) Label(label string) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Label(label))}
}

// Value sets the initial state of the toggle, true for on and false for off.
//
//gostrict:host
func (c TToggleField) Value(value bool) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Value(value))}
}

// ID sets the id of the toggle field, e.g. for testing or referencing.
//
//gostrict:host
func (c TToggleField) ID(id string) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).ID(id))}
}

// Padding sets the inner padding of the toggle field.
//
//gostrict:host
func (c TToggleField) Padding(padding Padding) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Padding(padding.nago()).(nui.TToggleField))}
}

// Frame sets the layout frame of the toggle field.
//
//gostrict:host
func (c TToggleField) Frame(frame Frame) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Frame(frame.nago()).(nui.TToggleField))}
}

// WithFrame changes the layout frame of the toggle field with fn, which gets the current one.
//
//gostrict:host
func (c TToggleField) WithFrame(fn func(Frame) Frame) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TToggleField))}
}

// Border sets the border styling of the toggle field.
//
//gostrict:host
func (c TToggleField) Border(border Border) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Border(border.nago()).(nui.TToggleField))}
}

// Visible shows or hides the toggle field.
//
//gostrict:host
func (c TToggleField) Visible(visible bool) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Visible(visible).(nui.TToggleField))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TToggleField) AccessibilityLabel(label string) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).AccessibilityLabel(label).(nui.TToggleField))}
}

// InputValue binds the toggle field to a state of the window for two-way data binding.
//
//gostrict:host
func (c TToggleField) InputValue(inputValue BoolState) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).InputValue(inputValue.nago()))}
}

// SupportingText sets the helper text, which is shown below the field.
//
//gostrict:host
func (c TToggleField) SupportingText(text string) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).SupportingText(text))}
}

// ErrorText sets the validation or error message, which is shown below the field.
//
//gostrict:host
func (c TToggleField) ErrorText(text string) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).ErrorText(text))}
}

// Disabled enables or disables the user interaction with the toggle field.
//
//gostrict:host
func (c TToggleField) Disabled(disabled bool) TToggleField {
	return TToggleField{ref: wrap(get[nui.TToggleField](c.ref).Disabled(disabled))}
}

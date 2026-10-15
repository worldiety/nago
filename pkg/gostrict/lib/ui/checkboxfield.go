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

// TCheckboxField is a checkbox with a label, see CheckboxField.
type TCheckboxField struct{ ref js.Ref }

func (TCheckboxField) isView() {}

// CheckboxField aggregates a checkbox together with form field typical labels, hints and error texts.
//
//gostrict:host
func CheckboxField(label string, value bool) TCheckboxField {
	return TCheckboxField{ref: wrap(nui.CheckboxField(label, value))}
}

// Enabled sets whether the checkbox field is interactive. It is equivalent to Disabled(!b).
//
//gostrict:host
func (c TCheckboxField) Enabled(b bool) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Enabled(b))}
}

// Disabled enables or disables the user interaction with the checkbox field.
//
//gostrict:host
func (c TCheckboxField) Disabled(b bool) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Disabled(b))}
}

// Padding sets the inner spacing around the checkbox field.
//
//gostrict:host
func (c TCheckboxField) Padding(padding Padding) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Padding(padding.nago()).(nui.TCheckboxField))}
}

// Frame sets the layout frame of the checkbox field, including size and positioning.
//
//gostrict:host
func (c TCheckboxField) Frame(frame Frame) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Frame(frame.nago()).(nui.TCheckboxField))}
}

// WithFrame changes the layout frame of the checkbox field with fn, which gets the current one.
//
//gostrict:host
func (c TCheckboxField) WithFrame(fn func(Frame) Frame) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TCheckboxField))}
}

// Border sets the border styling of the checkbox field.
//
//gostrict:host
func (c TCheckboxField) Border(border Border) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Border(border.nago()).(nui.TCheckboxField))}
}

// Visible shows or hides the checkbox field.
//
//gostrict:host
func (c TCheckboxField) Visible(visible bool) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Visible(visible).(nui.TCheckboxField))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TCheckboxField) AccessibilityLabel(label string) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).AccessibilityLabel(label).(nui.TCheckboxField))}
}

// InputValue binds the checkbox field to a state of the window.
//
//gostrict:host
func (c TCheckboxField) InputValue(inputValue BoolState) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).InputValue(inputValue.nago()))}
}

// SupportingText sets the helper text, which is shown below the label.
//
//gostrict:host
func (c TCheckboxField) SupportingText(text string) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).SupportingText(text))}
}

// ErrorText sets the validation or error message, which is shown below the field.
//
//gostrict:host
func (c TCheckboxField) ErrorText(text string) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).ErrorText(text))}
}

// ID sets the id of the checkbox field, e.g. for testing or referencing.
//
//gostrict:host
func (c TCheckboxField) ID(id string) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).ID(id))}
}

// Name sets the name of the checkbox field, e.g. for autocomplete.
//
//gostrict:host
func (c TCheckboxField) Name(name string) TCheckboxField {
	return TCheckboxField{ref: wrap(get[nui.TCheckboxField](c.ref).Name(name))}
}

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
	"time"

	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TPasswordField is a password field, see PasswordField. It is a secure input for passwords or secrets, which shows
// the input only as masked. It supports validation feedback, debouncing, autocomplete control and styling.
type TPasswordField struct{ ref js.Ref }

func (TPasswordField) isView() {}

// PasswordField returns a password field with the label and the initial value. Do not reuse the value after the user
// has finished the secret step. A secret must not be shown again, nor stored in plain text or as a simple hash.
//
//gostrict:host
func PasswordField(label string, value string) TPasswordField {
	return TPasswordField{ref: wrap(nui.PasswordField(label, value))}
}

// Value sets the initial value of the password field. Be careful with that, as you should not reuse the value after
// the user has finished the secret step.
//
//gostrict:host
func (c TPasswordField) Value(value string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Value(value))}
}

// AutoComplete enables or disables the browser autocomplete for the field.
//
//gostrict:host
func (c TPasswordField) AutoComplete(autoComplete bool) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).AutoComplete(autoComplete))}
}

// Padding sets the padding around the field. It is a placeholder, which does not change the field yet.
//
//gostrict:host
func (c TPasswordField) Padding(padding Padding) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Padding(padding.nago()).(nui.TPasswordField))}
}

// Border sets the border styling of the field. It is a placeholder, which does not change the field yet.
//
//gostrict:host
func (c TPasswordField) Border(border Border) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Border(border.nago()).(nui.TPasswordField))}
}

// AccessibilityLabel sets an accessibility label for screen readers. It is a placeholder, which does not change the
// field yet.
//
//gostrict:host
func (c TPasswordField) AccessibilityLabel(label string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).AccessibilityLabel(label).(nui.TPasswordField))}
}

// SupportingText sets the helper text displayed below the field.
//
//gostrict:host
func (c TPasswordField) SupportingText(text string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).SupportingText(text))}
}

// ErrorText sets the error message displayed when a validation fails.
//
//gostrict:host
func (c TPasswordField) ErrorText(text string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).ErrorText(text))}
}

// Style sets the visual style of the field. Without a style, TextFieldOutlined is applied.
//
//gostrict:host
func (c TPasswordField) Style(s TextFieldStyle) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Style(nui.TextFieldStyle(s)))}
}

// DebounceTime sets the time in milliseconds, which the field waits after the user typed, before it reports the
// text. By default, this is 500 milliseconds. See also Debounce.
//
//gostrict:host
func (c TPasswordField) DebounceTime(millis int32) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).DebounceTime(time.Duration(millis) * time.Millisecond))}
}

// Debounce enables or disables the debouncing, which is enabled by default. See also DebounceTime.
//
//gostrict:host
func (c TPasswordField) Debounce(enabled bool) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Debounce(enabled))}
}

// InputValue binds the password field to the state, which keeps the secret between two renders.
//
//gostrict:host
func (c TPasswordField) InputValue(input StringState) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).InputValue(input.nago()))}
}

// Disabled enables or disables the field.
//
//gostrict:host
func (c TPasswordField) Disabled(disabled bool) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Disabled(disabled))}
}

// Frame sets the layout frame of the field.
//
//gostrict:host
func (c TPasswordField) Frame(frame Frame) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Frame(frame.nago()).(nui.TPasswordField))}
}

// WithFrame changes the layout frame of the field with fn, which gets the current one.
//
//gostrict:host
func (c TPasswordField) WithFrame(fn func(Frame) Frame) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TPasswordField))}
}

// Lines sets the number of lines. Zero lines is a single-line field. Otherwise, a text area is created.
//
//gostrict:host
func (c TPasswordField) Lines(lines int32) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Lines(int(lines)))}
}

// Visible shows or hides the field.
//
//gostrict:host
func (c TPasswordField) Visible(visible bool) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Visible(visible).(nui.TPasswordField))}
}

// ID sets the unique identifier of the field.
//
//gostrict:host
func (c TPasswordField) ID(id string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).ID(id))}
}

// KeydownEnter calls fn, when the user presses the enter key.
//
//gostrict:host
func (c TPasswordField) KeydownEnter(fn func()) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).KeydownEnter(fn))}
}

// Autocomplete sets the autocomplete tags of the input.
//
//gostrict:host
func (c TPasswordField) Autocomplete(tags string) TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).Autocomplete(tags))}
}

// FullWidth expands the field to the full available width.
//
//gostrict:host
func (c TPasswordField) FullWidth() TPasswordField {
	return TPasswordField{ref: wrap(get[nui.TPasswordField](c.ref).FullWidth())}
}

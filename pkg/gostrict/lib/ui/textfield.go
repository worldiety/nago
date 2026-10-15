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

// TextFieldStyle is the visual style of a text field.
type TextFieldStyle uint32

const (
	// TextFieldReduced has no outlines and thus less visual disruption in larger forms.
	TextFieldReduced TextFieldStyle = 1

	// TextFieldOutlined is fine for smaller forms and helps to identify where to put text in the form.
	TextFieldOutlined TextFieldStyle = 0

	// TextFieldBasic removes as much as decorations as possible. There may be limitations based on the platform.
	// Note, that an implementation is allowed to ignore leading, trailing, supporting and errorText for this mode.
	// It may serve as a building-block for custom fields.
	TextFieldBasic TextFieldStyle = 2

	// TextFieldFilled has more subtle outlines and a filled background color.
	TextFieldFilled TextFieldStyle = 3
)

// TTextField is a text field, see TextField. It provides a text input with optional supporting and error text,
// leading and trailing views, debounce settings, styling and keyboard options.
type TTextField struct{ ref js.Ref }

func (TTextField) isView() {}

// TextField returns a text field with the label and the initial value. It is single-line and uncontrolled, until
// InputValue is set.
//
//gostrict:host
func TextField(label string, value string) TTextField {
	return TTextField{ref: wrap(nui.TextField(label, value))}
}

// IntField returns a text field for integers, which uses the keyboard for integers. The state keeps the value, see
// IntState. Invalid input is ignored.
//
//gostrict:host
func IntField(label string, value int32, state IntState) TTextField {
	return TTextField{ref: wrap(nui.IntField(label, int64(value), state.nago()))}
}

// FloatField returns a text field for decimal numbers, which uses the keyboard for decimal numbers. The state keeps
// the value, see FloatState. Invalid input is ignored.
//
//gostrict:host
func FloatField(label string, value float64, state FloatState) TTextField {
	return TTextField{ref: wrap(nui.FloatField(label, value, state.nago()))}
}

// Border is a placeholder, which does not change the field yet.
//
//gostrict:host
func (c TTextField) Border(border Border) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Border(border.nago()).(nui.TTextField))}
}

// AccessibilityLabel is a placeholder, which does not change the field yet.
//
//gostrict:host
func (c TTextField) AccessibilityLabel(label string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).AccessibilityLabel(label).(nui.TTextField))}
}

// Value sets a static text value for the field. It is used, if the constructor TextField cannot be used.
//
//gostrict:host
func (c TTextField) Value(value string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Value(value))}
}

// SupportingText sets the helper text below the field.
//
//gostrict:host
func (c TTextField) SupportingText(text string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).SupportingText(text))}
}

// ErrorText sets the error message below the field, which replaces the supporting text.
//
//gostrict:host
func (c TTextField) ErrorText(text string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).ErrorText(text))}
}

// Leading sets a view at the start of the field, e.g. an icon.
//
//gostrict:host
func (c TTextField) Leading(v View) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Leading(viewOf(v)))}
}

// Trailing sets a view at the end of the field, e.g. a clear button or an icon.
//
//gostrict:host
func (c TTextField) Trailing(v View) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Trailing(viewOf(v)))}
}

// TextAlignment aligns the text in the field.
//
//gostrict:host
func (c TTextField) TextAlignment(v TextAlignment) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).TextAlignment(nui.TextAlignment(v)))}
}

// Style sets the visual style of the field. Without a style, TextFieldOutlined is applied.
//
//gostrict:host
func (c TTextField) Style(s TextFieldStyle) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Style(nui.TextFieldStyle(s)))}
}

// DebounceTime sets the time in milliseconds, which the field waits after the user typed, before it reports the
// text. By default, this is 500 milliseconds. See also Debounce.
//
//gostrict:host
func (c TTextField) DebounceTime(millis int32) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).DebounceTime(time.Duration(millis) * time.Millisecond))}
}

// Debounce enables or disables the debouncing, which is enabled by default. See also DebounceTime.
//
//gostrict:host
func (c TTextField) Debounce(enabled bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Debounce(enabled))}
}

// InputValue binds the field to the state, which keeps the text between two renders.
//
//gostrict:host
func (c TTextField) InputValue(input StringState) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).InputValue(input.nago()))}
}

// Disabled disables or enables the field.
//
//gostrict:host
func (c TTextField) Disabled(disabled bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Disabled(disabled))}
}

// Frame sets the layout frame of the field, e.g. its size.
//
//gostrict:host
func (c TTextField) Frame(frame Frame) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Frame(frame.nago()).(nui.TTextField))}
}

// WithFrame changes the layout frame of the field with fn, which gets the current one.
//
//gostrict:host
func (c TTextField) WithFrame(fn func(Frame) Frame) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TTextField))}
}

// FullWidth expands the field to the full available width.
//
//gostrict:host
func (c TTextField) FullWidth() TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).FullWidth())}
}

// KeydownEnter calls fn, when the user presses the enter key. It only works for single-line fields, see Lines.
//
//gostrict:host
func (c TTextField) KeydownEnter(fn func()) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).KeydownEnter(fn))}
}

// Lines sets the number of lines. Zero lines is a single-line field. Otherwise, a text area is created.
//
//gostrict:host
func (c TTextField) Lines(lines int32) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Lines(int(lines)))}
}

// Visible shows or hides the field, which still takes its place in the layout when hidden.
//
//gostrict:host
func (c TTextField) Visible(visible bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Visible(visible).(nui.TTextField))}
}

// KeyboardOptions sets the options of the keyboard, e.g. the capitalization.
//
//gostrict:host
func (c TTextField) KeyboardOptions(options TKeyboardOptions) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).KeyboardOptions(get[nui.TKeyboardOptions](options.ref)))}
}

// KeyboardType sets the type of the keyboard, e.g. for numbers or e-mail addresses.
//
//gostrict:host
func (c TTextField) KeyboardType(keyboardType KeyboardType) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).KeyboardType(nui.KeyboardType(keyboardType)))}
}

// ID sets the unique identifier of the field, e.g. for tests or accessibility.
//
//gostrict:host
func (c TTextField) ID(id string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).ID(id))}
}

// ShowZero shows the character 0 for empty or zero values in number fields.
//
//gostrict:host
func (c TTextField) ShowZero(showZero bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).ShowZero(showZero))}
}

// Step sets the step size, which increases or decreases a number in number fields.
//
//gostrict:host
func (c TTextField) Step(step int32) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Step(int(step)))}
}

// Max sets the maximum value of number fields.
//
//gostrict:host
func (c TTextField) Max(max float64) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Max(max))}
}

// Min sets the minimum value of number fields.
//
//gostrict:host
func (c TTextField) Min(min float64) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Min(min))}
}

// Autocomplete sets the autocomplete tags of the input.
//
//gostrict:host
func (c TTextField) Autocomplete(tags string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Autocomplete(tags))}
}

// Optional sets whether the field is optional.
//
//gostrict:host
func (c TTextField) Optional(optional bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Optional(optional))}
}

// ClearButton shows a button, which clears the field, if it has no trailing view and is not empty.
//
//gostrict:host
func (c TTextField) ClearButton(clearButton bool) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).ClearButton(clearButton))}
}

// Placeholder sets the placeholder text, which is shown while the field is empty.
//
//gostrict:host
func (c TTextField) Placeholder(placeholder string) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Placeholder(placeholder))}
}

// Padding sets the padding of the field, which overwrites the defaults.
//
//gostrict:host
func (c TTextField) Padding(padding Padding) TTextField {
	return TTextField{ref: wrap(get[nui.TTextField](c.ref).Padding(padding.nago()).(nui.TTextField))}
}

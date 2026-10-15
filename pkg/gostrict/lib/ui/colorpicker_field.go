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
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

// TColorpickerPickerField is a color field, see ColorpickerPickerField.
type TColorpickerPickerField struct{ ref js.Ref }

func (TColorpickerPickerField) isView() {}

// ColorpickerPickerField creates a field showing the color, which opens a color picker on click, like
// colorpicker.PickerField of nago.
//
//gostrict:host
func ColorpickerPickerField(label string, value Color) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(colorpicker.PickerField(label, nui.Color(value)))}
}

// Mode selects the color space of the picker.
//
//gostrict:host
func (c TColorpickerPickerField) Mode(mode ColorpickerMode) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).Mode(colorpicker.Mode(mode)))}
}

// DisableAlpha hides the alpha bar of the picker, so that only opaque colors can be picked.
//
//gostrict:host
func (c TColorpickerPickerField) DisableAlpha(disable bool) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).DisableAlpha(disable))}
}

// SupportingText sets the supporting text.
//
//gostrict:host
func (c TColorpickerPickerField) SupportingText(text string) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).SupportingText(text))}
}

// ErrorText sets the error text.
//
//gostrict:host
func (c TColorpickerPickerField) ErrorText(text string) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).ErrorText(text))}
}

// Disabled disables or enables the field.
//
//gostrict:host
func (c TColorpickerPickerField) Disabled(disabled bool) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).Disabled(disabled))}
}

// Visible shows or hides the field.
//
//gostrict:host
func (c TColorpickerPickerField) Visible(visible bool) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).Visible(visible))}
}

// Frame sets the size.
//
//gostrict:host
func (c TColorpickerPickerField) Frame(frame Frame) TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).Frame(frame.nago()))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TColorpickerPickerField) FullWidth() TColorpickerPickerField {
	return TColorpickerPickerField{ref: wrap(get[colorpicker.TPickerField](c.ref).FullWidth())}
}

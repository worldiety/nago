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

// ColorpickerMode selects the color space of a color picker, like colorpicker.Mode of nago.
type ColorpickerMode int32

const (
	// ColorpickerModeHSV uses hue, saturation and value, the default.
	ColorpickerModeHSV ColorpickerMode = 0
	// ColorpickerModeHCT uses hue, chroma and tone.
	ColorpickerModeHCT ColorpickerMode = 1
)

// TColorpickerPicker is a color picker, see ColorpickerPicker.
type TColorpickerPicker struct{ ref js.Ref }

func (TColorpickerPicker) isView() {}

// ColorpickerPicker creates a color picker showing the value, like colorpicker.Picker of nago.
//
//gostrict:host
func ColorpickerPicker(label string, value Color) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(colorpicker.Picker(label, nui.Color(value)))}
}

// Mode selects the color space of the area and the hue bar.
//
//gostrict:host
func (c TColorpickerPicker) Mode(mode ColorpickerMode) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).Mode(colorpicker.Mode(mode)))}
}

// DisableAlpha hides the alpha bar, so that only opaque colors can be picked.
//
//gostrict:host
func (c TColorpickerPicker) DisableAlpha(disable bool) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).DisableAlpha(disable))}
}

// SupportingText sets the supporting text.
//
//gostrict:host
func (c TColorpickerPicker) SupportingText(text string) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).SupportingText(text))}
}

// ErrorText sets the error text.
//
//gostrict:host
func (c TColorpickerPicker) ErrorText(text string) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).ErrorText(text))}
}

// Disabled disables or enables the color picker.
//
//gostrict:host
func (c TColorpickerPicker) Disabled(disabled bool) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).Disabled(disabled))}
}

// Visible shows or hides the color picker.
//
//gostrict:host
func (c TColorpickerPicker) Visible(visible bool) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).Visible(visible))}
}

// Frame sets the size.
//
//gostrict:host
func (c TColorpickerPicker) Frame(frame Frame) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).Frame(frame.nago()))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TColorpickerPicker) FullWidth() TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).FullWidth())}
}

// ID sets an identifier, which must be unique within the view tree.
//
//gostrict:host
func (c TColorpickerPicker) ID(id string) TColorpickerPicker {
	return TColorpickerPicker{ref: wrap(get[colorpicker.TPicker](c.ref).ID(id))}
}

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

// TTimeField is a text field for a time of day, which the user types, see TimeField.
type TTimeField struct{ ref js.Ref }

func (TTimeField) isView() {}

// TimeField returns a text field for a time of day, which the user types, e.g. 8:30, 8.30, 0830 or 8, like
// ui.TimeField of nago. The state holds the time as HH:MM from 00:00 to 23:59, "" for an empty field or, if the text
// is no time, the text as typed, and the field then shows an error text; check the state before saving it. The field
// rewrites what has been typed only at the complete forms H:MM, HH:MM and four digits, e.g. 0830 to 08:30. The
// observers of the state are called, when its value changes, and a state, which has been set from outside, is shown
// as it is.
//
//gostrict:host
func TimeField(label string, state StringState) TTimeField {
	return TTimeField{ref: wrap(nui.TimeField(label, state.nago()))}
}

// SupportingText sets the helper text below the field.
//
//gostrict:host
func (c TTimeField) SupportingText(text string) TTimeField {
	return TTimeField{ref: wrap(get[nui.TTimeField](c.ref).SupportingText(text))}
}

// ErrorText sets the error message below the field. A text, which is not empty, replaces the error text of the
// field, an empty text does not.
//
//gostrict:host
func (c TTimeField) ErrorText(text string) TTimeField {
	return TTimeField{ref: wrap(get[nui.TTimeField](c.ref).ErrorText(text))}
}

// Disabled enables or disables the field.
//
//gostrict:host
func (c TTimeField) Disabled(disabled bool) TTimeField {
	return TTimeField{ref: wrap(get[nui.TTimeField](c.ref).Disabled(disabled))}
}

// Frame sets the size.
//
//gostrict:host
func (c TTimeField) Frame(frame Frame) TTimeField {
	return TTimeField{ref: wrap(get[nui.TTimeField](c.ref).Frame(frame.nago()))}
}

// Placeholder sets the text, which an empty field shows.
//
//gostrict:host
func (c TTimeField) Placeholder(placeholder string) TTimeField {
	return TTimeField{ref: wrap(get[nui.TTimeField](c.ref).Placeholder(placeholder))}
}

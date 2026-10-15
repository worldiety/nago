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

// TRadioButtonField is a radio button with a label, see RadioButtonField.
type TRadioButtonField struct{ ref js.Ref }

func (TRadioButtonField) isView() {}

// RadioButtonField combines a radio button with a label. The radio button selects the state at index of the group.
//
//gostrict:host
func RadioButtonField(label string, stateGroup RadioStateGroup, index int32) TRadioButtonField {
	g := b2GroupOf(stateGroup)
	return TRadioButtonField{ref: wrap(nui.RadioButtonField(label, &g, int(index)))}
}

// Label sets the label of the radio button field.
//
//gostrict:host
func (c TRadioButtonField) Label(label string) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).Label(label))}
}

// InputChecked binds the radio button to a state of the window, which keeps the selected state in sync.
//
//gostrict:host
func (c TRadioButtonField) InputChecked(input BoolState) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).InputChecked(input.nago()))}
}

// ID sets the id of the radio button field, e.g. for testing or referencing.
//
//gostrict:host
func (c TRadioButtonField) ID(id string) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).ID(id))}
}

// Disabled disables the radio button field, which prevents the user interaction.
//
//gostrict:host
func (c TRadioButtonField) Disabled(disabled bool) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).Disabled(disabled))}
}

// Visible shows or hides the radio button field.
//
//gostrict:host
func (c TRadioButtonField) Visible(visible bool) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).Visible(visible))}
}

// Name sets the name of the radio button field, e.g. for autocomplete.
//
//gostrict:host
func (c TRadioButtonField) Name(name string) TRadioButtonField {
	return TRadioButtonField{ref: wrap(get[nui.TRadioButtonField](c.ref).Name(name))}
}

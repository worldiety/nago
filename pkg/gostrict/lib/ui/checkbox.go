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

// TCheckbox is a checkbox, see Checkbox.
type TCheckbox struct{ ref js.Ref }

func (TCheckbox) isView() {}

// Checkbox represents a user interface element, which the user clicks or taps to check or uncheck. Use it for
// controls, which do not cause an immediate effect. See also Toggle.
//
//gostrict:host
func Checkbox(checked bool) TCheckbox {
	return TCheckbox{ref: wrap(nui.Checkbox(checked))}
}

// InputChecked binds the checkbox to a state of the window.
//
// Deprecated: use InputValue.
//
//gostrict:host
func (c TCheckbox) InputChecked(input BoolState) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).InputChecked(input.nago()))}
}

// InputValue binds the checkbox to a state of the window, which controls the checked state from outside.
//
//gostrict:host
func (c TCheckbox) InputValue(input BoolState) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).InputValue(input.nago()))}
}

// Disabled enables or disables the user interaction with the checkbox.
//
//gostrict:host
func (c TCheckbox) Disabled(disabled bool) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).Disabled(disabled))}
}

// Visible shows or hides the checkbox.
//
//gostrict:host
func (c TCheckbox) Visible(visible bool) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).Visible(visible))}
}

// ID sets the id of the checkbox, e.g. for testing or referencing.
//
//gostrict:host
func (c TCheckbox) ID(id string) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).ID(id))}
}

// Name sets the name of the checkbox.
//
//gostrict:host
func (c TCheckbox) Name(name string) TCheckbox {
	return TCheckbox{ref: wrap(get[nui.TCheckbox](c.ref).Name(name))}
}

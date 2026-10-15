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

// TToggle is a switch without a label, see Toggle.
type TToggle struct{ ref js.Ref }

func (TToggle) isView() {}

// Toggle is a switch-like control (on/off) without a label. Use it for immediate activation or deactivation of
// features. In contrast to that, use a Checkbox for form things without an immediate effect.
//
//gostrict:host
func Toggle(checked bool) TToggle {
	return TToggle{ref: wrap(nui.Toggle(checked))}
}

// ID sets the id of the toggle.
//
//gostrict:host
func (c TToggle) ID(id string) TToggle {
	return TToggle{ref: wrap(get[nui.TToggle](c.ref).ID(id))}
}

// InputChecked binds the toggle to a state of the window for two-way data binding.
//
//gostrict:host
func (c TToggle) InputChecked(input BoolState) TToggle {
	return TToggle{ref: wrap(get[nui.TToggle](c.ref).InputChecked(input.nago()))}
}

// Disabled enables or disables the interaction with the toggle.
//
//gostrict:host
func (c TToggle) Disabled(disabled bool) TToggle {
	return TToggle{ref: wrap(get[nui.TToggle](c.ref).Disabled(disabled))}
}

// Visible shows or hides the toggle.
//
//gostrict:host
func (c TToggle) Visible(visible bool) TToggle {
	return TToggle{ref: wrap(get[nui.TToggle](c.ref).Visible(visible))}
}

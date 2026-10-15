// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dropdown"
)

// DropdownOption is an option of a dropdown, like dropdown.Option of nago with a string as its value.
type DropdownOption struct {
	Value string
	Label string
	// Description is only visible in the styled dropdown, see TDropdown.StyledDropdown.
	Description string
	Disabled    bool
}

// TDropdown lets the user choose one of the options, like dropdown.TDropdown of nago.
type TDropdown struct{ ref js.Ref }

func (TDropdown) isView() {}

// Dropdown returns a dropdown with the options, whose value is selected, like dropdown.Dropdown of nago, which takes
// the options before the value.
//
//gostrict:host
func Dropdown(label string, value string, options ...DropdownOption) TDropdown {
	return TDropdown{ref: wrap(dropdown.Dropdown(label, dropdownOptionsOf(options), value))}
}

// Options sets the options.
//
//gostrict:host
func (c TDropdown) Options(options ...DropdownOption) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Options(dropdownOptionsOf(options)))}
}

// Label sets the label.
//
//gostrict:host
func (c TDropdown) Label(label string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Label(label))}
}

// Value sets the selected value.
//
//gostrict:host
func (c TDropdown) Value(value string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Value(value))}
}

// InputValue binds the selected value to the state.
//
//gostrict:host
func (c TDropdown) InputValue(input StringState) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).InputValue(input.nago()))}
}

// SupportingText sets the text below the dropdown.
//
//gostrict:host
func (c TDropdown) SupportingText(text string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).SupportingText(text))}
}

// ErrorText sets the error below the dropdown.
//
//gostrict:host
func (c TDropdown) ErrorText(text string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).ErrorText(text))}
}

// Disabled disables the dropdown.
//
//gostrict:host
func (c TDropdown) Disabled(disabled bool) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Disabled(disabled))}
}

// Leading shows the view before the value.
//
//gostrict:host
func (c TDropdown) Leading(v View) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Leading(viewOf(v)))}
}

// Style sets the style of the field.
//
//gostrict:host
func (c TDropdown) Style(s TextFieldStyle) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Style(nui.TextFieldStyle(s)))}
}

// Frame sets the size.
//
//gostrict:host
func (c TDropdown) Frame(frame Frame) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Frame(frame.nago()))}
}

// ID sets the id.
//
//gostrict:host
func (c TDropdown) ID(id string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).ID(id))}
}

// Autocomplete sets the autocomplete tags of the browser.
//
//gostrict:host
func (c TDropdown) Autocomplete(tags string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Autocomplete(tags))}
}

// StyledDropdown shows the styled dropdown of nago with descriptions instead of the one of the browser.
//
//gostrict:host
func (c TDropdown) StyledDropdown(b bool) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).StyledDropdown(b))}
}

// Searchable lets the user search the options.
//
//gostrict:host
func (c TDropdown) Searchable(b bool) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Searchable(b))}
}

// DropdownInfo shows the info in the open dropdown.
//
//gostrict:host
func (c TDropdown) DropdownInfo(info string) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).DropdownInfo(info))}
}

// Optional marks the field as optional.
//
//gostrict:host
func (c TDropdown) Optional(optional bool) TDropdown {
	return TDropdown{ref: wrap(get[dropdown.TDropdown[string]](c.ref).Optional(optional))}
}

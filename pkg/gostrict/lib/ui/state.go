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
	"go.wdy.de/nago/presentation/core"
)

// StringState is a text of the window, which it keeps between two renders, e.g. the value of a text field, see
// TTextField.InputValue.
type StringState struct{ ref js.Ref }

// BoolState is a boolean of the window, e.g. of a checkbox.
type BoolState struct{ ref js.Ref }

// StringStateOf returns the state with the id of the window, which is empty at first.
//
//gostrict:host
func StringStateOf(wnd Window, id string) StringState {
	return StringState{ref: wrap(core.StateOf[string](wnd.nago(), id))}
}

// Get returns the value.
//
//gostrict:host
func (s StringState) Get() string {
	return s.nago().Get()
}

// Set sets the value and renders the window again.
//
//gostrict:host
func (s StringState) Set(v string) {
	s.nago().Set(v)
}

// Init sets the value of fn, if the state is new.
//
//gostrict:host
func (s StringState) Init(fn func() string) StringState {
	return StringState{ref: wrap(s.nago().Init(fn))}
}

// Observe calls fn, when the user changes the value.
//
//gostrict:host
func (s StringState) Observe(fn func(v string)) StringState {
	return StringState{ref: wrap(s.nago().Observe(fn))}
}

// BoolStateOf returns the state with the id of the window, which is false at first.
//
//gostrict:host
func BoolStateOf(wnd Window, id string) BoolState {
	return BoolState{ref: wrap(core.StateOf[bool](wnd.nago(), id))}
}

// Get returns the value.
//
//gostrict:host
func (s BoolState) Get() bool {
	return s.nago().Get()
}

// Set sets the value and renders the window again.
//
//gostrict:host
func (s BoolState) Set(v bool) {
	s.nago().Set(v)
}

// Init sets the value of fn, if the state is new.
//
//gostrict:host
func (s BoolState) Init(fn func() bool) BoolState {
	return BoolState{ref: wrap(s.nago().Init(fn))}
}

// Observe calls fn, when the user changes the value.
//
//gostrict:host
func (s BoolState) Observe(fn func(v bool)) BoolState {
	return BoolState{ref: wrap(s.nago().Observe(fn))}
}

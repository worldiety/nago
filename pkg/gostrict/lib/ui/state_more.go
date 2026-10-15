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

// IntState is an integer of the window, e.g. a counter, which it keeps between two renders. Its value is an int32.
type IntState struct{ ref js.Ref }

// FloatState is a decimal number of the window, which it keeps between two renders.
type FloatState struct{ ref js.Ref }

// IntStateOf returns the state with the id of the window, which is 0 at first.
//
//gostrict:host
func IntStateOf(wnd Window, id string) IntState {
	return IntState{ref: wrap(core.StateOf[int64](wnd.nago(), id))}
}

// Get returns the value.
//
//gostrict:host
func (s IntState) Get() int32 {
	return int32(s.nago().Get())
}

// Set sets the value and renders the window again.
//
//gostrict:host
func (s IntState) Set(v int32) {
	s.nago().Set(int64(v))
}

// Init sets the value of fn, if the state is new.
//
//gostrict:host
func (s IntState) Init(fn func() int32) IntState {
	return IntState{ref: wrap(s.nago().Init(func() int64 { return int64(fn()) }))}
}

// Observe calls fn, when the user changes the value.
//
//gostrict:host
func (s IntState) Observe(fn func(v int32)) IntState {
	return IntState{ref: wrap(s.nago().Observe(func(v int64) { fn(int32(v)) }))}
}

// FloatStateOf returns the state with the id of the window, which is 0 at first.
//
//gostrict:host
func FloatStateOf(wnd Window, id string) FloatState {
	return FloatState{ref: wrap(core.StateOf[float64](wnd.nago(), id))}
}

// Get returns the value.
//
//gostrict:host
func (s FloatState) Get() float64 {
	return s.nago().Get()
}

// Set sets the value and renders the window again.
//
//gostrict:host
func (s FloatState) Set(v float64) {
	s.nago().Set(v)
}

// Init sets the value of fn, if the state is new.
//
//gostrict:host
func (s FloatState) Init(fn func() float64) FloatState {
	return FloatState{ref: wrap(s.nago().Init(fn))}
}

// Observe calls fn, when the user changes the value.
//
//gostrict:host
func (s FloatState) Observe(fn func(v float64)) FloatState {
	return FloatState{ref: wrap(s.nago().Observe(fn))}
}

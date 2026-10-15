// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program_test

import (
	"reflect"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/presentation/core"
)

// TestInputMirrorsCore keeps the input events of the programs identical to those of nago, because the frontend
// fills both from the same listeners.
func TestInputMirrorsCore(t *testing.T) {
	types := map[input.Type]core.InputEventType{
		input.PointerUp:     core.InputEventPointerUp,
		input.PointerDown:   core.InputEventPointerDown,
		input.PointerMove:   core.InputEventPointerMove,
		input.PointerCancel: core.InputEventPointerCancel,
		input.KeyDown:       core.InputEventKeyDown,
		input.KeyUp:         core.InputEventKeyUp,
		input.Invalidate:    core.InputEventInvalidate,
		input.MouseDown:     core.InputEventMouseDown,
		input.MouseEnter:    core.InputEventMouseEnter,
		input.MouseLeave:    core.InputEventMouseLeave,
		input.MouseOut:      core.InputEventMouseOut,
		input.MouseMove:     core.InputEventMouseMove,
		input.MouseOver:     core.InputEventMouseOver,
		input.MouseUp:       core.InputEventMouseUp,
		input.MouseWheel:    core.InputEventMouseWheel,
		input.PointerLeave:  core.InputEventPointerLeave,
		input.PointerOut:    core.InputEventPointerOut,
		input.PointerOver:   core.InputEventPointerOver,
		input.Blur:          core.InputEventBlur,
	}
	for a, b := range types {
		if uint64(a) != uint64(b) {
			t.Errorf("input type %d differs from core %d", a, b)
		}
	}

	in, co := reflect.TypeFor[input.Event](), reflect.TypeFor[core.InputEvent]()
	if in.NumField() != co.NumField() {
		t.Fatalf("input.Event has %d fields, core.InputEvent %d", in.NumField(), co.NumField())
	}
	for i := range in.NumField() {
		if in.Field(i).Name != co.Field(i).Name {
			t.Errorf("field %d: input.Event.%s, core.InputEvent.%s", i, in.Field(i).Name, co.Field(i).Name)
		}
	}
}

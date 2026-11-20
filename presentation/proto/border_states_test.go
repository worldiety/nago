// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package proto

import (
	"bytes"
	"reflect"
	"testing"
)

// TestBorderStatesFocusWithin: the state FocusWithin (field 5) of BorderStates survives Marshal and Unmarshal inside a
// stack, alone and beside the other states, and it is what makes a border of only this state non-zero, so the stack
// writes its border at all. Counter rows: without FocusWithin the decoded state stays empty, and the encoding of the
// same stack with and without FocusWithin differs, so the field is on the wire and not only in the struct.
func TestBorderStatesFocusWithin(t *testing.T) {
	border := func(c Color) Border {
		return Border{TopWidth: "1px", RightWidth: "1px", BottomWidth: "1px", LeftWidth: "1px",
			TopColor: c, RightColor: c, BottomColor: c, LeftColor: c,
			TopLeftRadius: "1rem", TopRightRadius: "1rem", BottomLeftRadius: "1rem", BottomRightRadius: "1rem"}
	}
	rows := []struct {
		name   string
		states BorderStates
		zero   bool
	}{
		{name: "FocusWithin only", states: BorderStates{FocusWithin: border("I0")}},
		{name: "Initial and FocusWithin", states: BorderStates{Initial: border("M5"), FocusWithin: border("I0")}},
		{name: "all five states", states: BorderStates{Initial: border("M5"), Hovered: border("M6"), Active: border("M7"), Focused: border("I1"), FocusWithin: border("I0")}},
		{name: "counter row: Initial and Focused without FocusWithin", states: BorderStates{Initial: border("M5"), Focused: border("I0")}},
		{name: "counter row: no state", states: BorderStates{}, zero: true},
	}

	encode := func(t *testing.T, v Writeable) []byte {
		t.Helper()
		var buf bytes.Buffer
		if err := Marshal(NewBinaryWriter(&buf), v); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if got := r.states.IsZero(); got != r.zero {
				t.Errorf("IsZero = %v, want %v", got, r.zero)
			}
			t0 := &RootViewInvalidated{RID: 1, Root: &Stack{Gap: "1rem", Border: r.states}}
			obj, err := Unmarshal(NewBinaryReader(bytes.NewBuffer(encode(t, t0))))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(obj, t0) {
				t.Errorf("expected: %#v, actual: %#v", t0, obj)
			}
			got := obj.(*RootViewInvalidated).Root.(*Stack).Border.FocusWithin
			if got != r.states.FocusWithin {
				t.Errorf("FocusWithin decoded as %+v, want %+v", got, r.states.FocusWithin)
			}
		})
	}

	t.Run("counter row: the field is encoded", func(t *testing.T) {
		with := &RootViewInvalidated{RID: 1, Root: &Stack{Border: BorderStates{Initial: border("M5"), FocusWithin: border("I0")}}}
		without := &RootViewInvalidated{RID: 1, Root: &Stack{Border: BorderStates{Initial: border("M5")}}}
		if a, b := encode(t, with), encode(t, without); len(a) <= len(b) {
			t.Errorf("with FocusWithin %d bytes, without %d bytes", len(a), len(b))
		}
	})
}

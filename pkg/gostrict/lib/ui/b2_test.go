// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"reflect"
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// TestViewsB2 compares the component of each constructor and builder of batch b2 with the component of nago. The
// batch has no enums, which the test could compare.
func TestViewsB2(t *testing.T) {
	views := []struct {
		name string
		view View
		want any
	}{
		{"Checkbox", Checkbox(true), nui.TCheckbox{}},
		{"Checkbox.InputValue", Checkbox(true).InputValue(BoolState{}), nui.TCheckbox{}},
		{"CheckboxField", CheckboxField("a", true), nui.TCheckboxField{}},
		{"CheckboxField.Padding", CheckboxField("a", true).Padding(Padding{}), nui.TCheckboxField{}},
		{"Toggle", Toggle(true), nui.TToggle{}},
		{"Toggle.InputChecked", Toggle(true).InputChecked(BoolState{}), nui.TToggle{}},
		{"ToggleField", ToggleField("a", true), nui.TToggleField{}},
		{"ToggleField.Padding", ToggleField("a", true).Padding(Padding{}), nui.TToggleField{}},
		{"RadioButton", RadioButton(true), nui.TRadioButton{}},
		{"RadioButton.Value", RadioButton(true).Value(false), nui.TRadioButton{}},
		{"RadioButtonField", RadioButtonField("a", RadioStateGroup{}, 0), nui.TRadioButtonField{}},
		{"RadioButtonField.Label", RadioButtonField("a", RadioStateGroup{}, 0).Label("b"), nui.TRadioButtonField{}},
	}
	for _, v := range views {
		got := Core(v.view)
		if reflect.TypeOf(got) != reflect.TypeOf(v.want) {
			t.Errorf("%s is %T, want %T", v.name, got, v.want)
		}
	}
}

// TestRadioStateGroupB2 checks, that the zero handle of a group behaves like the empty group of nago, which has no
// states and does not panic.
func TestRadioStateGroupB2(t *testing.T) {
	var s RadioStateGroup
	if got := s.SelectedIndex(); got != -1 {
		t.Errorf("SelectedIndex = %d, want -1", got)
	}
	s.SetSelectedIndex(1)
	s.InitIndex(1)
	s.Observe(func(int32) {})
	s.Notify()
}

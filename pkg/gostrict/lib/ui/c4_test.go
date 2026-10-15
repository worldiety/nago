// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/presentation/ui/stepper"
	"go.wdy.de/nago/presentation/ui/switcher"
)

// TestEnumsC4 compares the constants of the enums of batch c4 with the constants of nago.
func TestEnumsC4(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"ColorpickerModeHSV", int64(ColorpickerModeHSV), int64(colorpicker.ModeHSV)},
		{"ColorpickerModeHCT", int64(ColorpickerModeHCT), int64(colorpicker.ModeHCT)},
		{"StepperLayoutAuto", int64(StepperLayoutAuto), int64(stepper.StepperLayoutAuto)},
		{"StepperLayoutHorizontal", int64(StepperLayoutHorizontal), int64(stepper.StepperLayoutHorizontal)},
		{"StepperLayoutVertical", int64(StepperLayoutVertical), int64(stepper.StepperLayoutVertical)},
		{"StepperLayoutSimple", int64(StepperLayoutSimple), int64(stepper.StepperLayoutSimple)},
		{"StepperLayoutSimpleList", int64(StepperLayoutSimpleList), int64(stepper.StepperLayoutSimpleList)},
		{"SwitcherLayoutAuto", int64(SwitcherLayoutAuto), int64(switcher.SwitcherLayoutAuto)},
		{"SwitcherLayoutVertical", int64(SwitcherLayoutVertical), int64(switcher.SwitcherLayoutVertical)},
		{"SwitcherLayoutHorizontal", int64(SwitcherLayoutHorizontal), int64(switcher.SwitcherLayoutHorizontal)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/timepicker"
)

// TestEnumsC3 compares the constants of the enums of batch c3 with the constants of nago.
func TestEnumsC3(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"TimepickerClockFormat", int64(TimepickerClockFormat), int64(timepicker.ClockFormat)},
		{"TimepickerDecomposedFormat", int64(TimepickerDecomposedFormat), int64(timepicker.DecomposedFormat)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s: got %d, want %d", e.name, e.got, e.want)
		}
	}
}

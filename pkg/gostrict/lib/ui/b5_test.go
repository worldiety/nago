// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// TestEnumsB5 compares the constants of the enums of batch b5 with the constants of nago.
func TestEnumsB5(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"ScaffoldAlignmentTop", int64(ScaffoldAlignmentTop), int64(nui.ScaffoldAlignmentTop)},
		{"ScaffoldAlignmentLeading", int64(ScaffoldAlignmentLeading), int64(nui.ScaffoldAlignmentLeading)},
		{"SizeClassSmall", int64(SizeClassSmall), int64(core.SizeClassSmall)},
		{"SizeClassMedium", int64(SizeClassMedium), int64(core.SizeClassMedium)},
		{"SizeClassLarge", int64(SizeClassLarge), int64(core.SizeClassLarge)},
		{"SizeClassXL", int64(SizeClassXL), int64(core.SizeClassXL)},
		{"SizeClass2XL", int64(SizeClass2XL), int64(core.SizeClass2XL)},
		{"CountDownStyleClock", int64(CountDownStyleClock), int64(nui.CountDownStyleClock)},
		{"CountDownStyleProgress", int64(CountDownStyleProgress), int64(nui.CountDownStyleProgress)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}

// TestBreakpointB5 compares the default breakpoint of a scaffold with the one of nago.
func TestBreakpointB5(t *testing.T) {
	if int(DefaultScaffoldBreakpoint) != nui.DefaultScaffoldBreakpoint {
		t.Errorf("DefaultScaffoldBreakpoint = %d, want %d", DefaultScaffoldBreakpoint, nui.DefaultScaffoldBreakpoint)
	}
}

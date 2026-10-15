// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// TestEnumsB3 compares the constants of the enums of batch b3 with the constants of nago.
func TestEnumsB3(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"FitAuto", int64(FitAuto), int64(nui.FitAuto)},
		{"FitFill", int64(FitFill), int64(nui.FitFill)},
		{"FitCover", int64(FitCover), int64(nui.FitCover)},
		{"FitContain", int64(FitContain), int64(nui.FitContain)},
		{"FitNone", int64(FitNone), int64(nui.FitNone)},
		{"FitRepeat", int64(FitRepeat), int64(nui.FitRepeat)},
		{"ScrollViewAxisVertical", int64(ScrollViewAxisVertical), int64(nui.ScrollViewAxisVertical)},
		{"ScrollViewAxisHorizontal", int64(ScrollViewAxisHorizontal), int64(nui.ScrollViewAxisHorizontal)},
		{"ScrollViewAxisBoth", int64(ScrollViewAxisBoth), int64(nui.ScrollViewAxisBoth)},
		{"ScrollAnimationSmooth", int64(ScrollAnimationSmooth), int64(nui.ScrollAnimationSmooth)},
		{"ScrollAnimationInstant", int64(ScrollAnimationInstant), int64(nui.ScrollAnimationInstant)},
		{"ScrollBehaviorAlways", int64(ScrollBehaviorAlways), int64(nui.ScrollBehaviorAlways)},
		{"ScrollBehaviorAuto", int64(ScrollBehaviorAuto), int64(nui.ScrollBehaviorAuto)},
		{"ScrollBehaviorAsk", int64(ScrollBehaviorAsk), int64(nui.ScrollBehaviorAsk)},
		{"ScrollAlignmentEnd", int64(ScrollAlignmentEnd), int64(nui.ScrollAlignmentEnd)},
		{"ScrollAlignmentStart", int64(ScrollAlignmentStart), int64(nui.ScrollAlignmentStart)},
		{"SplitViewOrientationHorizontal", int64(SplitViewOrientationHorizontal), int64(nui.SplitViewOrientationHorizontal)},
		{"SplitViewOrientationVertical", int64(SplitViewOrientationVertical), int64(nui.SplitViewOrientationVertical)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}

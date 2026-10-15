// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// SplitViewOrientation is the orientation of a split view, see SplitView.Orientation.
type SplitViewOrientation int32

const (
	// SplitViewOrientationHorizontal places the contents side by side, which is the default.
	SplitViewOrientationHorizontal SplitViewOrientation = 1
	// SplitViewOrientationVertical places the contents on top of each other.
	SplitViewOrientationVertical SplitViewOrientation = 2
)

// TSplitView is a component, which divides the space into two resizable contents, see SplitView.
type TSplitView struct{ ref js.Ref }

func (TSplitView) isView() {}

// SplitView returns a split view of the two contents, which are split at the ratio 0.5 by default.
//
//gostrict:host
func SplitView(contentA View, contentB View) TSplitView {
	return TSplitView{ref: wrap(nui.SplitView(viewOf(contentA), viewOf(contentB)))}
}

// Value sets the split ratio from 0 to 1, which is the size of the first content.
//
//gostrict:host
func (c TSplitView) Value(value float64) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).Value(value))}
}

// InputValue binds the split ratio to the state, so that the user keeps the ratio between two renders.
//
//gostrict:host
func (c TSplitView) InputValue(state FloatState) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).InputValue(state.nago()))}
}

// Orientation sets the orientation, e.g. SplitViewOrientationVertical.
//
//gostrict:host
func (c TSplitView) Orientation(orientation SplitViewOrientation) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).Orientation(nui.SplitViewOrientation(orientation)))}
}

// Frame sets the layout frame of the split view.
//
//gostrict:host
func (c TSplitView) Frame(frame Frame) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).Frame(frame.nago()))}
}

// ContentA sets the first content, which is the left or the top one.
//
//gostrict:host
func (c TSplitView) ContentA(contentA View) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).ContentA(viewOf(contentA)))}
}

// ContentB sets the second content, which is the right or the bottom one.
//
//gostrict:host
func (c TSplitView) ContentB(contentB View) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).ContentB(viewOf(contentB)))}
}

// MinRatio sets a lower limit for the split ratio.
//
//gostrict:host
func (c TSplitView) MinRatio(minRatio float64) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).MinRatio(minRatio))}
}

// MaxRatio sets an upper limit for the split ratio.
//
//gostrict:host
func (c TSplitView) MaxRatio(maxRatio float64) TSplitView {
	return TSplitView{ref: wrap(get[nui.TSplitView](c.ref).MaxRatio(maxRatio))}
}

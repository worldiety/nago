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

// ScrollViewAxis is the scroll direction of a scroll view.
type ScrollViewAxis int32

const (
	// ScrollViewAxisVertical scrolls vertically. It is the default.
	ScrollViewAxisVertical ScrollViewAxis = 0
	// ScrollViewAxisHorizontal scrolls horizontally.
	ScrollViewAxisHorizontal ScrollViewAxis = 1
	// ScrollViewAxisBoth scrolls vertically and horizontally.
	ScrollViewAxisBoth ScrollViewAxis = 2
)

// ScrollAnimation tells, how a scroll view animates the scrolling into a view, see ScrollToView.
type ScrollAnimation int32

const (
	// ScrollAnimationSmooth scrolls smoothly.
	ScrollAnimationSmooth ScrollAnimation = 0
	// ScrollAnimationInstant jumps to the view.
	ScrollAnimationInstant ScrollAnimation = 1
)

// ScrollBehavior tells, how a scroll view behaves, when its scrollable content grows.
type ScrollBehavior int32

const (
	// ScrollBehaviorAlways always scrolls, when the content grows.
	ScrollBehaviorAlways ScrollBehavior = 0
	// ScrollBehaviorAuto scrolls automatically, when the content grows and is at the scroll end, otherwise it asks.
	ScrollBehaviorAuto ScrollBehavior = 1
	// ScrollBehaviorAsk asks the user, whether to scroll, when the content grows.
	ScrollBehaviorAsk ScrollBehavior = 2
)

// ScrollAlignment tells, how a scroll view aligns the scroll target, when it scrolls into view.
type ScrollAlignment int32

const (
	// ScrollAlignmentEnd aligns the bottom or right edge of the target with the edge of the scroll view.
	ScrollAlignmentEnd ScrollAlignment = 0
	// ScrollAlignmentStart aligns the top or left edge of the target with the edge of the scroll view.
	ScrollAlignmentStart ScrollAlignment = 1
)

// TScrollView is a scrollable container for a single child view, see ScrollView.
type TScrollView struct{ ref js.Ref }

func (TScrollView) isView() {}

// ScrollView returns a scroll view of the content. By default, it scrolls vertically.
//
//gostrict:host
func ScrollView(content View) TScrollView {
	return TScrollView{ref: wrap(nui.ScrollView(viewOf(content)))}
}

// Content sets the scrollable content.
//
//gostrict:host
func (c TScrollView) Content(content View) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Content(viewOf(content)))}
}

// Axis sets the scroll direction, e.g. ScrollViewAxisHorizontal.
//
//gostrict:host
func (c TScrollView) Axis(axis ScrollViewAxis) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Axis(nui.ScrollViewAxis(axis)))}
}

// Frame sets the layout frame of the scroll view.
//
//gostrict:host
func (c TScrollView) Frame(frame Frame) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Frame(frame.nago()))}
}

// Position sets the alignment of the content inside the scroll view.
//
//gostrict:host
func (c TScrollView) Position(position Position) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Position(position.nago()))}
}

// ScrollToView scrolls the view with the id into view, with the animation.
//
//gostrict:host
func (c TScrollView) ScrollToView(scrollToView string, animation ScrollAnimation) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).ScrollToView(scrollToView, nui.ScrollAnimation(animation)))}
}

// Border applies a border around the scroll view.
//
//gostrict:host
func (c TScrollView) Border(border Border) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Border(border.nago()))}
}

// Padding sets the inner padding of the scroll view.
//
//gostrict:host
func (c TScrollView) Padding(padding Padding) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).Padding(padding.nago()))}
}

// BackgroundColor sets the background color of the scroll view.
//
//gostrict:host
func (c TScrollView) BackgroundColor(color Color) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).BackgroundColor(nui.Color(color)))}
}

// ScrollBehavior defines, how the scroll view behaves, when its scrollable content grows.
//
//gostrict:host
func (c TScrollView) ScrollBehavior(behavior ScrollBehavior) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).ScrollBehavior(nui.ScrollBehavior(behavior)))}
}

// ScrollButtonLabel sets the label of the scroll button, when the scroll view asks whether to scroll.
//
//gostrict:host
func (c TScrollView) ScrollButtonLabel(label string) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).ScrollButtonLabel(label))}
}

// ScrollAlignment defines, how the scroll view aligns the scroll target, when it scrolls into view.
//
//gostrict:host
func (c TScrollView) ScrollAlignment(alignment ScrollAlignment) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).ScrollAlignment(nui.ScrollAlignment(alignment)))}
}

// ListLength sets the length of the content list. It is optional, but it should be set, if the content is a list,
// e.g. of chat messages or log entries.
//
//gostrict:host
func (c TScrollView) ListLength(length int32) TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).ListLength(int(length)))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TScrollView) FullWidth() TScrollView {
	return TScrollView{ref: wrap(get[nui.TScrollView](c.ref).FullWidth())}
}

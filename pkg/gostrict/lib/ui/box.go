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

// TBox is a layout component, see BoxAlign. It lays out its children by their alignment and clips them. This makes it
// suitable for overlapping layouts, but usually requires an absolute height and width.
type TBox struct{ ref js.Ref }

func (TBox) isView() {}

// BoxAlign returns a box with the child, which is aligned by alignment, e.g. Top, Center or BottomTrailing. The box
// clips its children, so shadows may require extra padding.
//
//gostrict:host
func BoxAlign(alignment Alignment, child View) TBox {
	return TBox{ref: wrap(nui.BoxAlign(nui.Alignment(alignment), viewOf(child)))}
}

// Padding sets the inner spacing around the box's children.
//
//gostrict:host
func (c TBox) Padding(p Padding) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).Padding(p.nago()).(nui.TBox))}
}

// BackgroundColor sets the background color of the box.
//
//gostrict:host
func (c TBox) BackgroundColor(backgroundColor Color) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).BackgroundColor(nui.Color(backgroundColor)).(nui.TBox))}
}

// Frame sets the layout frame of the box, including size and positioning.
//
//gostrict:host
func (c TBox) Frame(fr Frame) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).Frame(fr.nago()).(nui.TBox))}
}

// WithFrame changes the layout frame of the box with fn, which gets the current one.
//
//gostrict:host
func (c TBox) WithFrame(fn func(Frame) Frame) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TBox))}
}

// FullWidth sets the box to span the full available width.
//
//gostrict:host
func (c TBox) FullWidth() TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).FullWidth())}
}

// Font sets the font style for the text content inside the box.
//
//gostrict:host
func (c TBox) Font(font Font) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).Font(font.nago()).(nui.TBox))}
}

// Border sets the border styling of the box.
//
//gostrict:host
func (c TBox) Border(border Border) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).Border(border.nago()).(nui.TBox))}
}

// Visible shows or hides the box.
//
//gostrict:host
func (c TBox) Visible(visible bool) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).Visible(visible).(nui.TBox))}
}

// DisableOutsidePointerEvents controls, whether pointer events outside the content of the box are disabled.
//
//gostrict:host
func (c TBox) DisableOutsidePointerEvents(disable bool) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).DisableOutsidePointerEvents(disable))}
}

// AccessibilityLabel sets a label for screen readers.
//
//gostrict:host
func (c TBox) AccessibilityLabel(label string) TBox {
	return TBox{ref: wrap(get[nui.TBox](c.ref).AccessibilityLabel(label).(nui.TBox))}
}

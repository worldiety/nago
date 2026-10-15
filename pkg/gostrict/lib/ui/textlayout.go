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

// TTextLayout is a layout of multiple texts with a shared typography, spacing and alignment, see TextLayout.
type TTextLayout struct{ ref js.Ref }

func (TTextLayout) isView() {}

// TextLayout returns a layout, which arranges the views inline, e.g. the texts of a paragraph. The alignment of each
// Text is ignored.
//
//gostrict:host
func TextLayout(views ...View) TTextLayout {
	return TTextLayout{ref: wrap(nui.TextLayout(viewsOf(views)...))}
}

// Append appends the views.
//
//gostrict:host
func (c TTextLayout) Append(views ...View) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Append(viewsOf(views)...))}
}

// Action calls f, when the user clicks the layout.
//
//gostrict:host
func (c TTextLayout) Action(f func()) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Action(f))}
}

// Alignment aligns the text within the layout.
//
//gostrict:host
func (c TTextLayout) Alignment(alignment TextAlignment) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Alignment(nui.TextAlignment(alignment)))}
}

// Font sets the font of the texts in the layout.
//
//gostrict:host
func (c TTextLayout) Font(font Font) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Font(font.nago()))}
}

// Frame sets the size and the position of the layout.
//
//gostrict:host
func (c TTextLayout) Frame(f Frame) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Frame(f.nago()).(nui.TTextLayout))}
}

// WithFrame changes the size and the position with fn, which gets the current ones.
//
//gostrict:host
func (c TTextLayout) WithFrame(fn func(Frame) Frame) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TTextLayout))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TTextLayout) FullWidth() TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).FullWidth())}
}

// Padding sets the inner spacing of the layout.
//
//gostrict:host
func (c TTextLayout) Padding(padding Padding) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Padding(padding.nago()).(nui.TTextLayout))}
}

// BackgroundColor sets the color of the background of the layout.
//
//gostrict:host
func (c TTextLayout) BackgroundColor(backgroundColor Color) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).BackgroundColor(nui.Color(backgroundColor)).(nui.TTextLayout))}
}

// Border sets the border of the layout, e.g. its widths, radii, colors and shadow.
//
//gostrict:host
func (c TTextLayout) Border(border Border) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Border(border.nago()).(nui.TTextLayout))}
}

// Visible shows or hides the layout.
//
//gostrict:host
func (c TTextLayout) Visible(visible bool) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).Visible(visible).(nui.TTextLayout))}
}

// AccessibilityLabel sets the label for a screen reader, which describes the content or the purpose of the layout.
//
//gostrict:host
func (c TTextLayout) AccessibilityLabel(label string) TTextLayout {
	return TTextLayout{ref: wrap(get[nui.TTextLayout](c.ref).AccessibilityLabel(label).(nui.TTextLayout))}
}

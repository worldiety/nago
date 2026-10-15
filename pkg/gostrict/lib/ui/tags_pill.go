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
	"go.wdy.de/nago/presentation/ui/tags"
)

// TTagsColoredTextPill is a pill, which shows a short text in a rounded container with a background color, see
// tags.TColoredTextPill of nago.
type TTagsColoredTextPill struct{ ref js.Ref }

func (TTagsColoredTextPill) isView() {}

// TagsColoredTextPill returns a pill with the background color and the text, like tags.ColoredTextPill of nago.
//
//gostrict:host
func TagsColoredTextPill(color Color, text string) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(tags.ColoredTextPill(nui.Color(color), text))}
}

// Padding sets the inner spacing around the text of the pill.
//
//gostrict:host
func (c TTagsColoredTextPill) Padding(padding Padding) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).Padding(padding.nago()).(tags.TColoredTextPill))}
}

// WithFrame changes the layout frame of the pill with fn, which gets the current one.
//
//gostrict:host
func (c TTagsColoredTextPill) WithFrame(fn func(Frame) Frame) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(tags.TColoredTextPill))}
}

// Frame sets the layout frame of the pill, including size and alignment.
//
//gostrict:host
func (c TTagsColoredTextPill) Frame(frame Frame) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).Frame(frame.nago()).(tags.TColoredTextPill))}
}

// Border sets the border of the pill, such as radius or thickness.
//
//gostrict:host
func (c TTagsColoredTextPill) Border(border Border) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).Border(border.nago()).(tags.TColoredTextPill))}
}

// Visible shows or hides the pill.
//
//gostrict:host
func (c TTagsColoredTextPill) Visible(visible bool) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).Visible(visible).(tags.TColoredTextPill))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TTagsColoredTextPill) AccessibilityLabel(label string) TTagsColoredTextPill {
	return TTagsColoredTextPill{ref: wrap(get[tags.TColoredTextPill](c.ref).AccessibilityLabel(label).(tags.TColoredTextPill))}
}

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
	"go.wdy.de/nago/presentation/ui/switcher"
)

// SwitcherLayout is the layout of a switcher, like switcher.SwitcherLayout of nago.
type SwitcherLayout int32

const (
	// SwitcherLayoutAuto decides by the size of the window.
	SwitcherLayoutAuto SwitcherLayout = 0
	// SwitcherLayoutVertical stacks the pages vertically.
	SwitcherLayoutVertical SwitcherLayout = 1
	// SwitcherLayoutHorizontal shows the pages horizontally.
	SwitcherLayoutHorizontal SwitcherLayout = 2
)

// TSwitcher is a content component, which switches between its pages, see Switcher.
type TSwitcher struct{ ref js.Ref }

func (TSwitcher) isView() {}

// ID assigns a unique identifier to the switcher.
//
//gostrict:host
func (c TSwitcher) ID(id string) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).ID(id))}
}

// Append adds pages to the switcher.
//
//gostrict:host
func (c TSwitcher) Append(pages ...TSwitcherPage) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).Append(switcherPagesOf(pages)...))}
}

// Frame sets the frame of the switcher.
//
//gostrict:host
func (c TSwitcher) Frame(frame Frame) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).Frame(frame.nago()))}
}

// FullWidth sets the frame of the switcher to the full width.
//
//gostrict:host
func (c TSwitcher) FullWidth() TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).FullWidth())}
}

// ContentNoPadding lets the content keep its own padding.
//
//gostrict:host
func (c TSwitcher) ContentNoPadding() TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).ContentNoPadding())}
}

// DynamicHeight lets the switcher change its height by the active page.
//
//gostrict:host
func (c TSwitcher) DynamicHeight() TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).DynamicHeight())}
}

// InputValue binds the switcher to a string state, which holds the id of the active page.
//
//gostrict:host
func (c TSwitcher) InputValue(input StringState) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).InputValue(input.nago()))}
}

// With changes the switcher with fn, which gets the switcher and returns the changed one.
//
//gostrict:host
func (c TSwitcher) With(fn func(s TSwitcher) TSwitcher) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).With(func(s switcher.TSwitcher) switcher.TSwitcher {
		return get[switcher.TSwitcher](fn(TSwitcher{ref: wrap(s)}).ref)
	}))}
}

// Layout sets the layout of the switcher.
//
//gostrict:host
func (c TSwitcher) Layout(layout SwitcherLayout) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).Layout(switcher.SwitcherLayout(layout)))}
}

// ImageObjectFit sets how the images of the pages are fitted.
//
//gostrict:host
func (c TSwitcher) ImageObjectFit(objectFit ObjectFit) TSwitcher {
	return TSwitcher{ref: wrap(get[switcher.TSwitcher](c.ref).ImageObjectFit(nui.ObjectFit(objectFit)))}
}

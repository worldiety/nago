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
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/tabs"
)

// TTabsPage is a page of tabs, see TabsPage.
type TTabsPage struct{ ref js.Ref }

// TabsPage creates a page with the title and the body, which renders the content of the page, like tabs.Page of
// nago.
//
//gostrict:host
func TabsPage(title string, body func() View) TTabsPage {
	return TTabsPage{ref: wrap(tabs.Page(title, func() core.View {
		return viewOf(body())
	}))}
}

// Icon sets the icon, which is shown next to the title of the page.
//
//gostrict:host
func (c TTabsPage) Icon(ico SVG) TTabsPage {
	return TTabsPage{ref: wrap(get[tabs.TPage](c.ref).Icon(core.SVG(ico)))}
}

// Disabled marks the page as disabled, which is visible, but not selectable.
//
//gostrict:host
func (c TTabsPage) Disabled(disabled bool) TTabsPage {
	return TTabsPage{ref: wrap(get[tabs.TPage](c.ref).Disabled(disabled))}
}

// TTabs is a container of tabs, which shows the body of the active page, see Tabs.
type TTabs struct{ ref js.Ref }

func (TTabs) isView() {}

// Tabs returns a container of the pages, like tabs.Tabs of nago.
//
//gostrict:host
func Tabs(pages ...TTabsPage) TTabs {
	return TTabs{ref: wrap(tabs.Tabs(tabsPagesOf(pages)...))}
}

// Frame sets the layout frame of the container.
//
//gostrict:host
func (c TTabs) Frame(frame Frame) TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).Frame(frame.nago()))}
}

// FullWidth sets the container to the full width of the parent.
//
//gostrict:host
func (c TTabs) FullWidth() TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).FullWidth())}
}

// ButtonAlignment sets the alignment of the tab buttons.
//
//gostrict:host
func (c TTabs) ButtonAlignment(tabAlignment Alignment) TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).ButtonAlignment(nui.Alignment(tabAlignment)))}
}

// PageTabSpace sets the space between the tab buttons and the content of the page.
//
//gostrict:host
func (c TTabs) PageTabSpace(space Length) TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).PageTabSpace(nui.Length(space)))}
}

// Position sets the position of the tab buttons, e.g. top or bottom.
//
//gostrict:host
func (c TTabs) Position(pos Position) TTabs {
	return TTabs{ref: wrap(get[tabs.TTabs](c.ref).Position(pos.nago()))}
}

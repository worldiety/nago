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
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

// CardTitleStyle is the style of the title of a card, see CardTitleLarge. Like cardlayout.TitleStyle of
// nago.
type CardTitleStyle int32

const (
	// CardTitleLarge shows a large title above a line.
	CardTitleLarge CardTitleStyle = 0
	// CardTitleCompact shows a compact title on a colored bar.
	CardTitleCompact CardTitleStyle = 1
)

// TCardLayout is a responsive grid of children, see CardLayout.
type TCardLayout struct{ ref js.Ref }

func (TCardLayout) isView() {}

// CardLayout returns a grid of the children, which has three columns by default. Nil children are skipped. Like
// cardlayout.Layout of nago.
//
//gostrict:host
func CardLayout(children ...View) TCardLayout {
	return TCardLayout{ref: wrap(cardlayout.Layout(viewsOf(children)...))}
}

// Columns sets the number of columns for the window size class.
//
//gostrict:host
func (c TCardLayout) Columns(class WindowSizeClass, columns int32) TCardLayout {
	return TCardLayout{ref: wrap(get[cardlayout.TCardLayout](c.ref).Columns(core.WindowSizeClass(class), int(columns)))}
}

// Frame sets the frame of the layout.
//
//gostrict:host
func (c TCardLayout) Frame(frame Frame) TCardLayout {
	return TCardLayout{ref: wrap(get[cardlayout.TCardLayout](c.ref).Frame(frame.nago()))}
}

// Padding sets the inner spacing of the layout.
//
//gostrict:host
func (c TCardLayout) Padding(padding Padding) TCardLayout {
	return TCardLayout{ref: wrap(get[cardlayout.TCardLayout](c.ref).Padding(padding.nago()))}
}

// TCard is a card with a title, a body and an optional footer, see Card.
type TCard struct{ ref js.Ref }

func (TCard) isView() {}

// Card returns a card with the title and a default padding. Like cardlayout.Card of nago.
//
//gostrict:host
func Card(title string) TCard {
	return TCard{ref: wrap(cardlayout.Card(title))}
}

// Style sets the style of the title.
//
//gostrict:host
func (c TCard) Style(style CardTitleStyle) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).Style(cardlayout.TitleStyle(style)))}
}

// ID sets the unique identifier of the card.
//
//gostrict:host
func (c TCard) ID(id string) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).ID(id))}
}

// Body sets the main content of the card.
//
//gostrict:host
func (c TCard) Body(body View) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).Body(viewOf(body)))}
}

// Padding overrides the default padding of the card.
//
//gostrict:host
func (c TCard) Padding(padding Padding) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).Padding(padding.nago()))}
}

// Footer sets a view below the body, typically for actions.
//
//gostrict:host
func (c TCard) Footer(footer View) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).Footer(viewOf(footer)))}
}

// Frame sets the layout frame of the card.
//
//gostrict:host
func (c TCard) Frame(frame Frame) TCard {
	return TCard{ref: wrap(get[cardlayout.TCard](c.ref).Frame(frame.nago()))}
}

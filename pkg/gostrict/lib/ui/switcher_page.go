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
	"go.wdy.de/nago/presentation/ui/switcher"
)

// TSwitcherPage is a page of a switcher, see SwitcherPage.
type TSwitcherPage struct{ ref js.Ref }

// SwitcherPage creates a page with the id, the title, the icon and the content, like switcher.SwitcherPage of nago.
//
//gostrict:host
func SwitcherPage(id, title string, icon SVG, content View) TSwitcherPage {
	return TSwitcherPage{ref: wrap(switcher.SwitcherPage(id, title, core.SVG(icon), viewOf(content)))}
}

// Title sets the title of the page.
//
//gostrict:host
func (c TSwitcherPage) Title(title string) TSwitcherPage {
	return TSwitcherPage{ref: wrap(get[switcher.TSwitcherPage](c.ref).Title(title))}
}

// Icon sets the icon of the page, which the toggle shows.
//
//gostrict:host
func (c TSwitcherPage) Icon(icon SVG) TSwitcherPage {
	return TSwitcherPage{ref: wrap(get[switcher.TSwitcherPage](c.ref).Icon(core.SVG(icon)))}
}

// Content sets the content of the page.
//
//gostrict:host
func (c TSwitcherPage) Content(content View) TSwitcherPage {
	return TSwitcherPage{ref: wrap(get[switcher.TSwitcherPage](c.ref).Content(viewOf(content)))}
}

// Img sets an optional banner image of the page.
//
//gostrict:host
func (c TSwitcherPage) Img(imgUri URI) TSwitcherPage {
	return TSwitcherPage{ref: wrap(get[switcher.TSwitcherPage](c.ref).Img(core.URI(imgUri)))}
}

// ImgAdaptive sets an optional banner image of the page by the light or the dark mode.
//
//gostrict:host
func (c TSwitcherPage) ImgAdaptive(light, dark URI) TSwitcherPage {
	return TSwitcherPage{ref: wrap(get[switcher.TSwitcherPage](c.ref).ImgAdaptive(core.URI(light), core.URI(dark)))}
}

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
	"go.wdy.de/nago/presentation/ui/hero"
)

// THero is a prominent section with a title, a subtitle and actions, like hero.THero of nago. See Hero.
type THero struct{ ref js.Ref }

func (THero) isView() {}

// Hero returns a hero with the title and a full width height of 320, like hero.Hero of nago.
//
//gostrict:host
func Hero(title string) THero {
	return THero{ref: wrap(hero.Hero(title))}
}

// Frame sets the frame of the hero section.
//
//gostrict:host
func (c THero) Frame(frame Frame) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).Frame(frame.nago()))}
}

// Actions sets the action buttons or links of the hero section.
//
//gostrict:host
func (c THero) Actions(actions ...View) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).Actions(viewsOf(actions)...))}
}

// Padding sets the padding of the hero section.
//
//gostrict:host
func (c THero) Padding(padding Padding) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).Padding(padding.nago()))}
}

// Subtitle sets the subtitle text of the hero section.
//
//gostrict:host
func (c THero) Subtitle(text string) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).Subtitle(text))}
}

// SubtitleView sets the subtitle view of the hero section, which replaces the subtitle text.
//
//gostrict:host
func (c THero) SubtitleView(view View) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).SubtitleView(viewOf(view)))}
}

// Alignment sets the alignment of the hero section.
//
//gostrict:host
func (c THero) Alignment(alignment Alignment) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).Alignment(nui.Alignment(alignment)))}
}

// BackgroundImage places a cover image into the background.
//
//gostrict:host
func (c THero) BackgroundImage(img URI) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).BackgroundImage(core.URI(img)))}
}

// ForegroundColor sets the color of the foreground.
//
//gostrict:host
func (c THero) ForegroundColor(col Color) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).ForegroundColor(nui.Color(col)))}
}

// ForegroundColorAdaptive sets the color of the foreground for the light and the dark mode.
//
//gostrict:host
func (c THero) ForegroundColorAdaptive(onLight, onDark Color) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).ForegroundColorAdaptive(nui.Color(onLight), nui.Color(onDark)))}
}

// BackgroundColor sets the background color of the hero section.
//
//gostrict:host
func (c THero) BackgroundColor(color Color) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).BackgroundColor(nui.Color(color)))}
}

// TextColor sets the color of the texts.
//
//gostrict:host
func (c THero) TextColor(color Color) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).TextColor(nui.Color(color)))}
}

// SideView sets a view, which is displayed alongside the text content.
//
//gostrict:host
func (c THero) SideView(view View) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).SideView(viewOf(view)))}
}

// SideSVG sets an SVG, which is displayed alongside the text content.
//
//gostrict:host
func (c THero) SideSVG(svg SVG) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).SideSVG(core.SVG(svg)))}
}

// SideImage sets an image, which is displayed alongside the text content.
//
//gostrict:host
func (c THero) SideImage(img TImage) THero {
	return THero{ref: wrap(get[hero.THero](c.ref).SideImage(get[nui.TImage](img.ref)))}
}

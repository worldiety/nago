// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package switcher

import (
	"go.wdy.de/nago/presentation/core"
)

// TSwitcherPage is a content page, that is used in TSwitcher
type TSwitcherPage struct {
	id                string
	title             string
	icon              core.SVG
	content           core.View
	lightUri, darkUri core.URI
}

// SwitcherPage creates a page to be used in a TSwitcher
func SwitcherPage(id, title string, icon core.SVG, content core.View) TSwitcherPage {
	return TSwitcherPage{
		id:      id,
		title:   title,
		icon:    icon,
		content: content,
	}
}

// Title sets the title of a switcher page
func (c TSwitcherPage) Title(title string) TSwitcherPage {
	c.title = title
	return c
}

// Icon sets the toggle icon of a switcher page
func (c TSwitcherPage) Icon(icon core.SVG) TSwitcherPage {
	c.icon = icon
	return c
}

// Content sets the content of a switcher page
func (c TSwitcherPage) Content(content core.View) TSwitcherPage {
	c.content = content
	return c
}

// Img sets an optional banner image uri
func (c TSwitcherPage) Img(imgUri core.URI) TSwitcherPage {
	c.lightUri = imgUri
	return c
}

// ImgAdaptive sets an optional banner image uri by light/dark mode
func (c TSwitcherPage) ImgAdaptive(light, dark core.URI) TSwitcherPage {
	c.lightUri = light
	c.darkUri = dark
	return c
}

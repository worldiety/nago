// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package avatar

import (
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"

	"strings"
	"unicode"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Style int

const (
	Circle Style = iota
	Rounded
)

// TAvatar is a basic component (Avatar).
// This component displays a user or entity representation,
// either as an image (via URL, raw data, or image ID) or as
// a fallback with initials (paraphe). It can be styled with
// frame, border, text size, and color, and also supports an
// optional click action.
type TAvatar struct {
	paraphe  string
	url      core.URI
	data     []byte
	action   func()
	frame    ui.Frame
	textSize ui.Length
	border   ui.Border
	color    ui.Color
	imgID    image.ID
}

// TextOrImage creates an avatar from either an image (if provided) or falls back to a text-based avatar.
func TextOrImage(text string, img image.ID) TAvatar {
	if img != "" {
		c := URI(httpimage.URI(img, image.FitCover, 64, 64))
		c.imgID = img
		return c
	}

	return Text(text)
}

// Text creates a text-based avatar using initials derived from the given string.
func Text(paraphe string) TAvatar {
	if paraphe == "" {
		paraphe = "?"
	} else {

		var sb strings.Builder
		if len(paraphe) > 2 {
			tokens := strings.Split(paraphe, " ")
			for i, token := range tokens {
				if i > 1 {
					break
				}

				for _, r := range token {
					sb.WriteRune(unicode.ToUpper(r))
					break
				}
			}
		}

		if sb.Len() == 0 {
			paraphe = "?"
		} else {
			paraphe = sb.String()
		}
	}

	return TAvatar{
		paraphe: paraphe,
		color:   ui.M5,
		border:  ui.Border{}.Circle(),
	}.Size(ui.L40)
}

// URI creates an avatar from a given image URL.
func URI(uri core.URI) TAvatar {
	return TAvatar{
		url:    uri,
		color:  ui.M5,
		border: ui.Border{}.Circle(),
	}.Size(ui.L40)
}

// Embed creates an avatar directly from raw image data.
func Embed(data []byte) TAvatar {
	return TAvatar{
		data:   data,
		color:  ui.M5,
		border: ui.Border{}.Circle(),
	}.Size(ui.L40)
}

// Border sets the border style of the avatar.
func (c TAvatar) Border(border ui.Border) TAvatar {
	c.border = border
	return c
}

// Action sets an optional click action for the avatar.
func (c TAvatar) Action(fn func()) TAvatar {
	c.action = fn
	return c
}

// Size sets the avatar's size and adjusts text size and image resolution accordingly.
func (c TAvatar) Size(widthAndHeight ui.Length) TAvatar {
	c.frame = ui.Frame{}.Size(widthAndHeight, widthAndHeight)
	c.textSize = widthAndHeight.Mul(0.4)
	if c.imgID != "" {
		s := int(max(widthAndHeight.Estimate(), 64))
		c.url = httpimage.URI(c.imgID, image.FitCover, s, s)
	}

	return c
}

// Style sets the avatar’s border style (circle by default, rounded when specified).
func (c TAvatar) Style(style Style) TAvatar {
	switch style {
	default:
		c.border = ui.Border{}.Circle()
	case Rounded:
		c.border = ui.Border{}.Radius(ui.L8)
	}

	return c
}

// Render draws the avatar as either text initials or an image, applying frame, border, and optional action.
func (c TAvatar) Render(ctx core.RenderContext) core.RenderNode {
	c.frame.MinWidth = c.frame.Width // force the correct dimensions in flex layouts
	c.frame.MinHeight = c.frame.Height

	if c.paraphe != "" {
		return ui.VStack(ui.Text(c.paraphe).Font(ui.Font{Size: c.textSize}).Color("#000000")).
			Action(c.action).
			BackgroundColor(c.color).
			Frame(c.frame).
			Border(c.border).
			Render(ctx)
	}

	// must be an image
	img := ui.Image()
	if c.data != nil {
		img = img.Embed(c.data)
	}

	if c.url != "" {
		img = img.URI(c.url)
	}

	return ui.VStack(img).
		Action(c.action).
		Frame(c.frame).
		Border(c.border).
		Render(ctx)
}

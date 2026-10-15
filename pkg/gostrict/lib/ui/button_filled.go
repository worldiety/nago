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
)

// TFilledButton is a button with a solid background color, see FilledButton.
type TFilledButton struct{ ref js.Ref }

func (TFilledButton) isView() {}

// FilledButton returns a button with the background color, which calls action, when the user clicks it.
//
//gostrict:host
func FilledButton(fillColor Color, action func()) TFilledButton {
	return TFilledButton{ref: wrap(nui.FilledButton(nui.Color(fillColor), action))}
}

// Title sets the text label of the button.
//
//gostrict:host
func (c TFilledButton) Title(text string) TFilledButton {
	return TFilledButton{ref: wrap(get[nui.TFilledButton](c.ref).Title(text))}
}

// PreIcon sets the icon, which is shown before the text label.
//
//gostrict:host
func (c TFilledButton) PreIcon(svg SVG) TFilledButton {
	return TFilledButton{ref: wrap(get[nui.TFilledButton](c.ref).PreIcon(core.SVG(svg)))}
}

// PostIcon sets the icon, which is shown after the text label.
//
//gostrict:host
func (c TFilledButton) PostIcon(svg SVG) TFilledButton {
	return TFilledButton{ref: wrap(get[nui.TFilledButton](c.ref).PostIcon(core.SVG(svg)))}
}

// TextColor sets the color of the text label.
//
//gostrict:host
func (c TFilledButton) TextColor(color Color) TFilledButton {
	return TFilledButton{ref: wrap(get[nui.TFilledButton](c.ref).TextColor(nui.Color(color)))}
}

// Frame sets the size and the position of the button.
//
//gostrict:host
func (c TFilledButton) Frame(frame Frame) TFilledButton {
	return TFilledButton{ref: wrap(get[nui.TFilledButton](c.ref).Frame(frame.nago()))}
}

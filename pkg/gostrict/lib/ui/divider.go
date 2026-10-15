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

// TDivider is a divider, see HLine and VLine. It is typically used to visually separate sections of content.
type TDivider struct{ ref js.Ref }

func (TDivider) isView() {}

// HLineWithColor returns a horizontal hairline divider of the color, e.g. within a VStack.
//
//gostrict:host
func HLineWithColor(c Color) TDivider {
	return TDivider{ref: wrap(nui.HLineWithColor(nui.Color(c)))}
}

// HLine returns a horizontal hairline divider, e.g. within a VStack. The color is derived from the main color.
//
//gostrict:host
func HLine() TDivider {
	return TDivider{ref: wrap(nui.HLine())}
}

// VLineWithColor returns a vertical hairline divider of the color, e.g. within an HStack.
//
//gostrict:host
func VLineWithColor(c Color) TDivider {
	return TDivider{ref: wrap(nui.VLineWithColor(nui.Color(c)))}
}

// VLine returns a vertical hairline divider, e.g. within an HStack. The color is derived from the main color.
//
//gostrict:host
func VLine() TDivider {
	return TDivider{ref: wrap(nui.VLine())}
}

// Padding sets the spacing around the divider.
//
//gostrict:host
func (c TDivider) Padding(padding Padding) TDivider {
	return TDivider{ref: wrap(get[nui.TDivider](c.ref).Padding(padding.nago()))}
}

// Frame sets the layout frame of the divider, including size and positioning.
//
//gostrict:host
func (c TDivider) Frame(frame Frame) TDivider {
	return TDivider{ref: wrap(get[nui.TDivider](c.ref).Frame(frame.nago()))}
}

// Border sets the border styling of the divider, typically its line thickness and color.
//
//gostrict:host
func (c TDivider) Border(border Border) TDivider {
	return TDivider{ref: wrap(get[nui.TDivider](c.ref).Border(border.nago()))}
}

// Visible shows or hides the divider.
//
//gostrict:host
func (c TDivider) Visible(visible bool) TDivider {
	return TDivider{ref: wrap(get[nui.TDivider](c.ref).Visible(visible))}
}

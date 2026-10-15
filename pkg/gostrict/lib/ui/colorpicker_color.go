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
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

// TColorpickerColor is a color swatch, see ColorpickerColor.
type TColorpickerColor struct{ ref js.Ref }

func (TColorpickerColor) isView() {}

// ColorpickerColor returns a swatch of the color, like colorpicker.Color of nago.
//
//gostrict:host
func ColorpickerColor(color Color) TColorpickerColor {
	return TColorpickerColor{ref: wrap(colorpicker.Color(nui.Color(color)))}
}

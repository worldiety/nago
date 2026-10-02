// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

func init() {
	app.Register("palette-picker", func(wnd core.Window) core.View {
		color := core.AutoState[Color](wnd).Init(func() Color { return colorpicker.DefaultPalette[3] })

		return colorpicker.PalettePicker("Label color", colorpicker.DefaultPalette).
			Value(color.Get()).
			State(color).
			Title("Choose a color").
			SupportingText("Used for the label in the calendar").
			Frame(Frame{Width: L320})
	})
}

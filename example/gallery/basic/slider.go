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
	"go.wdy.de/nago/presentation/ui/slider"
)

func init() {
	app.Register("slider", func(wnd core.Window) core.View {
		volume := core.AutoState[float64](wnd).Init(func() float64 { return 40 })
		price := core.AutoState[slider.RangeSliderValue](wnd).Init(func() slider.RangeSliderValue {
			return slider.RangeSliderValue{From: 20, To: 60}
		})

		return VStack(
			slider.Slider(0, 100).
				Label("Volume").
				Unit("%").
				InputValue(volume).
				Frame(Frame{Width: L400}),
			slider.RangeSlider(0, 100).
				Label("Price").
				Unit("€").
				Step(5).
				ShowMarkers(true).
				InputValue(price).
				Frame(Frame{Width: L400}),
		).Alignment(Leading).Gap(L32)
	})
}

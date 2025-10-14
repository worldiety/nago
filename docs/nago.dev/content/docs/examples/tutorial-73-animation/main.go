// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_72")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			colorIdx := core.AutoState[int](wnd)
			colorValue := core.AutoState[ui.Color](wnd)
			opacityValue := core.AutoState[float64](wnd)
			animation := core.AutoState[ui.Animation](wnd).Init(func() ui.Animation {
				return ui.AnimateBounce
			})

			transformation := core.AutoState[ui.Transformation](wnd)

			paddingIdx := core.AutoState[int](wnd)
			paddings := []ui.Length{"1rem", "2rem", "4rem", "8rem"}

			return ui.VStack(
				ui.VStack(ui.Text("hello world")).
					BackgroundColor(colorpicker.DefaultPalette[colorIdx.Get()%len(colorpicker.DefaultPalette)]).
					With(func(stack ui.TVStack) ui.TVStack {
						if colorValue.Get() != "" {
							stack = stack.BackgroundColor(colorValue.Get())
						}

						stack = stack.Transformation(transformation.Get())

						return stack
					}).
					Animation(animation.Get()).
					Opacity(opacityValue.Get()).
					Padding(ui.Padding{}.All(paddings[paddingIdx.Get()%len(paddings)])).
					Border(ui.Border{}.Radius(ui.L16)),

				ui.PrimaryButton(func() {
					colorIdx.Set(colorIdx.Get() + 1)
					paddingIdx.Set(paddingIdx.Get() + 1)
					colorValue.Set("")
					opacityValue.Set(0)
				}).Title("next style"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateNone)
					colorValue.Set("")
					colorIdx.Set(0)
					transformation.Set(ui.Transformation{})
				}).Title("none"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateBounce)
				}).Title("bounce"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateSpin)
				}).Title("spin"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimatePulse)
				}).Title("pulse"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimatePing)
				}).Title("ping"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
				}).Title("transition"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					transformation.Set(ui.Transformation{RotateZ: 0})
				}).Title("rotate 0"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					transformation.Set(ui.Transformation{RotateZ: 90})
				}).Title("rotate 90"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					transformation.Set(ui.Transformation{RotateZ: 180})
				}).Title("rotate 180"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					transformation.Set(ui.Transformation{RotateZ: 270})
				}).Title("rotate 270"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					colorValue.Set("#E8442900")

					go func() {
						time.Sleep(500 * time.Millisecond)
						colorValue.Set("#E84429ff")
						transformation.Set(ui.Transformation{TranslateX: "10rem"})
					}()
				}).Title("custom fade in"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					colorValue.Set("#E84429ff")

					go func() {
						time.Sleep(500 * time.Millisecond)
						colorValue.Set("#E8442900")
						transformation.Set(ui.Transformation{TranslateX: ""})
					}()
				}).Title("custom fade out"),

				ui.SecondaryButton(func() {
					animation.Set(ui.AnimateTransition)
					opacityValue.Set(1)

					go func() {
						time.Sleep(500 * time.Millisecond)
						opacityValue.Set(0)
					}()
				}).Title("Opacity"),
			).Gap(ui.L8).
				Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

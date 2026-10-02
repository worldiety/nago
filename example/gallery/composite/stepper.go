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
	"go.wdy.de/nago/presentation/ui/stepper"
)

func init() {
	app.Register("stepper", func(wnd core.Window) core.View {
		current := core.AutoState[int](wnd).Init(func() int { return 1 })

		return VStack(
			stepper.Stepper(
				stepper.Step().Headline("Cart").SupportingText("Review your items"),
				stepper.Step().Headline("Address").SupportingText("Where to ship"),
				stepper.Step().Headline("Payment").SupportingText("Choose a method"),
				stepper.Step().Headline("Done").SupportingText("Order placed"),
			).InputValue(current).Layout(stepper.StepperLayoutHorizontal),
		).Frame(Frame{Width: L880})
	})
}

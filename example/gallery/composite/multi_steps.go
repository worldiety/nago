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
	"go.wdy.de/nago/presentation/ui/form"
)

func init() {
	app.Register("multi-steps", func(wnd core.Window) core.View {
		step := core.AutoState[int](wnd).Init(func() int { return 1 })
		street := core.AutoState[string](wnd).Init(func() string { return "Marie-Curie-Str. 1" })
		city := core.AutoState[string](wnd).Init(func() string { return "Oldenburg" })

		return form.MultiSteps(
			form.Step(Text("Your account is ready.")).Headline("Account"),
			form.Step(
				VStack(
					TextField("Street", street.Get()).InputValue(street).FullWidth(),
					TextField("City", city.Get()).InputValue(city).FullWidth(),
				).Gap(L16).FullWidth(),
			).Headline("Address").SupportingText("Where should we ship to?"),
			form.Step(Text("Check your input and confirm.")).Headline("Confirm"),
		).
			InputValue(step).
			ButtonDone(PrimaryButton(func() {
				// save the data
			}).Title("Finish")).
			Frame(Frame{Width: L560})
	})
}

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
	app.Register("form", func(wnd core.Window) core.View {
		firstName := core.AutoState[string](wnd).Init(func() string { return "Ada" })
		lastName := core.AutoState[string](wnd).Init(func() string { return "Lovelace" })
		email := core.AutoState[string](wnd)

		submit := func() {
			// validate and save the input
		}

		return Form(
			form.Fieldset(
				VStack(
					HStack(
						TextField("First name", firstName.Get()).InputValue(firstName),
						TextField("Last name", lastName.Get()).InputValue(lastName),
					).Gap(L16),
					TextField("E-mail", email.Get()).InputValue(email).FullWidth(),
				).Gap(L16).Alignment(Leading),
			).Title("Personal data"),
			HStack(PrimaryButton(submit).Title("Save")).Alignment(Trailing).FullWidth(),
		).Action(submit).Autocomplete(true).Frame(Frame{Width: L560})
	})
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func check(firstname, lastname, email *core.State[string]) core.View {
	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap
		ui.Text("Sind die Daten korrekt?"),
		ui.TextField("E-Mail Adresse", email.Get()).
			InputValue(email).
			Disabled(true).
			FullWidth(),
		ui.TextField("Vorname", firstname.Get()).
			InputValue(firstname).
			Disabled(true).
			FullWidth(),
		ui.TextField("Nachname", lastname.Get()).
			InputValue(lastname).
			Disabled(true).
			FullWidth(),
	).FullWidth().Gap(ui.L8)
}

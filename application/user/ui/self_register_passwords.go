// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func passwords(password, passwordRepeated, errPasswordRepeated *core.State[string]) core.View {
	strength := user.CalculatePasswordStrength(password.Get())

	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap
		ui.PasswordField("Passwort", password.Get()).
			ErrorText(errPasswordRepeated.Get()).
			InputValue(password).
			FullWidth(),
		ui.PasswordField("Passwort wiederholen", passwordRepeated.Get()).
			ErrorText(errPasswordRepeated.Get()).
			InputValue(passwordRepeated).
			FullWidth(),

		ui.Space(ui.L2),
		PasswordStrengthView(strength),
	).FullWidth().Gap(ui.L8)
}

func validatePasswords(password, passwordRepeated, errPasswordRepeated *core.State[string]) user.PasswordStrengthIndicator {
	errPasswordRepeated.Set("")

	strength := user.CalculatePasswordStrength(password.Get())

	if password.Get() != passwordRepeated.Get() {
		errPasswordRepeated.Set("Die Passwörter stimmen nicht überein")
		strength.Acceptable = false
	}

	return strength
}

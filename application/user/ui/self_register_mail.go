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
	"log/slog"
)

func emails(email, emailRepeated, errMailRepeated *core.State[string]) core.View {
	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap
		ui.Text("Die E-Mail Adresse wird mit einem Passwort als Login verwendet. Im Nachgang erhalten Sie eine E-Mail, mit der Sie das Konto bestätigen müssen."),
		ui.TextField("E-Mail Adresse", email.Get()).
			ErrorText(errMailRepeated.Get()).
			InputValue(email).
			FullWidth(),
		ui.TextField("E-Mail Adresse wiederholen", emailRepeated.Get()).
			ErrorText(errMailRepeated.Get()).
			InputValue(emailRepeated).
			FullWidth(),
	).FullWidth().Gap(ui.L8)
}

func validateEmails(emailUsed user.EMailUsed, email, emailRepeated, errMailRepeated *core.State[string]) bool {
	errMailRepeated.Set("")
	if email.Get() != emailRepeated.Get() {
		errMailRepeated.Set("Die E-Mail Adressen stimmen nicht überein.")
	} else if email.Get() == "" {
		errMailRepeated.Set("Sie müssen eine E-Mail Adresse angeben.")
	} else if !user.Email(email.Get()).Valid() {
		errMailRepeated.Set("Dieses E-Mail Format wird nicht akzeptiert.")
	} else {

		// security note: this exposes facts about the existence of a user and opens a potential attack vector.
		// I don't know, how to mitigate that effectively, however other platforms expose that information as well.
		used, err := emailUsed(user.Email(email.Get()))
		if err != nil {
			slog.Error("error checking mail", "err", err)
			return false
		}

		if used {
			errMailRepeated.Set("Diese E-Mail-Adresse ist bereits registriert.")
		}

	}

	return errMailRepeated.Get() == ""
}

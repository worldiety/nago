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
	"go.wdy.de/nago/presentation/ui/alert"
)

func registerResult(err error) core.View {
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.Text("Ihre Registrierung war erfolgreich. Prüfen Sie nun Ihr Postfach, um die Registrierung abzuschließen und das Konto zu verifizieren."),
	).FullWidth().Gap(ui.L8)
}

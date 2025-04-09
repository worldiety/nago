// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisession

import (
	"fmt"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func Logout(wnd core.Window, logoutFn session.Logout) core.View {
	return ui.VStack(
		ui.IfElse(wnd.Subject().Valid(),
			ui.VStack(
				ui.Text(fmt.Sprintf("Sie sind derzeit als %s angemeldet.", wnd.Subject().Name())).TextAlignment(ui.TextAlignCenter),
				ui.PrimaryButton(func() {
					_, err := logoutFn(wnd.Session().ID())
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.UpdateSubject(nil)
				}).Title("Jetzt abmelden"),
			).Gap(ui.L16),
			ui.Text("Sie sind abgemeldet.").TextAlignment(ui.TextAlignCenter),
		),
	).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
}

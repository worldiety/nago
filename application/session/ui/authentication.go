// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisession

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrInvalidNonce = i18n.MustString("nago.iam.login.sso.invalid_nonce", i18n.Values{language.English: "Invalid authentication flow. Try again.", language.German: "Die Anmeldung ist fehlgeschlagen. Probieren Sie es erneut."})
)

func PageNLSAuthentication(wnd core.Window, exchange session.ExchangeNLS) core.View {
	nonce := session.NLSNonce(wnd.Values()["nonce"])
	if nonce == "" {
		return ui.VStack(alert.Banner(rstring.LabelError.Get(wnd), StrInvalidNonce.Get(wnd))).Frame(ui.Frame{}.MatchScreen())
	}

	xsync.Go(func() error {
		redirect, err := exchange(wnd.Session().ID(), nonce)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		if redirect == "" {
			wnd.Navigation().ForwardTo(".", nil)
			return nil
		}

		core.HTTPOpen(wnd.Navigation(), core.URI(redirect), "_self")
		return nil
	}, func(err error) {
		if err != nil {
			alert.ShowBannerError(wnd, err)
		}
	})

	return ui.VStack(
		alert.BannerMessages(wnd),
		ui.Text(rstring.LabelPleaseWait.Get(wnd)),
	).Frame(ui.Frame{}.MatchScreen())
}

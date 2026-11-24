// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uionboarding

import (
	"cmp"
	"strings"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accordion"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/tags"
)

// fonts of the banners: one size, a reason emphasized
var (
	bannerFont     = ui.Font{Size: ui.L14}
	bannerLeadFont = ui.Font{Size: ui.L14, Weight: ui.HeadlineAndTitleFontWeight}
)

// problemView explains, why the setup cannot be done. It names only the reasons, never the address, accounts,
// servers or tokens.
func problemView(wnd core.Window, opts FlowOptions, st onboarding.State) core.View {
	noDelivery := st.Delivery.Level == onboarding.DeliveryNone

	var lines []core.View
	if noDelivery {
		lines = append(lines,
			ui.Text(StrNoDelivery.Get(wnd)).Font(bannerLeadFont),
			ui.Text(StrNoDeliveryCode.Get(wnd)).Font(bannerFont),
		)
	}

	if st.Problem != onboarding.NoProblem {
		lines = append(lines, ui.Text(StrNoAddress.Get(wnd)).Font(bannerLeadFont))
	}

	lines = append(lines, ui.Text(StrContactOperators.Get(wnd)).Font(bannerFont))

	views := []core.View{banner(alert.IntentWarning, IDProblem, lines...)}

	failed, next := st.Delivery.ExchangeFailedAt, st.Delivery.NextExchangeAt
	if noDelivery && !failed.IsZero() && !next.IsZero() {
		views = append(views, ui.HStack(
			ui.Text(StrExchangeFailedX.Get(wnd, i18n.String("time", clock(failed)), i18n.String("next", clock(next)))).
				Font(ui.BodySmall).
				Color(ui.ColorIconsMuted),
		).Alignment(ui.Leading).FullWidth().ID(IDExchangeFailed))
	}

	var hints []core.View
	for _, hint := range []string{
		StrHintEmailX.Get(wnd, i18n.String("env", cmp.Or(opts.EmailEnv, onboarding.EnvInitialUserEmail))),
		StrHintMailServiceX.Get(wnd, i18n.String("env", opts.MailServiceEnv)),
		StrHintSMTP.Get(wnd),
	} {
		hints = append(hints, ui.Text("• "+hint).Font(ui.MonoSmall))
	}

	views = append(views, ui.VStack(
		accordion.Accordion(
			ui.Text(StrForOperators.Get(wnd)).Font(ui.BodySmall),
			ui.VStack(hints...).Alignment(ui.Leading).Gap(ui.L4).FullWidth(),
			core.StateOf[bool](wnd, "nago.onboarding.operators"),
		).HideSeparator().FullWidth(),
	).FullWidth().ID(IDOperators))

	return ui.VStack(views...).Alignment(ui.Leading).Gap(ui.L12).FullWidth()
}

// clock is the time of day on the server.
func clock(t time.Time) string {
	return t.In(time.Local).Format("15:04")
}

// codeStatusView is the status line of the mail with the code. It shows only the status: no content, no address, no
// server. An unknown status has no line.
func codeStatusView(wnd core.Window, status onboarding.CodeStatus) core.View {
	var dot ui.Color
	var text string
	switch status {
	case onboarding.CodeStatusQueued:
		dot, text = ui.SV0, StrCodeStatusQueued.Get(wnd)
	case onboarding.CodeStatusAccepted:
		dot, text = ui.SG0, StrCodeStatusAccepted.Get(wnd)
	case onboarding.CodeStatusRetrying:
		dot, text = ui.SW0, StrCodeStatusRetrying.Get(wnd)
	case onboarding.CodeStatusFailed:
		dot, text = ui.SE0, StrCodeStatusFailed.Get(wnd)
	case onboarding.CodeStatusInLog:
		dot, text = "", StrCodeStatusInLog.Get(wnd)
	default:
		return nil
	}

	return ui.HStack(tags.StatusBadge(dot, text)).Alignment(ui.Leading).FullWidth().ID(IDCodeStatus)
}

// legalLinks are the links to the legal texts below the card, each one only with a value.
func legalLinks(wnd core.Window, legal Legal) core.View {
	var links []core.View
	add := func(text, href string) {
		if href == "" {
			return
		}

		if len(links) > 0 {
			links = append(links, ui.Text("·").Color(ui.ColorIconsMuted))
		}

		links = append(links, legalLink(wnd, text, href))
	}

	add(StrImpress.Get(wnd), legal.Impress)
	add(StrPrivacyPolicy.Get(wnd), legal.PrivacyPolicy)
	if len(links) == 0 {
		return nil
	}

	return ui.HStack(links...).Gap(ui.L8).Font(ui.BodySmall).ID(IDLegal)
}

// legalLink opens an external address in a new tab and navigates to a route of the application, which the setup
// exempts, see cfgonboarding.Options.Exempt.
func legalLink(wnd core.Window, text, href string) ui.TText {
	if strings.Contains(href, "://") {
		return ui.Link(wnd, text, href, ui.LinkTargetNewWindowOrTab)
	}

	// absolute, because a relative path resolves against the current route in the browser
	return ui.Link(wnd, text, "/"+strings.TrimPrefix(href, "/"), "")
}

// banner shows the lines as written in the colors of the alert banner, which would upper-case a title.
func banner(intent alert.Intent, id string, lines ...core.View) core.View {
	var fg, bg ui.Color
	var icon core.SVG
	switch intent {
	case alert.IntentOk:
		fg, bg, icon = ui.ColorBannerInfoText, ui.ColorBannerInfoBackground, core.SVG(alert.IconOK)
	case alert.IntentWarning:
		fg, bg, icon = ui.ColorBannerWarningText, ui.ColorBannerWarningBackground, core.SVG(alert.IconWarning)
	default:
		fg, bg, icon = ui.ColorBannerErrorText, ui.ColorBannerErrorBackground, core.SVG(alert.IconError)
	}

	for i, line := range lines {
		if text, ok := line.(ui.TText); ok {
			lines[i] = text.Color(fg)
		}
	}

	return ui.HStack(
		ui.Image().FillColor(fg).Embed(icon).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
		ui.VStack(lines...).Alignment(ui.Leading).Gap(ui.L4).FullWidth(),
	).Alignment(ui.TopLeading).
		Gap(ui.L12).
		FullWidth().
		ID(id).
		BackgroundColor(bg).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L12))
}

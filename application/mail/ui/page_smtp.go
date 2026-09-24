// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import (
	"fmt"
	"slices"
	"time"

	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/tags"
)

// SmtpPage shows the configured smtp servers and their health.
func SmtpPage(wnd core.Window, pages Pages, uc mail.UseCases) core.View {
	now := time.Now()
	stats, err := uc.Statistics(wnd.Subject(), mail.StatisticsOptions{From: now.Add(-24 * time.Hour), To: now})
	if err != nil {
		return alert.BannerError(err)
	}

	var trailing []core.View
	if pages.SecretVault != "" {
		trailing = append(trailing, ui.PrimaryButton(func() {
			wnd.Navigation().ForwardTo(pages.SecretVault, nil)
		}).Title("Server im Tresor verwalten"))
	}

	var cards []core.View
	for _, srv := range stats.Servers {
		cards = append(cards, serverCard(wnd, pages, srv, slices.Contains(stats.Scheduler.RateLimited, srv.Name)))
	}

	return ui.VStack(
		header(wnd, pages, tabSmtp, "SMTP-Server", []crumb{{title: "SMTP-Server"}}, trailing...),
		ui.IfFunc(len(stats.Servers) == 0, func() core.View {
			return alert.Banner("Kein SMTP-Server", "Es ist kein SMTP-Server mit der Systemgruppe geteilt. Legen Sie im Tresor ein SMTP-Secret an und teilen Sie es mit der Gruppe System.").Intent(alert.IntentError).Frame(ui.Frame{}.FullWidth())
		}),
		ui.Text("Der Scheduler verwendet den Server, dessen Name oder ID dem Server-Hinweis der Mail entspricht, ansonsten den ersten gefundenen Server.").Font(ui.BodySmall).Color(ui.ST0),
		grid([5]int{1, 2, 2, 3, 3}, cards...),
	).Alignment(ui.TopLeading).Gap(ui.L16).FullWidth()
}

func serverCard(wnd core.Window, pages Pages, srv mail.ServerInfo, rateLimited bool) core.View {
	h := srv.Health
	var state core.View
	switch {
	case h.ConsecutiveFailures >= 3:
		state = tags.StatusBadge(ui.SE0, "Gestört")
	case h.ConsecutiveFailures > 0:
		state = tags.StatusBadge(ui.SW0, "Fehler")
	case rateLimited:
		state = tags.StatusBadge(ui.SW0, "Ratenbegrenzt")
	case h.LastSuccessAt.IsZero():
		state = tags.StatusBadge(ui.ST0, "Unbenutzt")
	default:
		state = tags.StatusBadge(ui.SG0, "Gesund")
	}

	return card("",
		ui.HStack(ui.Text(srv.Name).Font(ui.TitleLarge), ui.Spacer(), state).FullWidth(),
		ui.Text(fmt.Sprintf("%s:%d", srv.Host, srv.Port)).Font(ui.MonoSmall).Color(ui.ST0),
		kvTable(
			kvText("Letzte 24 h", fmt.Sprintf("%d versendet · %d Fehlversuche", srv.Totals.Sent, srv.Totals.Failed)),
			kvText("Zuletzt erfolgreich", formatTime(wnd, h.LastSuccessAt)),
			kvText("Letzter Fehler", formatTime(wnd, h.LastErrorAt)),
			kvText("Fehler in Folge", fmt.Sprint(h.ConsecutiveFailures)),
			kvText("Ratenbegrenzung", rateLimitText(srv.RateLimitPerHour, srv.RateLimitPerDay)),
		),
		ui.IfFunc(h.ConsecutiveFailures > 0 && h.LastError != "", func() core.View {
			return ui.VStack(
				ui.Text(fmt.Sprintf("%s: %s", phaseLabel(h.LastErrorPhase), h.LastError)).Font(ui.MonoSmall),
				ui.IfFunc(errorHint(h.LastErrorPhase, h.LastErrorCode, h.LastError) != "", func() core.View {
					return ui.Text(errorHint(h.LastErrorPhase, h.LastErrorCode, h.LastError)).Font(ui.BodySmall)
				}),
			).Alignment(ui.Leading).Gap(ui.L4).BackgroundColor(ui.M1).Border(ui.Border{}.Radius(ui.L8)).Padding(ui.Padding{}.All(ui.L8))
		}),
		ui.HStack(
			ui.IfFunc(pages.SecretEdit != "", func() core.View {
				return ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(pages.SecretEdit, core.Values{"id": string(srv.SecretID)})
				}).Title("Bearbeiten")
			}),
			ui.SecondaryButton(func() {
				wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"server": srv.Name})
			}).Title("Mails anzeigen"),
			ui.IfFunc(pages.SendMailTest != "", func() core.View {
				return ui.TertiaryButton(func() {
					wnd.Navigation().ForwardTo(pages.SendMailTest, core.Values{"smtp": srv.Name})
				}).Title("Testmail")
			}),
		).Gap(ui.L8).Wrap(true),
	).Frame(ui.Frame{}.FullWidth())
}

func rateLimitText(perHour, perDay int) string {
	switch {
	case perHour <= 0 && perDay <= 0:
		return "unbegrenzt"
	case perDay <= 0:
		return fmt.Sprintf("%d pro Stunde", perHour)
	case perHour <= 0:
		return fmt.Sprintf("%d pro Tag", perDay)
	default:
		return fmt.Sprintf("%d pro Stunde, %d pro Tag", perHour, perDay)
	}
}

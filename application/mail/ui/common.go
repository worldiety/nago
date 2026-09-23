// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import (
	"fmt"
	"strings"
	"time"

	"github.com/worldiety/i18n/date"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/tags"
)

type navTab int

const (
	tabDashboard navTab = iota
	tabQueue
	tabSmtp
	tabTemplates
	tabTest
)

// header renders breadcrumbs, the title and the navigation bar of the mail area.
func header(wnd core.Window, pages Pages, active navTab, title string, crumbs []crumb, trailing ...core.View) core.View {
	bc := breadcrumb.Breadcrumbs().ClampLeading()
	bc = bc.Item("E-Mail", func() { wnd.Navigation().ForwardTo(pages.Dashboard, nil) })
	for _, c := range crumbs {
		bc = bc.Item(c.title, c.action)
	}

	type entry struct {
		tab   navTab
		title string
		path  core.NavigationPath
		query core.Values
	}

	entries := []entry{
		{tabDashboard, "Übersicht", pages.Dashboard, nil},
		{tabQueue, "Warteschlange", pages.OutgoingMailQueue, nil},
		{tabSmtp, "SMTP-Server", pages.SmtpServers, nil},
		{tabTemplates, "Vorlagen", pages.Templates, core.Values{"tag": "mail"}},
		{tabTest, "Test", pages.SendMailTest, nil},
	}

	var navButtons []core.View
	for _, e := range entries {
		if e.path == "" {
			continue
		}

		action := func() { wnd.Navigation().ForwardTo(e.path, e.query) }
		if e.tab == active {
			navButtons = append(navButtons, ui.PrimaryButton(action).Title(e.title))
		} else {
			navButtons = append(navButtons, ui.TertiaryButton(action).Title(e.title))
		}
	}

	return ui.VStack(
		bc,
		ui.HStack(
			ui.H1(title),
			ui.Spacer(),
		).Append(trailing...).Gap(ui.L8).Wrap(true).FullWidth(),
		ui.HStack(navButtons...).Gap(ui.L4).Wrap(true).Alignment(ui.Leading),
		ui.HLine(),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

type crumb struct {
	title  string
	action func()
}

func statusPill(status mail.Status) core.View {
	switch status {
	case mail.StatusSendSuccess:
		return tags.ColoredTextPill(ui.SG0, "Versendet")
	case mail.StatusError:
		return tags.ColoredTextPill(ui.SW0, "Fehler, wird wiederholt")
	case mail.StatusFailed:
		return tags.ColoredTextPill(ui.SE0, "Endgültig fehlgeschlagen")
	case mail.StatusQueued:
		return tags.ColoredTextPill(ui.SV0, "Wartet")
	default:
		return tags.ColoredTextPill(ui.ST0, "Unbekannt")
	}
}

func formatTime(wnd core.Window, t time.Time) string {
	if t.IsZero() {
		return "–"
	}

	return date.Format(wnd.Locale(), date.TimeMinute, t.In(wnd.Location()))
}

func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "–"
	case d < time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.0f s", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%.0f min", d.Minutes())
	default:
		if d%time.Hour == 0 {
			return fmt.Sprintf("%d h", int(d.Hours()))
		}
		return fmt.Sprintf("%.1f h", d.Hours())
	}
}

func phaseLabel(p mail.Phase) string {
	switch p {
	case mail.PhaseConfig:
		return "Konfiguration"
	case mail.PhaseDial:
		return "Verbindung"
	case mail.PhaseTLS:
		return "TLS"
	case mail.PhaseAuth:
		return "Anmeldung"
	case mail.PhaseMail:
		return "Absender"
	case mail.PhaseRcpt:
		return "Empfänger"
	case mail.PhaseData:
		return "Daten"
	case mail.PhaseQuit:
		return "Abschluss"
	default:
		return "Unbekannt"
	}
}

// errorHint returns a human-readable diagnosis for typical smtp errors.
func errorHint(phase mail.Phase, code int, msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case phase == mail.PhaseAuth || code == 535 || code == 534:
		return "Die Anmeldung am SMTP-Server ist fehlgeschlagen. Prüfen Sie Benutzername und Passwort des SMTP-Servers."
	case phase == mail.PhaseDial || strings.Contains(lower, "timeout") || strings.Contains(lower, "connection refused"):
		return "Der SMTP-Server ist nicht erreichbar. Prüfen Sie Host, Port und Firewall."
	case phase == mail.PhaseTLS || strings.Contains(lower, "certificate"):
		return "Die TLS-Verbindung konnte nicht aufgebaut werden. Prüfen Sie Port und Zertifikat des Servers."
	case phase == mail.PhaseMail:
		return "Der Server lehnt die Absenderadresse ab. Prüfen Sie die Absenderadresse des SMTP-Servers."
	case phase == mail.PhaseRcpt && code >= 500:
		return "Der Server lehnt den Empfänger dauerhaft ab, z.B. weil das Postfach nicht existiert."
	case code >= 400 && code < 500:
		return "Temporärer Fehler (z.B. Greylisting oder Rate-Limit). Der Versand wird automatisch wiederholt."
	default:
		return ""
	}
}

func card(title string, body ...core.View) ui.DecoredView {
	return ui.VStack(
		ui.IfElse(title == "", nil, ui.Text(title).Font(ui.TitleMedium)),
	).Append(body...).
		Alignment(ui.TopLeading).
		Gap(ui.L12).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Padding(ui.Padding{}.All(ui.L16))
}

func kpi(title, value string, valueColor ui.Color, hint string, hintColor ui.Color) core.View {
	v := ui.Text(value).Font(ui.HeadlineMedium)
	if valueColor != "" {
		v = v.Color(valueColor)
	}

	h := ui.Text(hint).Font(ui.BodySmall)
	if hintColor != "" {
		h = h.Color(hintColor)
	}

	return card("", ui.Text(title).Font(ui.LabelLarge), v, h).Frame(ui.Frame{MinWidth: ui.L160}.FullWidth())
}

// grid creates a card layout with explicit columns for all size classes: small, medium, large, xl and 2xl.
func grid(cols [5]int, children ...core.View) core.View {
	classes := []core.WindowSizeClass{core.SizeClassSmall, core.SizeClassMedium, core.SizeClassLarge, core.SizeClassXL, core.SizeClass2XL}
	l := cardlayout.Layout(children...)
	for i, c := range classes {
		l = l.Columns(c, cols[i])
	}

	return l.Frame(ui.Frame{}.FullWidth())
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"

	mail2 "go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/tabs"
	"go.wdy.de/nago/presentation/ui/tags"
)

// DetailPage shows a single outgoing mail identified by the query parameter id.
func DetailPage(wnd core.Window, pages Pages, uc mail2.UseCases) core.View {
	id := mail2.ID(wnd.Values()["id"])
	optOut, err := uc.FindOutgoingByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	crumbs := []crumb{{title: "Warteschlange", action: func() { wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, nil) }}}
	if optOut.IsNone() {
		return ui.VStack(
			header(wnd, pages, tabQueue, "Mail nicht gefunden", crumbs),
			alert.Banner("Nicht gefunden", "Die Mail existiert nicht mehr. Erfolgreich versendete Mails werden nach einiger Zeit automatisch entfernt.").Intent(alert.IntentWarning),
		).Alignment(ui.TopLeading).Gap(ui.L16).FullWidth()
	}

	out := optOut.Unwrap()
	title := out.Subject
	if title == "" {
		title = "(kein Betreff)"
	}

	deletePresented := core.AutoState[bool](wnd)

	actions := ui.HStack(
		statusPill(out.Status),
		ui.IfFunc(out.Status != mail2.StatusSendSuccess, func() core.View {
			return ui.PrimaryButton(func() {
				if err := uc.RetryOutgoing(wnd.Subject(), out.ID); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				alert.ShowBannerMessage(wnd, alert.Message{Title: "Erneut eingereiht", Message: "Die Mail wird beim nächsten Durchlauf des Schedulers versendet.", Intent: alert.IntentOk})
			}).PreIcon(heroOutline.ArrowPath).Title("Erneut versuchen")
		}),
		ui.SecondaryButton(func() {
			newID, err := uc.ResendOutgoing(wnd.Subject(), out.ID)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			wnd.Navigation().ForwardTo(pages.OutgoingMail, core.Values{"id": string(newID)})
		}).PreIcon(heroOutline.PaperAirplane).Title("Als neue Mail senden"),
		ui.SecondaryButton(func() {
			deletePresented.Set(true)
		}).PreIcon(heroOutline.Trash).Title("Löschen"),
	).Gap(ui.L8).Wrap(true)

	return ui.VStack(
		alert.Dialog("Mail löschen", ui.Text("Soll die Mail wirklich aus der Warteschlange gelöscht werden?"), deletePresented, alert.Cancel(nil), alert.Delete(func() {
			if err := uc.DeleteOutgoingByID(wnd.Subject(), out.ID); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, nil)
		})),
		header(wnd, pages, tabQueue, title, append(crumbs, crumb{title: title}), actions),
		ui.IfFunc(out.Status == mail2.StatusSuppressed, func() core.View {
			return alert.Banner("Vom Spam-Schutz unterdrückt", out.LastError+". Prüfen Sie, ob die Anwendung diese Mail in einer Schleife erzeugt. „Erneut versuchen“ versendet die Mail trotzdem.").Intent(alert.IntentError).Frame(ui.Frame{}.FullWidth())
		}),
		grid([5]int{1, 1, 2, 2, 2},
			card("Metadaten", metadata(wnd, out)).Frame(ui.Frame{}.FullWidth()),
			card("Versuche", attempts(wnd, pages, out)).Frame(ui.Frame{}.FullWidth()),
		),
		card("Inhalt", content(wnd, out)).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.TopLeading).Gap(ui.L16).FullWidth()
}

func addresses(list []mail.Address) string {
	if len(list) == 0 {
		return "–"
	}

	var tmp []string
	for _, a := range list {
		tmp = append(tmp, a.String())
	}

	return strings.Join(tmp, ", ")
}

func metadata(wnd core.Window, out mail2.Outgoing) core.View {
	from := "Standard des SMTP-Servers"
	if out.Mail.From.Address != "" {
		from = out.Mail.From.String()
	}

	next := "–"
	if out.Status == mail2.StatusError || out.Status == mail2.StatusQueued {
		next = "so bald wie möglich"
		if !out.NextAttemptAt.IsZero() {
			next = formatTime(wnd, out.NextAttemptAt)
		}
	}

	return kvTable(
		kvText("An", addresses(out.Mail.To)),
		kvText("CC", addresses(out.Mail.CC)),
		kvText("BCC", addresses(out.Mail.BCC)),
		kvText("Von", from),
		kvText("Server-Hinweis", out.Mail.SmtpHint),
		kvText("Versendet über", out.ServerName),
		kvText("Eingereiht", formatTime(wnd, out.QueuedAt)),
		kvText("Versendet", formatTime(wnd, out.SentAt)),
		kvText("Versuche", fmt.Sprint(out.Attempted())),
		kvText("Nächster Versuch", next),
		kvText("ID", string(out.ID)),
	)
}

func attempts(wnd core.Window, pages Pages, out mail2.Outgoing) core.View {
	if len(out.Attempts) == 0 {
		if out.LastError != "" {
			return ui.VStack(
				ui.Text("Für diese Mail wurde noch keine Versuchshistorie erfasst. Letzter Fehler:").Color(ui.ST0),
				ui.Text(out.LastError).Font(ui.MonoSmall),
			).Alignment(ui.Leading).Gap(ui.L8)
		}

		return ui.Text("Noch keine Versandversuche.").Color(ui.ST0)
	}

	list := slices.Clone(out.Attempts)
	slices.Reverse(list)
	offset := out.AttemptCount

	var rows []core.View
	for i, a := range list {
		nr := offset - i
		var pill core.View
		if a.Success {
			pill = tags.ColoredTextPill(ui.SG0, "Erfolgreich")
		} else {
			txt := phaseLabel(a.Phase)
			if a.Code > 0 {
				txt = fmt.Sprintf("%s · %d", txt, a.Code)
			}
			pill = tags.ColoredTextPill(ui.SE0, txt)
		}

		rows = append(rows, ui.VStack(
			ui.HStack(
				ui.Text(fmt.Sprintf("#%d · %s", nr, formatTime(wnd, a.At))).Font(ui.TitleSmall),
				ui.Text(a.Server).Color(ui.ST0),
				pill,
				ui.Spacer(),
				ui.Text(formatDuration(a.Duration)).Font(ui.BodySmall).Color(ui.ST0),
			).Gap(ui.L8).Wrap(true).FullWidth(),
			ui.IfElse(a.Message == "", nil, ui.Text(a.Message).Font(ui.MonoSmall)),
		).Alignment(ui.Leading).Gap(ui.L4).FullWidth())
	}

	if last, ok := out.LastAttempt(); ok && !last.Success {
		if hint := errorHint(last.Phase, last.Code, last.Message); hint != "" {
			rows = append(rows, alert.Banner("Diagnose", hint).Intent(alert.IntentWarning))
			if pages.SmtpServers != "" && (last.Phase == mail2.PhaseAuth || last.Phase == mail2.PhaseDial || last.Phase == mail2.PhaseTLS || last.Phase == mail2.PhaseMail) {
				rows = append(rows, ui.TertiaryButton(func() {
					wnd.Navigation().ForwardTo(pages.SmtpServers, nil)
				}).Title("SMTP-Server prüfen"))
			}
		}
	}

	return ui.VStack(rows...).Alignment(ui.Leading).Gap(ui.L12).FullWidth()
}

type partView struct {
	kind  string // html, text, attachment
	name  string
	value string
	size  int
}

func headerValue(h mail.Header, key string) string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			return strings.Join(v, ";")
		}
	}

	return ""
}

func classifyParts(out mail2.Outgoing) []partView {
	var res []partView
	for _, p := range out.Mail.Parts {
		ct := strings.ToLower(headerValue(p.Header, "Content-Type"))
		switch {
		case strings.Contains(ct, "text/html"):
			res = append(res, partView{kind: "html", value: string(p.Encoded)})
		case strings.Contains(ct, "text/plain"):
			res = append(res, partView{kind: "text", value: string(p.Encoded)})
		default:
			name := "Anhang"
			for _, v := range strings.Split(headerValue(p.Header, "Content-Type"), ";") {
				if n, ok := strings.CutPrefix(strings.TrimSpace(v), "name="); ok {
					name = n
				}
			}
			res = append(res, partView{kind: "attachment", name: name, size: len(p.Encoded) * 3 / 4})
		}
	}

	return res
}

func content(wnd core.Window, out mail2.Outgoing) core.View {
	parts := classifyParts(out)
	var pages []tabs.TPage
	var attachments []partView
	for _, p := range parts {
		switch p.kind {
		case "html":
			pages = append(pages, tabs.Page("HTML", func() core.View {
				return ui.VStack(ui.RichText(p.value).FullWidth()).
					Alignment(ui.TopLeading).
					BackgroundColor(ui.M1).
					Border(ui.Border{}.Radius(ui.L8)).
					Padding(ui.Padding{}.All(ui.L16)).
					Frame(ui.Frame{}.FullWidth())
			}))
			pages = append(pages, tabs.Page("HTML-Quelltext", func() core.View {
				return ui.CodeEditor(p.value).Disabled(true).Language("html").Frame(ui.Frame{Height: ui.L480}.FullWidth())
			}))
		case "text":
			pages = append(pages, tabs.Page("Text", func() core.View {
				return ui.CodeEditor(p.value).Disabled(true).Frame(ui.Frame{Height: ui.L480}.FullWidth())
			}))
		default:
			attachments = append(attachments, p)
		}
	}

	pages = append(pages, tabs.Page(fmt.Sprintf("Anhänge (%d)", len(attachments)), func() core.View {
		if len(attachments) == 0 {
			return ui.Text("Keine Anhänge.").Color(ui.ST0)
		}

		var rows []core.View
		for _, a := range attachments {
			rows = append(rows, ui.HStack(ui.Text(a.name), ui.Text(fmt.Sprintf("%.1f KB", float64(a.size)/1024)).Color(ui.ST0)).Gap(ui.L8))
		}

		return ui.VStack(rows...).Alignment(ui.Leading).Gap(ui.L8)
	}))

	if out.LastError != "" {
		pages = append(pages, tabs.Page("Rohfehler", func() core.View {
			return ui.CodeEditor(out.LastError).Disabled(true).Frame(ui.Frame{Height: ui.L200}.FullWidth())
		}))
	}

	return tabs.Tabs(pages...).InputValue(core.AutoState[int](wnd)).FullWidth()
}

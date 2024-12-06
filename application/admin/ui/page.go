package uiadmin

import (
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type Pages struct {
	Mail std.Option[uimail.MailPages]
}

func SettingsOverviewPage(wnd core.Window, pages Pages) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(iam.InvalidSubjectError("not logged in"))
	}

	query := core.AutoState[string](wnd)

	adminGroups := filter(wnd.Subject(), groups(pages), query.Get())

	var viewBuilder xslices.Builder[core.View]
	viewBuilder.Append(
		ui.H1("Admin Center"),

		ui.HStack(
			ui.TextField("", query.Get()).
				InputValue(query).
				Style(ui.TextFieldReduced),
		).Alignment(ui.Trailing).
			FullWidth(),
	)

	for _, grp := range adminGroups {
		viewBuilder.Append(ui.H2(grp.Title))
		var cardLayoutViews xslices.Builder[core.View]
		for i, entry := range grp.Entries {
			cardLayoutViews.Append(
				cardlayout.Card(entry.Title).
					Body(ui.Text(entry.Text)).
					Footer(
						ui.IfElse(i == 0,
							ui.PrimaryButton(func() {
								wnd.Navigation().ForwardTo(entry.Target, nil)
							}).Title("Auswählen"),
							ui.SecondaryButton(func() {
								wnd.Navigation().ForwardTo(entry.Target, nil)
							}).Title("Auswählen"),
						),
					),
			)
		}

		viewBuilder.Append(
			cardlayout.Layout(cardLayoutViews.Collect()...),
		)

	}

	return ui.VStack(
		viewBuilder.Collect()...,
	).FullWidth().Alignment(ui.Leading)

}

type DashboardModel struct {
	Title  string
	Text   string
	Target core.NavigationPath
	Role   auth.RID
}

type Group struct {
	Title   string
	Entries []DashboardModel
}

func groups(pages Pages) []Group {
	var grps []Group
	if pages.Mail.IsSome() {
		pages := pages.Mail.Unwrap()
		grps = append(grps, Group{
			Title: "eMail und SMTP",
			Entries: []DashboardModel{
				{
					Title:  "SMTP",
					Text:   "Das System unterstützt verschiedene EMail-Ausgangsserver. Ein Ausgangsserver ist z.B. für die Self-Service Funktionen der Nutzer erforderlich.",
					Target: pages.SMTPServer,
				},
				{
					Title:  "Warteschlange",
					Text:   "E-Mails werden über eine Postausgangs-Warteschlange versendet.",
					Target: pages.OutgoingMailQueue,
				},
				{
					Title:  "Vorlagen",
					Text:   "Hierüber kann die aktuelle Mail-Server Konfiguration inkl. Templating und co. getestet werden.",
					Target: pages.Templates,
				},
				{
					Title:  "Scheduler",
					Text:   "Der Mail Scheduler bearbeitet die Warteschlange des Postausgangs und bietet ebenfalls ein paar Einstelloptionen.",
					Target: pages.MailScheduler,
				},
				{
					Title:  "Test",
					Text:   "Hierüber kann die aktuelle Mail-Server Konfiguration inkl. Templating und co. getestet werden.",
					Target: pages.SendMailTest,
				},
			},
		})
	}

	return grps
}

func filter(subject auth.Subject, groups []Group, text string) []Group {
	// TODO implement role filter
	var res []Group

	predicate := rquery.SimplePredicate[string](text)

	for _, group := range groups {
		fgrp := Group{
			Title: group.Title,
		}
		for _, entry := range group.Entries {
			if text != "" {
				if predicate(entry.Title) || predicate(entry.Text) {
					fgrp.Entries = append(fgrp.Entries, entry)
				}
			} else {
				fgrp.Entries = append(fgrp.Entries, entry)
			}

		}

		if len(fgrp.Entries) > 0 {
			res = append(res, fgrp)
		}
	}

	return res
}

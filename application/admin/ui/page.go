package uiadmin

import (
	"go.wdy.de/nago/application/group"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type IAMPages struct {
	Accounts    core.NavigationPath
	Permissions core.NavigationPath
	Groups      core.NavigationPath
	Roles       core.NavigationPath
}

type Pages struct {
	Mail      std.Option[uimail.Pages]
	IAM       std.Option[IAMPages]
	Dashboard core.NavigationPath
}

func SettingsOverviewPage(wnd core.Window, pages Pages) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
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
	Title      string
	Text       string
	Target     core.NavigationPath
	Role       role.ID
	Permission permission.ID
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

	if pages.IAM.IsSome() {
		pages := pages.IAM.Unwrap()
		grps = append(grps, Group{
			Title: "Nutzerverwaltung",
			Entries: []DashboardModel{
				{
					Title:      "Konten",
					Text:       "Über die Kontenverwaltung können die einzelnen bekannten Identitäten der Nutzer verwaltet werden. Hierüber können Rollen, Gruppen und Einzelberechtigungen einem Individuum zugeordnet werden.",
					Target:     pages.Accounts,
					Permission: group.PermFindAll,
				},
				{
					Title:      "Rollen",
					Text:       "Über die Rollenverwaltung können einzelne Berechtigungen in einer Rolle zusammengefasst werden. Dies ist die empfohlene Art einem Konto eine Menge an Berechtigungen zuzuteilen. Die konkreten Rollen ergeben sich aus der Domäne.",
					Target:     pages.Roles,
					Permission: role.PermFindAll,
				},
				{
					Title:      "Gruppen",
					Text:       "Mittels der Gruppenverwaltung können Nutzer in Gruppen organisiert werden. Dieses Szenario wird typischerweise genutzt, um Nutzergruppen dynamisch und unabhängig von ihren Rollen zu organisieren. Damit dies Sinn macht, muss die Domäne auch Gruppen unterstützen.",
					Target:     pages.Groups,
					Permission: role.PermFindAll,
				},
				{
					Title:      "Berechtigungen",
					Text:       "Jeder in der Domäne modellierte Anwendungsfall hat eine individuelle Berechtigung, sodass im Zweifel jedes Konto mit feingranularen Berechtigungen ausgestattet werden kann. Die Anwendungsfälle werden zur Entwicklungszeit festgelegt und können daher nicht editiert werden.",
					Target:     pages.Permissions,
					Permission: role.PermFindAll,
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

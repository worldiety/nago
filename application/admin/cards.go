package admin

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
)

func DefaultGroups(pages Pages) []Group {
	var grps []Group
	grps = append(grps, Group{
		Title: "Nutzerverwaltung",
		Entries: []Card{
			{
				Title:      "Konten",
				Text:       "Über die Kontenverwaltung können die einzelnen bekannten Identitäten der Nutzer verwaltet werden. Hierüber können Rollen, Gruppen und Einzelberechtigungen einem Individuum zugeordnet werden.",
				Target:     pages.User.Users,
				Permission: user.PermFindAll,
			},
			{
				Title:      "Rollen",
				Text:       "Über die Rollenverwaltung können einzelne Berechtigungen in einer Rolle zusammengefasst werden. Dies ist die empfohlene Art einem Konto eine Menge an Berechtigungen zuzuteilen. Die konkreten Rollen ergeben sich aus der Domäne.",
				Target:     pages.Role.Roles,
				Permission: role.PermFindAll,
			},
			{
				Title:      "Gruppen",
				Text:       "Mittels der Gruppenverwaltung können Nutzer in Gruppen organisiert werden. Dieses Szenario wird typischerweise genutzt, um Nutzergruppen dynamisch und unabhängig von ihren Rollen zu organisieren. Damit dies Sinn macht, muss die Domäne auch Gruppen unterstützen.",
				Target:     pages.Group.Groups,
				Permission: group.PermFindAll,
			},
			{
				Title:      "Berechtigungen",
				Text:       "Jeder in der Domäne modellierte Anwendungsfall hat eine individuelle Berechtigung, sodass im Zweifel jedes Konto mit feingranularen Berechtigungen ausgestattet werden kann. Die Anwendungsfälle werden zur Entwicklungszeit festgelegt und können daher nicht editiert werden.",
				Target:     pages.Permission.Permissions,
				Permission: permission.PermFindAll,
			},
		},
	})

	grps = append(grps, Group{
		Title: "eMail und SMTP",
		Entries: []Card{
			{
				Title:      "SMTP",
				Text:       "Das System unterstützt verschiedene EMail-Ausgangsserver. Ein Ausgangsserver ist z.B. für die Self-Service Funktionen der Nutzer erforderlich.",
				Target:     pages.Mail.SMTPServer,
				Permission: mail.PermFindAllSmtp,
			},
			{
				Title:      "Warteschlange",
				Text:       "E-Mails werden über eine Postausgangs-Warteschlange versendet.",
				Target:     pages.Mail.OutgoingMailQueue,
				Permission: mail.PermFindAllOutgoing,
			},
			{
				Title:  "Vorlagen",
				Text:   "Hierüber kann die aktuelle Mail-Server Konfiguration inkl. Templating und co. getestet werden.",
				Target: pages.Mail.Templates,
				// TODO
			},
			{
				Title:  "Scheduler",
				Text:   "Der Mail Scheduler bearbeitet die Warteschlange des Postausgangs und bietet ebenfalls ein paar Einstelloptionen.",
				Target: pages.Mail.MailScheduler,
				// TODO
			},
			{
				Title:      "Test",
				Text:       "Hierüber kann die aktuelle Mail-Server Konfiguration inkl. Templating und co. getestet werden.",
				Target:     pages.Mail.SendMailTest,
				Permission: mail.PermSendMail,
			},
		},
	})

	return grps
}

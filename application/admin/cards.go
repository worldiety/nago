package admin

import (
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
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
				Title:      "Warteschlange",
				Text:       "E-Mails werden über eine Postausgangs-Warteschlange versendet.",
				Target:     pages.Mail.OutgoingMailQueue,
				Permission: mail.PermOutgoingFindAll,
			},
			/*	{
				Title:  "Scheduler",
				Text:   "Der Mail Scheduler bearbeitet die Warteschlange des Postausgangs und bietet ebenfalls ein paar Einstelloptionen.",
				Target: pages.Mail.MailScheduler,
				Permission: mail.PermScheduler,
			},*/
			{
				Title:      "Test",
				Text:       "Hierüber kann die aktuelle Mail-Server Konfiguration inkl. Templating und co. getestet werden.",
				Target:     pages.Mail.SendMailTest,
				Permission: mail.PermSendMail,
			},
		},
	})

	grps = append(grps, Group{
		Title: "Vorlagen & Templates",
		Entries: []Card{
			{
				Title:        "E-Mail Vorlagen",
				Text:         "Alle E-Mail Vorlagen sichten und bearbeiten.",
				Target:       pages.Template.Projects,
				TargetParams: core.Values{"tag": string(template.TagMail)},
				Permission:   template.PermFindAll,
			},
			{
				Title:        "PDF Vorlagen",
				Text:         "Alle PDF Vorlagen sichten und bearbeiten.",
				Target:       pages.Template.Projects,
				TargetParams: core.Values{"tag": string(template.TagPDF)},
				Permission:   template.PermFindAll,
			},
			{
				Title:      "Alle Vorlagen",
				Text:       "Alle Projekte, Templates und Vorlagen durchsuchen oder bearbeiten.",
				Target:     pages.Template.Projects,
				Permission: template.PermFindAll,
			},
		},
	})

	grps = append(grps, Group{
		Title: "Lizenzen",
		Entries: []Card{
			{
				Title:      "Nutzer-Lizenzen",
				Text:       "Das System unterstützt verschiedene Lizenztypen. Eine Nutzer-Lizenz ist eine mengenlimitierte Lizenz und wird einzelnen Konten bis zur Obergrenze zugewiesen.",
				Target:     pages.License.UserLicenses,
				Permission: license.PermFindAllUserLicenses,
			},
			{
				Title:      "Modul-Lizenzen",
				Text:       "Das System unterstützt verschiedene Lizenztypen. Eine Modul bzw. Anwendungs-Lizenz ist eine statische Lizenz und gilt für die gesamte Anwendungsinstanz, sobald sie aktiv ist.",
				Target:     pages.License.AppLicenses,
				Permission: license.PermFindAllAppLicenses,
			},
		},
	})

	grps = append(grps, Group{
		Title: "Abrechnung",
		Entries: []Card{
			{
				Title:      "Lizensierte Module",
				Text:       "Darstellung der grundsätzlich verfügbaren und tatsächlich gebuchten Anwendungs- bzw. Applikationslizenzen zum Zweck der Abrechnung.",
				Target:     pages.Billing.AppLicenses,
				Permission: billing.PermAppLicenses,
			},
			{
				Title:      "Kontingente Nutzer-Lizenzen",
				Text:       "Darstellung der grundsätzlich verfügbaren und tatsächlich gebuchten Nutzerlizenzen zum Zweck der Abrechnung.",
				Target:     pages.Billing.UserLicenses,
				Permission: billing.PermUserLicenses,
			},
		}})

	grps = append(grps, Group{
		Title: "System",
		Entries: []Card{
			{
				Title:      "Backup und Wiederherstellung",
				Text:       "Die komplette Anwendung sichern und wiederherstellen.",
				Target:     pages.Backup.BackupAndRestore,
				Permission: backup.PermBackup,
			},
		}})

	grps = append(grps, Group{
		Title: "Tresor & Fremdsysteme",
		Entries: []Card{
			{
				Title:      "Tresor und Geheimnisverwaltung",
				Text:       "Verwaltung von Secrets, Geheimnissen und Zugangsdaten zu Fremdsystemen.",
				Target:     pages.Secret.Vault,
				Permission: secret.PermFindMySecrets,
			},
		}})

	return grps
}

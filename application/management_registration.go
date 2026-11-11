// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"log/slog"
	"math"
	netmail "net/mail"
	"time"

	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/presentation/core"
)

// sendRegistrationLink sends the link, which continues the self registration, to the given address.
func (c *Configurator) sendRegistrationLink(to user.Email, id user.RegistrationID, code string, validUntil time.Time) error {
	model := tplmail.RegistrationLinkModel{
		Email:           to,
		RegisterURL:     core.URI(c.ContextPathURI(string(c.userManagement.Pages.Register), core.Values{"id": string(id), "code": code})),
		ValidHours:      int(math.Ceil(time.Until(validUntil).Hours())),
		ApplicationName: c.applicationName,
	}

	return c.sendMailTemplateOrText(to, tplmail.RegistrationLinkSubject, tplmail.RegistrationLink, model,
		fmt.Sprintf("Konto erstellen bei %s", model.ApplicationName),
		fmt.Sprintf("Für %s wurde ein Konto mit dieser E-Mail-Adresse angefragt. Öffne den folgenden Link, um Deine Angaben zu ergänzen und das Konto zu erstellen:\n\n%s\n\nDer Link ist %d Stunden gültig. Falls Du kein Konto angefragt hast, kannst Du diese E-Mail ignorieren.",
			model.ApplicationName, model.RegisterURL, model.ValidHours),
	)
}

// sendAlreadyRegistered tells the owner of a registered address, that a registration has been requested for it.
func (c *Configurator) sendAlreadyRegistered(to user.Email) error {
	var login string
	if c.sessionManagement != nil {
		login = string(c.sessionManagement.Pages.Login)
	}

	model := tplmail.AlreadyRegisteredModel{
		Email:           to,
		LoginURL:        core.URI(c.ContextPathURI(login, nil)),
		ApplicationName: c.applicationName,
	}

	return c.sendMailTemplateOrText(to, tplmail.AlreadyRegisteredSubject, tplmail.AlreadyRegistered, model,
		fmt.Sprintf("Dein Konto bei %s", model.ApplicationName),
		fmt.Sprintf("Für %s wurde ein Konto mit dieser E-Mail-Adresse angefragt. Für diese Adresse gibt es bereits ein Konto, Du kannst Dich direkt anmelden:\n\n%s\n\nFalls Du Dein Kennwort vergessen hast, nutze auf der Anmeldeseite die Funktion „Passwort vergessen“. Falls Du kein Konto angefragt hast, kannst Du diese E-Mail ignorieren.",
			model.ApplicationName, model.LoginURL),
	)
}

// sendMailTemplateOrText sends the mail by the template of the system mail project. The template project of an
// instance, which has been created by an older version, lacks newer templates, so a plain mail is sent instead.
func (c *Configurator) sendMailTemplateOrText(to user.Email, subjName, bodyName template.DefinedTemplateName, model any, subject, text string) error {
	err := c.SendMailTemplate(to, tplmail.ID, subjName, bodyName, model)
	if err == nil {
		return nil
	}

	slog.Warn("cannot send mail by template, sending a plain mail", "template", bodyName, "err", err)
	mails, err := c.MailManagement()
	if err != nil {
		return err
	}

	_, err = mails.UseCases.SendMail(c.SysUser(), mail.Mail{
		To:      []netmail.Address{{Address: string(to)}},
		Subject: subject,
		Parts:   []mail.Part{mail.NewTextPart(text)},
	})

	return err
}

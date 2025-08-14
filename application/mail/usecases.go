// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"fmt"
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"log/slog"
)

var _ = enum.Variant[secret.Credentials, secret.SMTP]()

type Repository data.Repository[Outgoing, ID]

// SendMail takes the Mail and will try to publish it into either the given [Smtp] hint or whatever is currently defined
// as primary.
type SendMail func(subject auth.Subject, mail Mail) (ID, error)

type FindMailByID func(auth.Subject, ID) (std.Option[Outgoing], error)
type DeleteMailByID func(auth.Subject, ID) error
type FindAllMails func(auth.Subject) iter.Seq2[Outgoing, error]
type SaveMail func(auth.Subject, Outgoing) (ID, error)

type UseCases struct {
	Outgoing struct {
		FindByID   FindMailByID
		DeleteByID DeleteMailByID
		FindAll    FindAllMails
		Save       SaveMail
		repository Repository // intentionally not exposed, to avoid that devs can simply destroy invariants
	}

	SendMail SendMail
}

func NewUseCases(bus events.Bus, outgoingRepo Repository, ensureBuildIn template.EnsureBuildIn) (UseCases, error) {
	outgoingCrud := rcrud.DecorateRepository(rcrud.DecoratorOptions{EntityName: "Ausgehende Mails", PermissionPrefix: "nago.mail.outgoing"}, outgoingRepo)
	sendMailFn := NewSendMail(outgoingRepo)

	PermOutgoingFindAll = outgoingCrud.PermFindAll
	PermOutgoingDeleteByID = outgoingCrud.PermDeleteByID
	PermOutgoingFindByID = outgoingCrud.PermFindByID

	err := ensureBuildIn(user.NewSystem()(), template.NewProjectData{
		ID:          tplmail.ID,
		Name:        "Mailvorlagen Berechtigungssystem",
		Description: "Standardmailvorlagen für Nutzerregistrierung, Passwort vergessen, MFA Code und anderes.",
		ExecType:    template.TreeTemplateHTML,
		Tags:        []template.Tag{template.TagMail, template.TagHTML},
		Files:       tplmail.Files,
	}, false)

	if err != nil {
		return UseCases{}, fmt.Errorf("cannot ensure mail template: %w", err)
	}

	var uc UseCases
	uc.SendMail = sendMailFn

	uc.Outgoing.DeleteByID = outgoingCrud.DeleteByID
	uc.Outgoing.FindByID = outgoingCrud.FindByID
	uc.Outgoing.FindAll = outgoingCrud.FindAll
	uc.Outgoing.Save = outgoingCrud.Upsert

	events.SubscribeFor[SendMailRequested](bus, func(evt SendMailRequested) {
		var parts []Part
		if len(evt.TextBody) != 0 {
			parts = append(parts, NewTextPart(evt.TextBody))
		}

		if len(evt.HTMLBody) != 0 {
			parts = append(parts, NewHtmlPart(evt.HTMLBody))
		}

		for name, buf := range evt.Attachments {
			parts = append(parts, NewAttachmentPart(name, buf))
		}

		_, err := uc.SendMail(user.SU(), Mail{
			To:       evt.To,
			CC:       evt.CC,
			BCC:      evt.BCC,
			Subject:  evt.Subject,
			Parts:    parts,
			SmtpHint: evt.SmtpHint,
		})

		if err != nil {
			slog.Error("cannot send mail by SendMailRequested event", "err", err)
			return
		}
	})
	return uc, nil
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"fmt"
	"iter"
	"log/slog"
	"sync"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

var _ = enum.Variant[secret.Credentials, secret.SMTP]()

type Repository data.Repository[Outgoing, ID]

// SendMail takes the Mail and will try to publish it into either the given [Smtp] hint or whatever is currently defined
// as primary.
type SendMail func(subject auth.Subject, mail Mail) (ID, error)

// Deprecated: use [FindOutgoingByID].
type FindMailByID func(auth.Subject, ID) (std.Option[Outgoing], error)

// Deprecated: use [DeleteOutgoingByID].
type DeleteMailByID func(auth.Subject, ID) error

// Deprecated: use [FindOutgoingIDs] and [FindOutgoingByID].
type FindAllMails func(auth.Subject) iter.Seq2[Outgoing, error]

// Deprecated: the queue is managed by the scheduler, use [RetryOutgoing] or [ResendOutgoing].
type SaveMail func(auth.Subject, Outgoing) (ID, error)

type UseCases struct {
	// Deprecated: use the other use cases like [UseCases.FindOutgoingIDs].
	Outgoing struct {
		FindByID   FindMailByID
		DeleteByID DeleteMailByID
		FindAll    FindAllMails
		Save       SaveMail
	}

	SendMail           SendMail
	FindOutgoingIDs    FindOutgoingIDs
	FindOutgoingByID   FindOutgoingByID
	DeleteOutgoingByID DeleteOutgoingByID
	RetryOutgoing      RetryOutgoing
	ResendOutgoing     ResendOutgoing
	Statistics         Statistics
}

// Deprecated: use [NewUseCasesWithStats]. This variant has no statistics and no smtp server information.
func NewUseCases(bus events.Bus, outgoingRepo Repository, ensureBuildIn template.EnsureBuildIn, sysUser user.SysUser) (UseCases, error) {
	return NewUseCasesWithStats(bus, outgoingRepo, nil, nil, nil, nil, ensureBuildIn, sysUser)
}

// NewUseCasesWithStats creates the mail use cases. The stats, health, secrets and wakeup parameters are optional.
// The wakeup function is invoked whenever a mail becomes due, to trigger the scheduler immediately, see [NewWakeup].
func NewUseCasesWithStats(bus events.Bus, outgoingRepo Repository, stats StatsRepository, health HealthRepository, secrets secret.FindGroupSecrets, wakeup func(), ensureBuildIn template.EnsureBuildIn, sysUser user.SysUser) (UseCases, error) {
	if wakeup == nil {
		wakeup = func() {}
	}

	sendMailFn := SendMail(notifyOnSuccess[Mail](NewSendMail(outgoingRepo), wakeup))
	var mutex sync.Mutex

	err := ensureBuildIn(sysUser(), template.NewProjectData{
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
	uc.FindOutgoingIDs = NewFindOutgoingIDs(outgoingRepo)
	uc.FindOutgoingByID = NewFindOutgoingByID(outgoingRepo)
	uc.DeleteOutgoingByID = NewDeleteOutgoingByID(&mutex, outgoingRepo)
	retry := NewRetryOutgoing(&mutex, outgoingRepo)
	uc.RetryOutgoing = func(subject auth.Subject, ids ...ID) error {
		err := retry(subject, ids...)
		if err == nil {
			wakeup()
		}
		return err
	}
	uc.ResendOutgoing = notifyOnSuccess[ID](NewResendOutgoing(&mutex, outgoingRepo), wakeup)
	uc.Statistics = NewStatistics(outgoingRepo, stats, health, sysUser, secrets)

	// deprecated compatibility layer
	uc.Outgoing.FindByID = FindMailByID(uc.FindOutgoingByID)
	uc.Outgoing.DeleteByID = DeleteMailByID(uc.DeleteOutgoingByID)
	uc.Outgoing.FindAll = func(subject auth.Subject) iter.Seq2[Outgoing, error] {
		return func(yield func(Outgoing, error) bool) {
			if err := subject.Audit(PermOutgoingFindAll); err != nil {
				yield(Outgoing{}, err)
				return
			}

			for o, err := range outgoingRepo.All() {
				if !yield(o, err) {
					return
				}
			}
		}
	}
	uc.Outgoing.Save = func(subject auth.Subject, o Outgoing) (ID, error) {
		if err := subject.Audit(PermOutgoingUpdate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if o.ID == "" {
			o.ID = data.RandIdent[ID]()
		}

		if err := outgoingRepo.Save(o); err != nil {
			return "", err
		}

		wakeup()
		return o.ID, nil
	}

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
			To:        evt.To,
			CC:        evt.CC,
			BCC:       evt.BCC,
			Subject:   evt.Subject,
			Parts:     parts,
			SmtpHint:  evt.SmtpHint,
			InReplyTo: evt.InReplyTo,
		})

		if err != nil {
			slog.Error("cannot send mail by SendMailRequested event", "err", err)
			return
		}
	})
	return uc, nil
}

func notifyOnSuccess[T any](fn func(auth.Subject, T) (ID, error), wakeup func()) func(auth.Subject, T) (ID, error) {
	return func(subject auth.Subject, t T) (ID, error) {
		id, err := fn(subject, t)
		if err == nil {
			wakeup()
		}

		return id, err
	}
}

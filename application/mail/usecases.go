package mail

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
)

type Repository data.Repository[Outgoing, ID]

// SendMail takes the Mail and will try to publish it into either the given [Smtp] hint or whatever is currently defined
// as primary.
type SendMail func(subject auth.Subject, mail Mail) (ID, error)

type FindMailByID func(auth.Subject, ID) (std.Option[Outgoing], error)
type DeleteMailByID func(auth.Subject, ID) error
type FindAllMails func(auth.Subject) iter.Seq2[Outgoing, error]
type SaveMail func(auth.Subject, Outgoing) (ID, error)

type TemplateRepository data.Repository[Template, TemplateID]

type FindTemplateByID func(auth.Subject, TemplateID) (std.Option[Template], error)
type DeleteTemplateByID func(auth.Subject, TemplateID) error
type FindAllTemplates func(auth.Subject) iter.Seq2[Template, error]
type SaveTemplate func(auth.Subject, Template) (TemplateID, error)
type FindTemplateByNameAndLanguage func(subject auth.Subject, name string, languageTag string) (std.Option[Template], error)

type FindSmtpByID func(auth.Subject, SmtpID) (std.Option[Smtp], error)
type DeleteSmtpByID func(auth.Subject, SmtpID) error
type FindAllSmtp func(auth.Subject) iter.Seq2[Smtp, error]
type SaveSmtp func(auth.Subject, Smtp) (SmtpID, error)

type UseCases struct {
	Outgoing struct {
		FindByID   FindMailByID
		DeleteByID DeleteMailByID
		FindAll    FindAllMails
		Save       SaveMail
		repository Repository // intentionally not exposed, to avoid that devs can simply destroy invariants
	}

	Templates struct {
		FindTemplateByID              FindTemplateByID
		DeleteTemplateByID            DeleteTemplateByID
		FindAllTemplates              FindAllTemplates
		SaveTemplate                  SaveTemplate
		InitDefaultTemplates          InitDefaultTemplates
		FindTemplateByNameAndLanguage FindTemplateByNameAndLanguage
		repository                    TemplateRepository
	}

	Smtp struct {
		FindByID   FindSmtpByID
		DeleteByID DeleteSmtpByID
		FindAll    FindAllSmtp
		Save       SaveSmtp
		repository Repository // intentionally not exposed, to avoid that devs can simply destroy invariants
	}

	SendMail SendMail
}

func NewUseCases(outgoingRepo Repository, smtpRepo SmtpRepository) UseCases {
	outgoingCrud := auth.DecorateRepository(auth.DecoratorOptions{EntityName: "Ausgehende Mails", PermissionPrefix: "nago.mail.outgoing"}, outgoingRepo)
	sendMailFn := NewSendMail(outgoingRepo)

	smtpCrud := auth.DecorateRepository(auth.DecoratorOptions{EntityName: "SMTP Server", PermissionPrefix: "nago.mail.smtp"}, smtpRepo)

	var uc UseCases
	uc.SendMail = sendMailFn

	uc.Outgoing.DeleteByID = outgoingCrud.DeleteByID
	uc.Outgoing.FindByID = outgoingCrud.FindByID
	uc.Outgoing.FindAll = outgoingCrud.FindAll
	uc.Outgoing.Save = outgoingCrud.Upsert

	uc.Smtp.FindByID = smtpCrud.FindByID
	uc.Smtp.DeleteByID = smtpCrud.DeleteByID
	uc.Smtp.FindAll = smtpCrud.FindAll
	uc.Smtp.Save = smtpCrud.Upsert

	return uc
}

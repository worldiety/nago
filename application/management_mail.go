package application

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/mail/uimail"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
	"log/slog"
)

// HasMailManagement returns false, as long as [MailManagement] has not been requested to get initialized.
func (c *Configurator) HasMailManagement() bool {
	return c.mailManagement != nil
}

// MailManagement initializes and returns the default mailing subsystem.
// Without calling this, there will be no send mail support and no mail scheduler.
// Note, that neither the required permission will be registered nor any root view.
func (c *Configurator) MailManagement() (MailManagement, error) {
	if c.mailManagement == nil {
		tmp, err := NewMailManagement(c, MailPages{})
		if err != nil {
			return MailManagement{}, err
		}
		c.mailManagement = &tmp

		mail.StartScheduler(c.Context(), mail.ScheduleOptions{}, c.mailManagement.Smtp.Repository, c.mailManagement.Outgoing.Repository)

		c.RootView(c.mailManagement.Pages.SendMailTest, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.SendTestMailPage(wnd, c.mailManagement.SendMail)
		}))

		smtpUseCases := crud.UseCasesFromFuncs[mail.Smtp, mail.SmtpID](
			c.mailManagement.Smtp.FindByID,
			c.mailManagement.Smtp.FindAll,
			c.mailManagement.Smtp.DeleteByID,
			c.mailManagement.Smtp.Save,
		)

		c.RootView(c.mailManagement.Pages.SMTPServer, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.SmtpPage(wnd, smtpUseCases)
		}))

		outgoingUseCases := crud.UseCasesFromFuncs[mail.Outgoing, mail.ID](
			c.mailManagement.Outgoing.FindByID,
			c.mailManagement.Outgoing.FindAll,
			c.mailManagement.Outgoing.DeleteByID,
			c.mailManagement.Outgoing.Save,
		)

		c.RootView(c.mailManagement.Pages.OutgoingMailQueue, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.OutgoingQueuePage(wnd, outgoingUseCases)
		}))

		slog.Info("mail management configured")
	}

	return *c.mailManagement, nil
}

type MailPages struct {
	SMTPServer        core.NavigationPath
	OutgoingMailQueue core.NavigationPath
	MailScheduler     core.NavigationPath
	SendMailTest      core.NavigationPath
}

type MailManagement struct {
	Outgoing struct {
		FindByID   mail.FindMailByID
		DeleteByID mail.DeleteMailByID
		FindAll    mail.FindAllMails
		Save       mail.SaveMail
		Repository mail.Repository
	}

	Smtp struct {
		FindByID   mail.FindSmtpByID
		DeleteByID mail.DeleteSmtpByID
		FindAll    mail.FindAllSmtp
		Save       mail.SaveSmtp
		Repository mail.SmtpServerRepository
	}
	
	SendMail mail.SendMail

	Pages MailPages
}

func NewMailManagement(storage EntityStorageFactory, pages MailPages) (MailManagement, error) {
	if pages.MailScheduler == "" {
		pages.MailScheduler = "admin/mail/scheduler"
	}

	if pages.SendMailTest == "" {
		pages.SendMailTest = "admin/mail/test"
	}

	if pages.OutgoingMailQueue == "" {
		pages.OutgoingMailQueue = "admin/mail/outgoing"
	}

	if pages.SMTPServer == "" {
		pages.SMTPServer = "admin/mail/smtp"
	}

	outgoingMailsStore, err := storage.EntityStore("nago.mail.outgoing")
	if err != nil {
		return MailManagement{}, err
	}

	outgoingMailRepo := json.NewSloppyJSONRepository[mail.Outgoing, mail.ID](outgoingMailsStore)
	sendMail := mail.NewSendMail(outgoingMailRepo)

	outgoingUseCases := crud.NewUseCases[mail.Outgoing, mail.ID]("nago.mail.outgoing", outgoingMailRepo)

	smtpStore, err := storage.EntityStore("nago.mail.smtp")
	if err != nil {
		return MailManagement{}, err
	}

	smtpRepo := json.NewSloppyJSONRepository[mail.Smtp, mail.SmtpID](smtpStore)
	smtpUseCases := crud.NewUseCases[mail.Smtp, mail.SmtpID]("nago.mail.smtp", smtpRepo)

	return MailManagement{
		SendMail: sendMail,
		Outgoing: struct {
			FindByID   mail.FindMailByID
			DeleteByID mail.DeleteMailByID
			FindAll    mail.FindAllMails
			Save       mail.SaveMail
			Repository mail.Repository
		}{FindByID: outgoingUseCases.FindByID, DeleteByID: outgoingUseCases.DeleteByID, FindAll: outgoingUseCases.All, Save: outgoingUseCases.Save, Repository: outgoingMailRepo},
		Smtp: struct {
			FindByID   mail.FindSmtpByID
			DeleteByID mail.DeleteSmtpByID
			FindAll    mail.FindAllSmtp
			Save       mail.SaveSmtp
			Repository mail.SmtpServerRepository
		}{FindByID: smtpUseCases.FindByID, DeleteByID: smtpUseCases.DeleteByID, FindAll: smtpUseCases.All, Save: smtpUseCases.Save, Repository: smtpRepo},

		Pages: pages,
	}, nil
}

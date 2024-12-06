package application

import (
	"go.wdy.de/nago/application/mail"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
	"log/slog"
)

// HasMailManagement returns false, as long as [MailManagement] has not been requested to get initialized.
func (c *Configurator) HasMailManagement() bool {
	return c.mailManagement != nil
}

// MailManagementHandler configures a function which is invoked to post-modify the just instantiated
// [Configurator.MailManagement]. Usually, this should be called before the first call to [Configurator.MailManagement].
// If MailManagementHandler is called after MailManagement has been initialized, it will be applied immediately.
func (c *Configurator) MailManagementHandler(fn func(*MailManagement)) {
	if c.mailManagement == nil {
		// invoke later, if not yet created
		c.mailManagementHandler = fn
		return
	}

	// invoke immediately
	fn(c.mailManagement)
}

// MailManagement initializes and returns the default mailing subsystem.
// Without calling this, there will be no send mail support and no mail scheduler.
// Note, that neither the required permission will be registered nor any root view.
func (c *Configurator) MailManagement() (MailManagement, error) {
	if c.mailManagement == nil {
		tmp, err := NewMailManagement(c, uimail.MailPages{})
		if err != nil {
			return MailManagement{}, err
		}
		c.mailManagement = &tmp

		mail.StartScheduler(c.Context(), mail.ScheduleOptions{}, c.mailManagement.Smtp.repository, c.mailManagement.Outgoing.repository)

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

		templateUseCases := crud.UseCasesFromFuncs[mail.Template, mail.TemplateID](
			c.mailManagement.Templates.FindTemplateByID,
			c.mailManagement.Templates.FindAllTemplates,
			c.mailManagement.Templates.DeleteTemplateByID,
			c.mailManagement.Templates.SaveTemplate,
		)
		c.RootView(c.mailManagement.Pages.Templates, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.TemplatePage(wnd, templateUseCases)
		}))

		slog.Info("mail management configured")
	}

	return *c.mailManagement, nil
}

type MailManagement struct {
	Outgoing struct {
		FindByID   mail.FindMailByID
		DeleteByID mail.DeleteMailByID
		FindAll    mail.FindAllMails
		Save       mail.SaveMail
		repository mail.Repository // intentionally not exposed, to avoid that devs can simply destroy invariants
	}

	Templates struct {
		FindTemplateByID              mail.FindTemplateByID
		DeleteTemplateByID            mail.DeleteTemplateByID
		FindAllTemplates              mail.FindAllTemplates
		SaveTemplate                  mail.SaveTemplate
		InitDefaultTemplates          mail.InitDefaultTemplates
		FindTemplateByNameAndLanguage mail.FindTemplateByNameAndLanguage
		repository                    mail.TemplateRepository
	}

	Smtp struct {
		FindByID   mail.FindSmtpByID
		DeleteByID mail.DeleteSmtpByID
		FindAll    mail.FindAllSmtp
		Save       mail.SaveSmtp
		repository mail.SmtpServerRepository // intentionally not exposed, to avoid that devs can simply destroy invariants
	}

	SendMail mail.SendMail

	Pages uimail.MailPages
}

func NewMailManagement(storage EntityStorageFactory, pages uimail.MailPages) (MailManagement, error) {
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

	if pages.Templates == "" {
		pages.Templates = "admin/mail/templates"
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

	// templates
	templateStore, err := storage.EntityStore("nago.mail.template")
	if err != nil {
		return MailManagement{}, err
	}

	templateRepo := json.NewSloppyJSONRepository[mail.Template, mail.TemplateID](templateStore)
	templateUseCases := crud.NewUseCases[mail.Template, mail.TemplateID]("nago.mail.template", templateRepo)
	findTemplateByNameAndLanguage := mail.NewFindTemplateByNameAndLanguage(templateRepo)
	initDefaultTemplates := mail.NewInitDefaultTemplates(findTemplateByNameAndLanguage, templateUseCases.Save)

	if err := initDefaultTemplates(iam.Sys{}); err != nil {
		return MailManagement{}, err
	}

	return MailManagement{
		SendMail: sendMail,
		Templates: struct {
			FindTemplateByID              mail.FindTemplateByID
			DeleteTemplateByID            mail.DeleteTemplateByID
			FindAllTemplates              mail.FindAllTemplates
			SaveTemplate                  mail.SaveTemplate
			InitDefaultTemplates          mail.InitDefaultTemplates
			FindTemplateByNameAndLanguage mail.FindTemplateByNameAndLanguage
			repository                    mail.TemplateRepository
		}{FindTemplateByID: templateUseCases.FindByID, FindTemplateByNameAndLanguage: findTemplateByNameAndLanguage, DeleteTemplateByID: templateUseCases.DeleteByID, FindAllTemplates: templateUseCases.All, SaveTemplate: templateUseCases.Save, InitDefaultTemplates: initDefaultTemplates, repository: templateRepo},
		Outgoing: struct {
			FindByID   mail.FindMailByID
			DeleteByID mail.DeleteMailByID
			FindAll    mail.FindAllMails
			Save       mail.SaveMail

			repository mail.Repository
		}{FindByID: outgoingUseCases.FindByID, DeleteByID: outgoingUseCases.DeleteByID, FindAll: outgoingUseCases.All, Save: outgoingUseCases.Save, repository: outgoingMailRepo},
		Smtp: struct {
			FindByID   mail.FindSmtpByID
			DeleteByID mail.DeleteSmtpByID
			FindAll    mail.FindAllSmtp
			Save       mail.SaveSmtp
			repository mail.SmtpServerRepository
		}{FindByID: smtpUseCases.FindByID, DeleteByID: smtpUseCases.DeleteByID, FindAll: smtpUseCases.All, Save: smtpUseCases.Save, repository: smtpRepo},

		Pages: pages,
	}, nil
}

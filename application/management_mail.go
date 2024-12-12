package application

import (
	"go.wdy.de/nago/application/mail"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
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
		c.mailManagement = &MailManagement{}

		outgoingMailsStore, err := c.EntityStore("nago.mail.outgoing")
		if err != nil {
			return MailManagement{}, err
		}

		outgoingMailRepo := json.NewSloppyJSONRepository[mail.Outgoing, mail.ID](outgoingMailsStore)

		//	if err := initDefaultTemplates(iam.Sys{}); err != nil {
		//		return MailManagement{}, err
		//	}

		smtpStore, err := c.EntityStore("nago.mail.smtp")
		if err != nil {
			return MailManagement{}, err
		}

		smtpRepo := json.NewSloppyJSONRepository[mail.Smtp, mail.SmtpID](smtpStore)

		mail.StartScheduler(c.Context(), mail.ScheduleOptions{}, smtpRepo, outgoingMailRepo)

		c.mailManagement.Pages = uimail.Pages{
			SMTPServer:        "admin/mail/smtp",
			OutgoingMailQueue: "admin/mail/outgoing",
			MailScheduler:     "admin/mail/scheduler",
			SendMailTest:      "admin/mail/test",
			Templates:         "admin/mail/templates",
		}

		c.RootView(c.mailManagement.Pages.SendMailTest, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.SendTestMailPage(wnd, c.mailManagement.UseCases.SendMail)
		}))

		c.RootView(c.mailManagement.Pages.SMTPServer, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.SmtpPage(wnd, c.mailManagement.UseCases)
		}))
		c.RootView(c.mailManagement.Pages.OutgoingMailQueue, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.OutgoingQueuePage(wnd, c.mailManagement.UseCases)
		}))

		/*c.RootView(c.mailManagement.Pages.Templates, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.TemplatePage(wnd, c.mailManagement.UseCases)
		}))*/
	}

	return *c.mailManagement, nil
}

type MailManagement struct {
	UseCases mail.UseCases
	Pages    uimail.Pages
}

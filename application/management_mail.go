package application

import (
	"fmt"
	"go.wdy.de/nago/application/mail"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
	"log/slog"
	mail2 "net/mail"
)

// HasMailManagement returns false, as long as [MailManagement] has not been requested to get initialized.
func (c *Configurator) HasMailManagement() bool {
	return c.mailManagement != nil
}

// MailManagementHandler installs a mutator for a future invocation or immediately mutates the current configuration.
// Note, that even though most build-in implementations will perform a dynamic lookup, you may still want to install
// the handler BEFORE any *Management system has been initialized.
func (c *Configurator) MailManagementHandler(fn func(*MailManagement)) {
	c.mailManagementMutator = fn

	if c.mailManagement != nil {
		fn(c.mailManagement)
	}
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

		// we need the secret system to lookup the smtp
		secrets, err := c.SecretManagement()
		if err != nil {
			return MailManagement{}, fmt.Errorf("cannot get secret management: %w", err)
		}

		templates, err := c.TemplateManagement()
		if err != nil {
			return MailManagement{}, fmt.Errorf("cannot get template management: %w", err)
		}

		mail.StartScheduler(c.Context(), mail.ScheduleOptions{}, outgoingMailRepo, c.SysUser, secrets.UseCases.FindGroupSecrets)

		c.mailManagement.Pages = uimail.Pages{
			OutgoingMailQueue: "admin/mail/outgoing",
			MailScheduler:     "admin/mail/scheduler",
			SendMailTest:      "admin/mail/test",
		}

		c.mailManagement.UseCases, err = mail.NewUseCases(outgoingMailRepo, templates.UseCases.EnsureBuildIn)
		if err != nil {
			return MailManagement{}, fmt.Errorf("cannot create mail usecases: %w", err)
		}

		c.RootView(c.mailManagement.Pages.SendMailTest, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.SendTestMailPage(wnd, c.mailManagement.UseCases.SendMail)
		}))

		c.RootView(c.mailManagement.Pages.OutgoingMailQueue, c.DecorateRootView(func(wnd core.Window) core.View {
			return uimail.OutgoingQueuePage(wnd, c.mailManagement.UseCases)
		}))

		events.SubscribeFor[user.Created](c.eventBus, func(evt user.Created) {
			if !evt.NotifyUser {
				return
			}

			if err := c.SendVerificationMail(evt.ID); err != nil {
				slog.Error("user created but cannot send verification mail: %w", err)
			}
		})

	}

	return *c.mailManagement, nil
}

func (c *Configurator) SendPasswordResetMail(mail user.Email) error {
	usm, err := c.UserManagement()
	if err != nil {
		return fmt.Errorf("cannot get user management: %w", err)
	}

	optUser, err := usm.UseCases.FindByMail(c.SysUser(), mail)
	if err != nil {
		// this is a technical error we want to bubble up through user->tech support->admin->dev
		return fmt.Errorf("cannot find user: %w", err)
	}

	if optUser.IsNone() {
		// security note: intentionally do not expose this information to the frontend
		slog.Error("shall send verification mail but user not found", "mail", mail)
		return nil
	}

	// security note: intentionally create a new security code
	code, err := usm.UseCases.ResetPasswordRequestCode(mail, user.DefaultVerificationLifeTime)
	if err != nil {
		return fmt.Errorf("cannot reset password request code: %w", err)
	}

	usr := optUser.Unwrap()

	prefLang, _ := language.Parse(usr.Contact.DisplayLanguage)

	model := tplmail.PasswordResetModel{
		ID:                usr.ID,
		Firstname:         usr.Contact.Firstname,
		Lastname:          usr.Contact.Lastname,
		Email:             usr.Email,
		PreferredLanguage: prefLang,
		ConfirmURL:        core.URI(c.ContextPathURI(string(c.userManagement.Pages.ResetPassword), core.Values{"id": string(usr.ID), "code": code})), // here we expose our internal user id, not sure if this is a problem
		ApplicationName:   c.applicationName,
	}

	return c.SendMailTemplate(usr.Email, tplmail.ID, tplmail.ResetPasswordSubject, tplmail.ResetPassword, model)
}

func (c *Configurator) SendVerificationMail(uid user.ID) error {
	usm, err := c.UserManagement()
	if err != nil {
		return fmt.Errorf("cannot get user management: %w", err)
	}

	optUser, err := usm.UseCases.FindByID(c.SysUser(), uid)
	if err != nil {
		// this is a technical error we want to bubble up through user->tech support->admin->dev
		return fmt.Errorf("cannot find user: %w", err)
	}

	if optUser.IsNone() {
		// security note: intentionally do not expose this information to the frontend
		slog.Error("shall send verification mail but user not found", "id", uid)
		return nil
	}

	// security note: intentionally create a new security code
	code, err := usm.UseCases.ResetVerificationCode(uid, user.DefaultVerificationLifeTime)
	if err != nil {
		return fmt.Errorf("cannot reset confirm code: %w", err)
	}

	usr := optUser.Unwrap()

	prefLang, _ := language.Parse(usr.Contact.DisplayLanguage)

	model := tplmail.MailVerificationModel{
		ID:                usr.ID,
		Firstname:         usr.Contact.Firstname,
		Lastname:          usr.Contact.Lastname,
		Email:             usr.Email,
		PreferredLanguage: prefLang,
		ConfirmURL:        core.URI(c.ContextPathURI(string(c.userManagement.Pages.ConfirmMail), core.Values{"id": string(usr.ID), "code": code})), // here we expose our internal user id, not sure if this is a problem
		ApplicationName:   c.applicationName,
	}

	return c.SendMailTemplate(usr.Email, tplmail.ID, tplmail.MailVerificationSubject, tplmail.MailVerification, model)
}

func (c *Configurator) SendMailTemplate(to user.Email, tpl template.ID, subjName, bodyName template.DefinedTemplateName, tplModel any) error {
	mails, err := c.MailManagement()
	if err != nil {
		return err
	}

	subject, err := c.TemplateString(c.SysUser(), tpl, subjName, tplModel)
	if err != nil {
		return fmt.Errorf("cannot render subject: %w", err)
	}

	body, err := c.TemplateString(c.SysUser(), tpl, bodyName, tplModel)
	if err != nil {
		return fmt.Errorf("cannot render body: %w", err)
	}

	_, err = mails.UseCases.SendMail(user.NewSystem()(), mail.Mail{
		To: []mail2.Address{{
			Address: string(to),
		}},
		Subject: subject,
		Parts:   []mail.Part{mail.NewHtmlPart(body)},
	})

	return err
}

type MailManagement struct {
	UseCases mail.UseCases
	Pages    uimail.Pages
}

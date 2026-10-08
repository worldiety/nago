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
	"net/http"
	mail2 "net/mail"
	"os"
	"strings"

	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/mail/nms"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

// MailManagement is a nago system (Mail Management).
// It is responsible for sending emails within the platform,
// including notifications, password resets, and user registration confirmations.
// It requires an SMTP secret shared with the system group or the Nago Mail Service, see [EnvMailService],
// and uses predefined templates from TemplateManagement, which can be customized as needed.
type MailManagement struct {
	UseCases mail.UseCases
	Pages    uimail.Pages
	// MailService is the HTTP transport, which is used if no SMTP server is shared with the system group.
	MailService *nms.Service
}

const (
	// EnvMailService is the endpoint of the Nago Mail Service. If not set, [nms.DefaultEndpoint] is used. An empty
	// value or "off" disables the service.
	EnvMailService = "NAGO_MAIL_SERVICE"
	// EnvMailServiceToken is an optional refresh token of the Nago Mail Service. Without it, the instance enrolls by
	// a token exchange: a public https origin is called back, any other origin like localhost must call from an
	// address the service knows, see [mailServiceOrigin].
	EnvMailServiceToken = "NAGO_MAIL_SERVICE_TOKEN"
)

// mailServiceEndpoint returns the configured endpoint or the default one, an empty string disables the service.
func mailServiceEndpoint() string {
	v, ok := os.LookupEnv(EnvMailService)
	if !ok {
		return nms.DefaultEndpoint
	}

	v = strings.TrimSpace(v)
	if strings.EqualFold(v, "off") {
		return ""
	}

	return v
}

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

		statsStore, err := c.EntityStore("nago.mail.stats")
		if err != nil {
			return MailManagement{}, err
		}

		statsRepo := json.NewSloppyJSONRepository[mail.StatsBucket, mail.StatsBucketID](statsStore)

		healthStore, err := c.EntityStore("nago.mail.smtp_health")
		if err != nil {
			return MailManagement{}, err
		}

		healthRepo := json.NewSloppyJSONRepository[mail.ServerHealth, string](healthStore)

		serviceStore, err := c.EntityStore("nago.mail.nms")
		if err != nil {
			return MailManagement{}, err
		}

		nonces := nms.NewNonces()
		c.mailManagement.MailService = nms.NewService(nms.Options{
			Endpoint: mailServiceEndpoint(),
			Token:    strings.TrimSpace(os.Getenv(EnvMailServiceToken)),
			Origin: func() string {
				return c.PublicOrigin()
			},
			Nonces: nonces,
			States: json.NewSloppyJSONRepository[nms.State, string](serviceStore),
		})

		if c.mailManagement.MailService.Enabled() {
			c.HandleMethod(http.MethodGet, nms.NoncePath+"{nonce}", nonces.Handler())
			slog.Info("nago mail service enabled", "endpoint", c.mailManagement.MailService.Status().Endpoint)
		}

		notifyScheduler, wakeupScheduler := mail.NewWakeup()
		mail.StartScheduler(c.Context(), mail.ScheduleOptions{Stats: statsRepo, Health: healthRepo, Wakeup: wakeupScheduler, MailService: c.mailManagement.MailService}, outgoingMailRepo, c.SysUser, secrets.UseCases.FindGroupSecrets)

		c.mailManagement.Pages = uimail.Pages{
			Dashboard:         "admin/mail",
			OutgoingMailQueue: "admin/mail/outgoing",
			OutgoingMail:      "admin/mail/outgoing/detail",
			SmtpServers:       "admin/mail/smtp",
			SendMailTest:      "admin/mail/test",
			Templates:         templates.Pages.Projects,
			SecretVault:       secrets.Pages.Vault,
			SecretEdit:        secrets.Pages.EditSecret,
		}

		c.mailManagement.UseCases, err = mail.NewUseCasesWithStats(c.EventBus(), outgoingMailRepo, statsRepo, healthRepo, secrets.UseCases.FindGroupSecrets, notifyScheduler, templates.UseCases.EnsureBuildIn, c.SysUser, c.mailManagement.MailService)
		if err != nil {
			return MailManagement{}, fmt.Errorf("cannot create mail usecases: %w", err)
		}

		c.RootViewWithDecoration(c.mailManagement.Pages.Dashboard, func(wnd core.Window) core.View {
			return uimail.DashboardPage(wnd, c.mailManagement.Pages, c.mailManagement.UseCases)
		})

		c.RootViewWithDecoration(c.mailManagement.Pages.OutgoingMailQueue, func(wnd core.Window) core.View {
			return uimail.QueuePage(wnd, c.mailManagement.Pages, c.mailManagement.UseCases)
		})

		c.RootViewWithDecoration(c.mailManagement.Pages.OutgoingMail, func(wnd core.Window) core.View {
			return uimail.DetailPage(wnd, c.mailManagement.Pages, c.mailManagement.UseCases)
		})

		c.RootViewWithDecoration(c.mailManagement.Pages.SmtpServers, func(wnd core.Window) core.View {
			return uimail.SmtpPage(wnd, c.mailManagement.Pages, c.mailManagement.UseCases)
		})

		c.RootViewWithDecoration(c.mailManagement.Pages.SendMailTest, func(wnd core.Window) core.View {
			return uimail.SendTestMailPage(wnd, c.mailManagement.Pages, c.mailManagement.UseCases.SendMail, templates.UseCases.Execute)
		})

		events.SubscribeFor[user.Created](c.eventBus, func(evt user.Created) {
			if !evt.NotifyUser {
				return
			}

			if err := c.SendVerificationMail(evt.ID); err != nil {
				slog.Error("user created but cannot send verification mail", "err", err)
			}
		})

		events.SubscribeFor[user.EMailChanged](c.eventBus, func(evt user.EMailChanged) {
			if !evt.NotifyUser {
				return
			}

			if err := c.SendVerificationMail(evt.ID); err != nil {
				slog.Error("user mail changed but cannot send verification mail", "err", err)
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
		Title:             usr.Contact.Title,
		Salutation:        usr.Contact.Salutation,
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
		Title:             usr.Contact.Title,
		Salutation:        usr.Contact.Salutation,
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

	_, err = mails.UseCases.SendMail(c.SysUser(), mail.Mail{
		To: []mail2.Address{{
			Address: string(to),
		}},
		Subject: subject,
		Parts:   []mail.Part{mail.NewHtmlPart(body)},
	})

	return err
}

// PublicOrigin returns the origin the application is reachable under, like https://my-app.example.com, or
// http://localhost:<port> as long as nothing else is known. Services that enroll an instance by calling it back, like
// the Nago Mail Service or the Nago AI Service, announce it.
func (c *Configurator) PublicOrigin() string {
	return mailServiceOrigin(c.ContextPath(), c.getPort())
}

// mailServiceOrigin turns the context path into the origin of a token exchange.
//
// The context path is not always an origin: without HOSTNAME it is empty until the first window connects, and then
// it is the bare host of that request, like localhost:3000. The mail scheduler does not wait for a window, so a
// local instance falls back to http://localhost and its port, and a bare host is taken as http. The mail service
// decides whether it admits such an origin.
func mailServiceOrigin(contextPath string, port int) string {
	origin := strings.TrimSpace(contextPath)
	if origin == "" {
		return fmt.Sprintf("http://localhost:%d", port)
	}

	if !strings.Contains(origin, "://") {
		origin = "http://" + origin
	}

	return strings.TrimRight(origin, "/")
}

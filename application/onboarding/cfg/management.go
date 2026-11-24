// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"
	netmail "net/mail"
	"os"
	"slices"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/onboarding"
	uionboarding "go.wdy.de/nago/application/onboarding/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

type Options struct {
	// Email is the address of the first user. If empty, it is read from the environment variable EmailEnv.
	Email string
	// EmailEnv is the environment variable with the address of the first user, which is read, if Email is empty, and
	// which the hints to the operators name. Default is [onboarding.EnvInitialUserEmail], which is not read, if
	// another variable is set.
	EmailEnv string

	// Roles, Groups and Permissions are assigned to the first user.
	Roles       []role.ID
	Groups      []group.ID
	Permissions []permission.ID

	// ConfigurePermissions is optional and invoked after the first user has got the roles, groups and permissions.
	// An error aborts the setup and the user is removed again.
	ConfigurePermissions func(usr user.User) error

	// Page optionally replaces the default setup page, e.g. for a custom workflow. The [uionboarding.Flow] wraps
	// the use cases for the session of the window.
	Page func(wnd core.Window, flow *uionboarding.Flow) core.View

	// Scaffold renders the setup page within the decorator of the application, see
	// [application.Configurator.SetDecorator]. By default, the page stands alone like the login page.
	Scaffold bool

	// Exempt lists routes, which are shown even as long as the setup is pending, e.g. a legal notice. The internal
	// routes of the legal texts of the theme settings at the start and of Legal are always exempt.
	Exempt []core.NavigationPath

	// Legal optionally replaces the links to the legal texts below the card of the setup, which are taken from the
	// theme settings otherwise, see [theme.Settings]. It replaces both links: an empty value has no link, even if
	// the theme settings have one. A value is an external address with "://" or a route of the application.
	Legal *uionboarding.Legal

	// Deliver optionally replaces the mail with the code, e.g. in tests or by a log during development.
	Deliver onboarding.Deliver

	// Delivery optionally replaces the check, whether the code can be sent. By default, a custom Deliver is
	// assured. Otherwise, the mail is assured by an SMTP secret shared with the system group or the enrollment at
	// the Nago Mail Service, and tryable, while the Nago Mail Service is available for a token exchange.
	Delivery onboarding.CheckDelivery

	// CodeStatus optionally replaces the status of the mail with the code, e.g. [onboarding.CodeStatusInLog] for a
	// custom Deliver, which logs the code. By default, the status of the newest outgoing mail to the address since
	// the code has been sent is shown, and nothing for a custom Deliver.
	CodeStatus onboarding.ReadCodeStatus

	CodeLifetime   time.Duration // default 15 minutes
	ResendCooldown time.Duration // default 1 minute
	MaxAttempts    int           // wrong codes per code, default 5

	// NLSServer is the login service, which the setup offers for the single sign-on instead of a password. Default
	// is [onboarding.DefaultNLSServer].
	NLSServer string
	// DisableSSO offers no single sign-on, thus the first user always gets a password.
	DisableSSO bool
}

type Management struct {
	UseCases onboarding.UseCases
	Pages    uionboarding.Pages
}

// Enable installs the setup of the first user. As long as the instance has no user, every route shows the setup.
// It requires the user, session and mail management.
func Enable(cfg *application.Configurator, opts Options) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	users, err := cfg.UserManagement()
	if err != nil {
		return Management{}, fmt.Errorf("onboarding requires the user management: %w", err)
	}

	sessions, err := cfg.SessionManagement()
	if err != nil {
		return Management{}, fmt.Errorf("onboarding requires the session management: %w", err)
	}

	// a custom delivery is assured and tells no status by default, which an unset check and status mean
	deliver, checkDelivery, codeStatus := opts.Deliver, opts.Delivery, opts.CodeStatus
	if deliver == nil {
		mails, err := cfg.MailManagement()
		if err != nil {
			return Management{}, fmt.Errorf("onboarding requires the mail management: %w", err)
		}

		deliver = mailDeliver(cfg, mails)

		if checkDelivery == nil {
			secrets, err := cfg.SecretManagement()
			if err != nil {
				return Management{}, fmt.Errorf("onboarding requires the secret management: %w", err)
			}

			checkDelivery = mailDelivery(cfg.SysUser, secrets.UseCases.FindGroupSecrets, mails.MailService)
		}

		if codeStatus == nil {
			codeStatus = mailCodeStatus(cfg.SysUser, mails.UseCases.FindOutgoingIDs, mails.UseCases.FindOutgoingByID)
		}
	}

	emailEnv := cmp.Or(opts.EmailEnv, onboarding.EnvInitialUserEmail)
	email := opts.Email
	if email == "" {
		email = os.Getenv(emailEnv)
	}

	settingsMgmt, err := cfg.SettingsManagement()
	if err != nil {
		return Management{}, fmt.Errorf("onboarding requires the settings management: %w", err)
	}

	pendingSSO, err := application.JSONRepository[onboarding.PendingSSO, onboarding.PendingSSOID](cfg, "nago.onboarding.sso")
	if err != nil {
		return Management{}, fmt.Errorf("cannot get the store of the onboarding: %w", err)
	}

	uc := onboarding.NewUseCases(onboarding.Options{
		Email:                email,
		Roles:                opts.Roles,
		Groups:               opts.Groups,
		Permissions:          opts.Permissions,
		ConfigurePermissions: opts.ConfigurePermissions,
		CodeLifetime:         opts.CodeLifetime,
		ResendCooldown:       opts.ResendCooldown,
		MaxAttempts:          opts.MaxAttempts,
		NLSServer:            opts.NLSServer,
		DisableSSO:           opts.DisableSSO,
	}, onboarding.Deps{
		CountUsers:             users.UseCases.CountUsers,
		Create:                 users.UseCases.Create,
		UpdateVerification:     users.UseCases.UpdateVerification,
		UpdateOtherRoles:       users.UseCases.UpdateOtherRoles,
		UpdateOtherGroups:      users.UseCases.UpdateOtherGroups,
		UpdateOtherPermissions: users.UseCases.UpdateOtherPermissions,
		Delete:                 users.UseCases.Delete,
		LoginUser:              sessions.UseCases.LoginUser,
		Deliver:                deliver,
		Bus:                    cfg.EventBus(),
		CheckDelivery:          checkDelivery,
		CodeStatus:             codeStatus,
		LoadGlobal:             settingsMgmt.UseCases.LoadGlobal,
		StoreGlobal:            settingsMgmt.UseCases.StoreGlobal,
		StartNLSFlow:           sessions.UseCases.StartNLSFlow,
		PendingSSO:             pendingSSO,
	})

	// the access, which the setup has chosen for the single sign-on, applies after the first user has signed in by
	// the login service
	events.SubscribeFor[session.Authenticated](cfg.EventBus(), applyOnSignIn(users.UseCases.FindByID, uc.ApplySSOAccess))

	applyPendingSSO(pendingSSO, users.UseCases.FindByID, uc.ApplySSOAccess)

	management = Management{
		UseCases: uc,
		Pages:    uionboarding.Pages{Setup: "onboarding/setup"},
	}

	flowOpts := uionboarding.FlowOptions{
		Legal:          opts.Legal,
		EmailEnv:       emailEnv,
		MailServiceEnv: application.EnvMailService,
	}

	render := func(wnd core.Window) core.View {
		flow := uionboarding.NewFlow(wnd, uc, users.UseCases.SubjectFromUser, flowOpts)
		if opts.Page != nil {
			return opts.Page(wnd, flow)
		}

		return uionboarding.Setup(wnd, flow)
	}

	if opts.Scaffold {
		render = cfg.DecorateRootView(render)
	}

	cfg.RootView(management.Pages.Setup, render)

	exempt := exemptRoutes(settings.ReadGlobal[theme.Settings](settingsMgmt.UseCases.LoadGlobal), opts)
	cfg.AddRootViewInterceptor(func(wnd core.Window) (core.View, bool) {
		if slices.Contains(exempt, wnd.Path()) || !uc.State().Pending {
			return nil, false
		}

		return render(wnd), true
	})

	if st := uc.State(); st.Pending {
		switch {
		case st.Problem != onboarding.NoProblem:
			slog.Error("onboarding: the instance has no user, but the setup cannot be done", "problem", st.Problem, "env", emailEnv)
		case st.Delivery.Level == onboarding.DeliveryNone:
			slog.Error("onboarding: the instance has no user, but the code cannot be sent: no smtp server is shared with the system group and the nago mail service is not available", "env", application.EnvMailService)
		default:
			slog.Info("onboarding: the instance has no user, the setup is pending", "email", st.MaskedEmail())
		}
	}

	cfg.AddContextValue(core.ContextValue("nago.onboarding.management", management))
	slog.Info("installed onboarding module")
	return management, nil
}

// applyPendingSSO grants the chosen access, if the first user has signed in by single sign-on, while the
// application did not run anymore.
func applyPendingSSO(repo onboarding.PendingSSORepository, findByID user.FindByID, apply onboarding.ApplySSOAccess) {
	optPending, err := repo.FindByID(onboarding.PendingSSOKey)
	if err != nil || optPending.IsNone() {
		return
	}

	applyOnSignIn(findByID, apply)(session.Authenticated{User: optPending.Unwrap().Admin})
}

// applyOnSignIn grants the chosen access, when the first user has signed in by the login service. A sign-in by
// password grants nothing, because the gate exists to make the first user sign in by the login service first.
func applyOnSignIn(findByID user.FindByID, apply onboarding.ApplySSOAccess) func(evt session.Authenticated) {
	return func(evt session.Authenticated) {
		optUser, err := findByID(user.SU(), evt.User)
		if err != nil {
			slog.Error("onboarding: cannot find the signed-in user", "user", evt.User, "err", err)
			return
		}

		if optUser.IsNone() || !optUser.Unwrap().NLSManagedUser {
			return
		}

		if err := apply(evt.User); err != nil {
			slog.Error("onboarding: cannot grant the access of the single sign-on", "err", err)
		}
	}
}

// mailDeliver sends the code by the template of the system mail project. The template project of an instance,
// which has been created by an older version, lacks the template, so a plain mail is sent instead.
func mailDeliver(cfg *application.Configurator, mails application.MailManagement) onboarding.Deliver {
	return func(to user.Email, code string, validUntil time.Time) error {
		model := tplmail.OnboardingCodeModel{
			Email:           to,
			Code:            code,
			ValidMinutes:    int(math.Ceil(time.Until(validUntil).Minutes())),
			ApplicationName: cfg.Name(),
		}

		err := cfg.SendMailTemplate(to, tplmail.ID, tplmail.OnboardingCodeSubject, tplmail.OnboardingCode, model)
		if err == nil {
			return nil
		}

		slog.Warn("onboarding: cannot send the code by template, sending a plain mail", "err", err)
		_, err = mails.UseCases.SendMail(cfg.SysUser(), mail.Mail{
			To:      []netmail.Address{{Address: string(to)}},
			Subject: fmt.Sprintf("Einrichtungscode für %s: %s", model.ApplicationName, code),
			Parts: []mail.Part{mail.NewTextPart(fmt.Sprintf(
				"Für %s wurde die Einrichtung des ersten Kontos gestartet.\n\nDein Code: %s\n\nDer Code ist %d Minuten gültig. Falls Du die Einrichtung nicht gestartet hast, kannst Du diese E-Mail ignorieren.",
				model.ApplicationName, code, model.ValidMinutes,
			))},
		})

		return err
	}
}

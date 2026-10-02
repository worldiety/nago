// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding

import (
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
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/user/tplmail"
	"go.wdy.de/nago/presentation/core"
)

type Options struct {
	// Email overrides the environment variable [onboarding.EnvInitialUserEmail].
	Email string

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

	// Exempt lists routes, which are shown even as long as the setup is pending, e.g. a legal notice.
	Exempt []core.NavigationPath

	// Deliver optionally replaces the mail with the code, e.g. in tests.
	Deliver onboarding.Deliver

	CodeLifetime   time.Duration // default 15 minutes
	ResendCooldown time.Duration // default 1 minute
	MaxAttempts    int           // wrong codes per code, default 5
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

	deliver := opts.Deliver
	if deliver == nil {
		mails, err := cfg.MailManagement()
		if err != nil {
			return Management{}, fmt.Errorf("onboarding requires the mail management: %w", err)
		}

		deliver = mailDeliver(cfg, mails)
	}

	email := opts.Email
	if email == "" {
		email = os.Getenv(onboarding.EnvInitialUserEmail)
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
	})

	management = Management{
		UseCases: uc,
		Pages:    uionboarding.Pages{Setup: "onboarding/setup"},
	}

	render := func(wnd core.Window) core.View {
		flow := uionboarding.NewFlow(wnd, uc, users.UseCases.SubjectFromUser)
		if opts.Page != nil {
			return opts.Page(wnd, flow)
		}

		return uionboarding.Setup(wnd, flow)
	}

	if opts.Scaffold {
		render = cfg.DecorateRootView(render)
	}

	cfg.RootView(management.Pages.Setup, render)

	exempt := slices.Clone(opts.Exempt)
	cfg.AddRootViewInterceptor(func(wnd core.Window) (core.View, bool) {
		if slices.Contains(exempt, wnd.Path()) || !uc.State().Pending {
			return nil, false
		}

		return render(wnd), true
	})

	if st := uc.State(); st.Pending {
		switch st.Problem {
		case onboarding.NoProblem:
			slog.Info("onboarding: the instance has no user, the setup is pending", "email", st.MaskedEmail())
		default:
			slog.Error("onboarding: the instance has no user, but the setup cannot be done", "problem", st.Problem, "env", onboarding.EnvInitialUserEmail)
		}
	}

	cfg.AddContextValue(core.ContextValue("nago.onboarding.management", management))
	slog.Info("installed onboarding module")
	return management, nil
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

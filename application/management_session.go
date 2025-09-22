// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"go.wdy.de/nago/application/session"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

// SessionManagement is a nago system(Session Management).
// It provides functionality for handling user sessions,
// including login, logout, authentication state, and Single Sign-On (SSO)
// via the Nago Login Service (NLS).
//
// A session is identified by a unique cookie-based ID and represents the
// persistent state of a client. This ID is stable across tabs and device restarts.
//
// Key features include:
//   - Session lifecycle management (create, find, clear, timeout handling)
//   - Authentication via email/password or direct user ID
//   - Single Sign-On support (start, exchange, refresh NLS flows)
//   - Logout and session invalidation
//   - Tracking of creation and authentication timestamps
//   - Storing small key-value pairs in session context
//
// SessionManagement is automatically initialized when the application starts.
type SessionManagement struct {
	UseCases session.UseCases
	Pages    uisession.Pages
}

func (c *Configurator) Authentication() (any, error) {
	return c.SessionManagement()
}

func (c *Configurator) SessionManagement() (SessionManagement, error) {
	if c.sessionManagement == nil {

		// permissions are required
		if _, err := c.PermissionManagement(); err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get permission management: %w", err)
		}

		// sessions means very likely the full function set, we must edit users, therefore we need admin
		if _, err := c.AdminManagement(); err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get admin management: %w", err)
		}

		// sessions also means user registration and self-service, thus we need mailing
		if _, err := c.MailManagement(); err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get admin management: %w", err)
		}

		plainStore, err := c.EntityStore("nago.iam.session")
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get session store: %w", err)
		}

		key, err := c.MasterKey()
		if err != nil {
			return SessionManagement{}, fmt.Errorf("could not load master key: %w", err)
		}

		cryptSessionStore := crypto.NewBlobStore(plainStore, key)

		repo := json.NewSloppyJSONRepository[session.Session, session.ID](cryptSessionStore)

		userMgmt, err := c.UserManagement()
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get user management: %w", err)
		}

		setMgmt, err := c.SettingsManagement()
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get settings management: %w", err)
		}

		repoNonces, err := JSONRepository[session.NLSNonceEntry, session.NLSNonce](c, "nago.iam.nls.nonce")

		useCases := session.NewUseCases(
			c.ContextPathURI("", nil),
			setMgmt.UseCases.LoadGlobal,
			userMgmt.UseCases.MergeSingleSignOnUser,
			repo,
			repoNonces,
			userMgmt.UseCases.AuthenticateByPassword,
		)

		c.sessionManagement = &SessionManagement{
			UseCases: useCases,
			Pages: uisession.Pages{
				Login:          "account/login",
				Logout:         "account/logout",
				Authentication: "account/nls/authentication",
			},
		}

		settingsManagement, err := c.SettingsManagement()
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get settings management: %w", err)
		}

		c.RootView(c.sessionManagement.Pages.Login, func(wnd core.Window) core.View {
			return uisession.Login(
				wnd,
				c.sessionManagement.UseCases.Login,
				c.sessionManagement.UseCases.StartNLSFlow,
				c.userManagement.UseCases.SysUser,
				c.userManagement.UseCases.FindByMail,
				c.SendPasswordResetMail,
				c.SendVerificationMail,
				settingsManagement.UseCases.LoadGlobal,
				userMgmt.Pages.Register,
			)
		})

		c.RootView(c.sessionManagement.Pages.Authentication, func(wnd core.Window) core.View {
			return uisession.PageNLSAuthentication(wnd, c.sessionManagement.UseCases.ExchangeNLS)
		})

		c.RootView(c.sessionManagement.Pages.Logout, c.DecorateRootView(func(wnd core.Window) core.View {
			return uisession.Logout(wnd, c.sessionManagement.UseCases.Logout)
		}))

		c.AddOnWindowCreatedObserver(func(wnd core.Window) {
			optSession, err := useCases.FindSessionByID(wnd.Session().ID())
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			if optSession.IsNone() {
				wnd.UpdateSubject(nil)
				return
			}

			ses := optSession.Unwrap()

			if ses.User.IsNone() {
				wnd.UpdateSubject(nil)
				return
			}

			usrId := ses.User.Unwrap()

			optSubject, err := c.userManagement.UseCases.SubjectFromUser(wnd.Subject(), usrId)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			if optSubject.IsSome() {
				wnd.UpdateSubject(optSubject.Unwrap())
			} else {
				wnd.UpdateSubject(nil)
			}

		})

		c.AddContextValue(core.ContextValue("", c.sessionManagement.Pages))
	}

	return *c.sessionManagement, nil
}

package application

import (
	"fmt"
	"go.wdy.de/nago/application/session"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

type SessionManagement struct {
	UseCases session.UseCases
	Pages    uisession.Pages
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

		store, err := c.EntityStore("nago.iam.session")
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get session store: %w", err)
		}

		repo := json.NewSloppyJSONRepository[session.Session, session.ID](store)

		userMgmt, err := c.UserManagement()
		if err != nil {
			return SessionManagement{}, fmt.Errorf("cannot get user management: %w", err)
		}

		useCases := session.NewUseCases(
			repo,
			userMgmt.UseCases.AuthenticateByPassword,
		)

		c.sessionManagement = &SessionManagement{
			UseCases: useCases,
			Pages: uisession.Pages{
				Login:   "account/login",
				Logout:  "account/logout",
				Profile: "account/profile",
			},
		}

		c.RootView(c.sessionManagement.Pages.Login, c.DecorateRootView(func(wnd core.Window) core.View {
			return uisession.Login(wnd, c.sessionManagement.UseCases.Login)
		}))

		c.RootView(c.sessionManagement.Pages.Logout, c.DecorateRootView(func(wnd core.Window) core.View {
			return uisession.Logout(wnd, c.sessionManagement.UseCases.Logout)
		}))

		c.AddOnWindowCreatedObserver(func(wnd core.Window) {
			optSession, err := useCases.FindSessionByID(session.ID(wnd.SessionID()))
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			if optSession.IsNone() {
				wnd.UpdateSubject(auth.InvalidSubject{})
				return
			}

			ses := optSession.Unwrap()

			if ses.User.IsNone() {
				wnd.UpdateSubject(auth.InvalidSubject{})
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
				wnd.UpdateSubject(auth.InvalidSubject{})
			}

		})

	}

	return *c.sessionManagement, nil
}

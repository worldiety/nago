package application

import (
	"fmt"
	"go.wdy.de/nago/application/session"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

type SessionManagement struct {
	UseCases session.UseCases
	Pages    uisession.Pages
}

func (c *Configurator) SessionManagement() (SessionManagement, error) {
	if c.sessionManagement == nil {
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
			userMgmt.UseCases.FindByID,
			userMgmt.UseCases.System,
			userMgmt.UseCases.ViewOf,
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

	}

	return *c.sessionManagement, nil
}

package application

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

type UserManagement struct {
	UseCases user.UseCases
	Pages    uiuser.Pages
}

func (c *Configurator) UserManagement() (UserManagement, error) {
	if c.userManagement == nil {
		userStore, err := c.EntityStore("nago.iam.user")
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		userRepo := json.NewSloppyJSONRepository[user.User, user.ID](userStore)

		roleUseCases, err := c.RoleManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get role usecases: %w", err)
		}

		c.userManagement = &UserManagement{
			UseCases: user.NewUseCases(userRepo, roleUseCases.roleRepository),
			Pages:    uiuser.Pages{},
		}

		c.AddOnWindowCreatedObserver(func(wnd core.Window) {
			/*	optView, err := c.userManagement.UseCases.ViewOf(wnd.Subject(), session.ID(wnd.SessionID()))
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if optView.IsSome() {
					wnd.UpdateSubject(optView.Unwrap())
				} else {
					wnd.UpdateSubject(auth.InvalidSubject{})
				}*/

		})
	}

	return *c.userManagement, nil
}

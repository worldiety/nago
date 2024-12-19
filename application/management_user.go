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

		licenseUseCases, err := c.LicenseManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get license usecases: %w", err)
		}

		roleUseCases, err := c.RoleManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get role usecases: %w", err)
		}

		permissions, err := c.PermissionManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get permission usecases: %w", err)
		}

		groups, err := c.GroupManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get group usecases: %w", err)
		}

		_ = licenseUseCases

		c.userManagement = &UserManagement{
			UseCases: user.NewUseCases(userRepo, roleUseCases.roleRepository),
			Pages: uiuser.Pages{
				Users: "admin/accounts",
			},
		}

		// Oo we got some cycle
		billing, err := c.BillingManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get billing usecases: %w", err)
		}

		c.RootView(c.userManagement.Pages.Users, c.DecorateRootView(func(wnd core.Window) core.View {
			return uiuser.Users(wnd,
				c.userManagement.UseCases.Delete,
				c.userManagement.UseCases.FindAll,
				c.userManagement.UseCases.Create,
				c.userManagement.UseCases.UpdateOtherContact,
				c.userManagement.UseCases.UpdateOtherGroups,
				c.userManagement.UseCases.UpdateOtherRoles,
				c.userManagement.UseCases.UpdateOtherPermissions,
				c.userManagement.UseCases.UpdateOtherLicenses,
				roleUseCases.UseCases.FindAll,
				permissions.UseCases.FindAll,
				groups.UseCases.FindAll,
				c.userManagement.UseCases.SubjectFromUser,
				billing.UseCases.UserLicenses,
			)
		}))

	}

	return *c.userManagement, nil
}

package application

import (
	"fmt"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
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
			UseCases: user.NewUseCases(c.EventBus(), userRepo, roleUseCases.roleRepository),
			Pages: uiuser.Pages{
				Users: "admin/accounts",
			},
		}

		// Oo we got some cycle
		/*	billing, err := c.BillingManagement()
			if err != nil {
				return UserManagement{}, fmt.Errorf("cannot get billing usecases: %w", err)
			}*/

		c.RootView(c.userManagement.Pages.Users, c.DecorateRootView(func(wnd core.Window) core.View {
			var ucBillingUserLicense billing.UserLicenses
			if c.billingManagement != nil {
				ucBillingUserLicense = c.billingManagement.UseCases.UserLicenses
			} else {
				ucBillingUserLicense = func(subject auth.Subject) (billing.UserLicenseStatistics, error) {
					slog.Warn("User license billing not configured")
					return billing.UserLicenseStatistics{}, nil
				}
			}
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
				ucBillingUserLicense,
			)
		}))

	}

	return *c.userManagement, nil
}

func (c *Configurator) SysUser() auth.Subject {
	if c.userManagement == nil || c.userManagement.UseCases.SysUser == nil {
		return user.NewSystem()()
	}

	return c.userManagement.UseCases.SysUser()
}

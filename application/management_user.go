package application

import (
	"fmt"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"iter"
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

		settings, err := c.SettingsManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get settings usecases: %w", err)
		}
		_ = licenseUseCases

		c.userManagement = &UserManagement{
			UseCases: user.NewUseCases(c.EventBus(), settings.UseCases.LoadGlobal, userRepo, roleUseCases.roleRepository),
			Pages: uiuser.Pages{
				Users:         "admin/accounts",
				MyProfile:     "account/profile",
				MyContact:     "account/profile/contact",
				ConfirmMail:   "account/confirm",
				ResetPassword: "account/password/reset",
				Register:      "account/register",
			},
		}

		c.RootViewWithDecoration(c.userManagement.Pages.MyProfile, func(wnd core.Window) core.View {
			return uiuser.ProfilePage(
				wnd,
				c.userManagement.Pages,
				c.userManagement.UseCases.ChangeMyPassword,
				c.userManagement.UseCases.ReadMyContact,
				c.roleManagement.UseCases.FindMyRoles,
			)
		})

		c.RootViewWithDecoration(c.userManagement.Pages.MyContact, func(wnd core.Window) core.View {
			return uiuser.ContactPage(
				wnd,
				c.userManagement.Pages,
				c.userManagement.UseCases.UpdateMyContact,
				c.userManagement.UseCases.ReadMyContact,
			)
		})

		c.RootViewWithDecoration(c.userManagement.Pages.ConfirmMail, func(wnd core.Window) core.View {
			var path core.NavigationPath
			if c.sessionManagement != nil {
				path = c.sessionManagement.Pages.Login
			}

			return uiuser.ConfirmPage(
				wnd,
				path,
				c.userManagement.UseCases.ConfirmMail,
				c.SendVerificationMail,
				c.userManagement.UseCases.RequiresPasswordChange,
				c.userManagement.UseCases.SysUser,
				c.userManagement.UseCases.ChangeOtherPassword,
				c.sessionManagement.UseCases.Logout,
			)
		})

		c.RootViewWithDecoration(c.userManagement.Pages.ResetPassword, func(wnd core.Window) core.View {
			var path core.NavigationPath
			if c.sessionManagement != nil {
				path = c.sessionManagement.Pages.Login
			}

			return uiuser.ResetPasswordPage(
				wnd,
				path,
				c.userManagement.UseCases.ChangePasswordWithCode,
				c.sessionManagement.UseCases.Logout,
			)
		})

		c.RootView(c.userManagement.Pages.Register, func(wnd core.Window) core.View {
			return uiuser.PageSelfRegister(wnd, c.userManagement.UseCases.EMailUsed, c.userManagement.UseCases.Create)
		})

		c.RootViewWithDecoration(c.userManagement.Pages.Users, func(wnd core.Window) core.View {
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
		})

		c.AddSystemService("nago.users", form.AnyUseCaseList[user.User, user.ID](func(subject auth.Subject) iter.Seq2[user.User, error] {
			return c.userManagement.UseCases.FindAll(subject)
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

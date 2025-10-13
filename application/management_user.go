// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"iter"
	"log/slog"

	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
)

// UserManagement is a nago system(User Management).
// It is responsible for creating, managing, and maintaining user accounts within the platform.
// It provides both administrative and self-service features, allowing administrators to manage users and permissions, while enabling end users to maintain their own profiles.
// Typical workflows include:
//   - Creating and deleting user accounts
//   - Assigning roles, groups, and permissions
//   - Managing user profile data and contact details
//   - Password management (self-service and administrative)
//   - Email verification and account activation notifications
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

		userRepo := data.NewNotifyRepository(nil, json.NewSloppyJSONRepository[user.User, user.ID](userStore))

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

		sets, err := c.SettingsManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get settings usecases: %w", err)
		}
		_ = licenseUseCases

		storeGrants, err := c.EntityStore("nago.iam.grant")
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get entity store for grants: %w", err)
		}

		repoGrants := json.NewSloppyJSONRepository[user.Granting, user.GrantingKey](storeGrants)

		c.AddContextValue(core.ContextValue("nago.consent.options", form.AnyUseCaseList[user.ConsentOption, consent.ID](func(subject auth.Subject) iter.Seq2[user.ConsentOption, error] {
			return func(yield func(user.ConsentOption, error) bool) {
				usrSettings := settings.ReadGlobal[user.Settings](std.Must(c.SettingsManagement()).UseCases.LoadGlobal)
				for _, option := range usrSettings.Consents {
					if !yield(option, nil) {
						return
					}
				}
			}
		})))

		c.userManagement = &UserManagement{
			UseCases: user.NewUseCases(
				c.EventBus(),
				sets.UseCases.LoadGlobal,
				userRepo,
				repoGrants,
				roleUseCases.roleRepository,
				licenseUseCases.UseCases.PerUser.FindByID,
				roleUseCases.UseCases.FindByID,
			),
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
				c.userManagement.UseCases.FindByID,
				c.userManagement.UseCases.Consent,
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

			_ = ucBillingUserLicense
			_ = permissions
			return uiuser.PageUsers(wnd,
				c.userManagement.UseCases,
				groups.UseCases,
				roleUseCases.UseCases,
				permissions.UseCases,
			)
		})

		c.AddContextValue(core.ContextValue("nago.users", form.AnyUseCaseList[user.User, user.ID](func(subject auth.Subject) iter.Seq2[user.User, error] {
			return c.userManagement.UseCases.FindAll(subject)
		})))

		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.DisplayName))
		c.AddContextValue(core.ContextValue("", c.userManagement.Pages))
		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.FindByID))
		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.FindAllIdentifiers))

		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.GrantPermissions))
		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.ListGrantedUsers))
		c.AddContextValue(core.ContextValue("", c.userManagement.UseCases.ListGrantedPermissions))
	}

	return *c.userManagement, nil
}

func (c *Configurator) SysUser() auth.Subject {
	if c.userManagement == nil || c.userManagement.UseCases.SysUser == nil {
		return user.NewSystem()()
	}

	return c.userManagement.UseCases.SysUser()
}

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

	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/migration"
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
		userStore, err := c.EntityStore(string(user.Namespace))
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		userRepo := data.NewNotifyRepository(nil, json.NewSloppyJSONRepository[user.User, user.ID](userStore))

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

		images, err := c.ImageManagement()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get images: %w", err)
		}

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

		mg, err := c.Migrations()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get migrations: %w", err)
		}

		rdb, err := c.RDB()
		if err != nil {
			return UserManagement{}, fmt.Errorf("cannot get ReBAC database: %w", err)
		}

		// implementation note: it is important to first apply all user migrations, otherwise
		// we may risk data loss due to missing fields in current user entities
		if err := mg.Declare(newMigrateUserPermsToReBAC(userStore, rdb), migration.Options{Immediate: true}); err != nil {
			return UserManagement{}, fmt.Errorf("cannot declare migration: %w", err)
		}

		c.userManagement = &UserManagement{
			UseCases: user.NewUseCases(
				c.ctx,
				c.EventBus(),
				rdb,
				sets.UseCases.LoadGlobal,
				userRepo,
				roleUseCases.roleRepository,
				groups.UseCases.FindAll,
				roleUseCases.UseCases.FindByID,
				roleUseCases.UseCases.ListPermissions,
				images.UseCases.CreateSrcSet,
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

		rdb.RegisterResources(c.userManagement.UseCases.Resources)

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
		return user.NewSystem(c.Context())()
	}

	return c.userManagement.UseCases.SysUser()
}

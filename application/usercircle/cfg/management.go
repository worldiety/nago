// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgusercircle

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/usercircle"
	uiusercircles "go.wdy.de/nago/application/usercircle/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type Management struct {
	circleRepo usercircle.Repository
	UseCases   usercircle.UseCases
	Pages      uiusercircles.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := application.SystemServiceFor[Management](cfg, "")
	if ok {
		return management, nil
	}

	users, err := cfg.UserManagement()
	if err != nil {
		return Management{}, err
	}

	roles, err := cfg.RoleManagement()
	if err != nil {
		return Management{}, err
	}

	groups, err := cfg.GroupManagement()
	if err != nil {
		return Management{}, err
	}

	licenses, err := cfg.LicenseManagement()
	if err != nil {
		return Management{}, err
	}

	entityStore, err := cfg.EntityStore("nago.usercircle.circle")
	if err != nil {
		return Management{}, err
	}

	circleRepo := json.NewSloppyJSONRepository[usercircle.Circle, usercircle.ID](entityStore)
	useCases := usercircle.NewUseCases(circleRepo, users.UseCases, groups.UseCases.FindByID, roles.UseCases.FindByID)
	funcs := rcrud.Funcs[usercircle.Circle, usercircle.ID]{
		PermFindByID:   usercircle.PermFindByID,
		PermFindAll:    usercircle.PermFindAll,
		PermDeleteByID: usercircle.PermDeleteByID,
		PermCreate:     usercircle.PermCreate,
		PermUpdate:     usercircle.PermUpdate,
		FindByID:       useCases.FindByID,
		FindAll:        useCases.FindAll,
		DeleteByID:     useCases.DeleteByID,
		Create:         useCases.Create,
		Update:         useCases.Update,
		Upsert:         nil,
	}

	management = Management{
		circleRepo: circleRepo,
		UseCases:   useCases,
		Pages: uiusercircles.Pages{
			CirclesAdmin:          "admin/user/circles",
			MyCircle:              "admin/user/my-circle",
			MyCircleUsers:         "admin/user/my-circle/users",
			MyCircleRoles:         "admin/user/my-circle/roles",
			MyCircleRolesUsers:    "admin/user/my-circle/roles/users",
			MyCircleGroups:        "admin/user/my-circle/groups",
			MyCircleGroupsUsers:   "admin/user/my-circle/groups/users",
			MyCircleLicenses:      "admin/user/my-circle/licenses",
			MyCircleLicensesUsers: "admin/user/my-circle/license/users",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.CirclesAdmin, func(wnd core.Window) core.View {
		return uiusercircles.PageOverview(wnd, rcrud.UseCasesFrom(&funcs))
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircle, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleDashboard(wnd, management.Pages, useCases)
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircleUsers, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleUsers(wnd, useCases)
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircleRoles, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleRoles(wnd, management.Pages, useCases, roles.UseCases.FindByID)
	})
	cfg.RootViewWithDecoration(management.Pages.MyCircleRolesUsers, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleRolesUsers(wnd, management.Pages, useCases, roles.UseCases.FindByID)
	})
	cfg.RootViewWithDecoration(management.Pages.MyCircleGroups, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleGroups(wnd, management.Pages, useCases, groups.UseCases.FindByID)
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircleGroupsUsers, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleGroupsUsers(wnd, management.Pages, useCases, groups.UseCases.FindByID)
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircleLicenses, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleLicenses(wnd, management.Pages, useCases, licenses.UseCases.PerUser.FindByID)
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircleLicensesUsers, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircleLicensesUsers(
			wnd,
			management.Pages,
			useCases,
			licenses.UseCases.PerUser.FindByID,
			users.UseCases.AssignUserLicense,
			users.UseCases.UnassignUserLicense,
			users.UseCases.CountAssignedUserLicense,
		)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		group := admin.Group{
			Title: "Nutzerkreise",
			Entries: []admin.Card{
				{
					Title:      "Nutzerkreise",
					Text:       "Nutzerkreise verwalten, d.h. anlegen, bearbeiten und löschen.",
					Target:     management.Pages.CirclesAdmin,
					Permission: usercircle.PermFindAll,
				},
			},
		}

		for circle, _ := range useCases.MyCircles(subject) {
			group.Entries = append(group.Entries, admin.Card{
				Title:        circle.Name,
				Text:         circle.Description,
				Target:       management.Pages.MyCircle,
				TargetParams: core.Values{"circle": string(circle.ID)},
			})
		}

		return group
	})
	cfg.AddSystemService("nago.usercircles", management)

	slog.Info("installed user circle management")

	return management, nil
}

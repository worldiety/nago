// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/token"
	uitoken "go.wdy.de/nago/application/token/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

type TokenManagement struct {
	UseCases token.UseCases
	Pages    uitoken.Pages
}

func (c *Configurator) TokenManagement() (TokenManagement, error) {
	if c.tokenManagement == nil {
		tokenStore, err := c.EntityStore("nago.iam.token")
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get entity store: %w", err)
		}

		tokenRepo := json.NewSloppyJSONRepository[token.Token, token.ID](tokenStore)

		users, err := c.UserManagement()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get user usecases: %w", err)
		}

		groups, err := c.GroupManagement()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get group usecases: %w", err)
		}

		roles, err := c.RoleManagement()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get role usecases: %w", err)
		}

		licenses, err := c.LicenseManagement()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get license usecases: %w", err)
		}

		uc, err := token.NewUseCases(
			tokenRepo,
			users.UseCases.SubjectFromUser,
			groups.UseCases.FindByID,
			roles.UseCases.FindByID,
			users.UseCases.FindByID,
			licenses.UseCases.PerUser.FindByID,
		)
		
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get token usecases: %w", err)
		}

		c.tokenManagement = &TokenManagement{
			UseCases: uc,
			Pages: uitoken.Pages{
				Tokens:   "admin/iam/tokens",
				MyTokens: "",
			},
		}

		c.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
			if !subject.HasPermission(token.PermFindAll) {
				return admin.Group{}
			}

			return admin.Group{
				Title: "Access Tokens",
				Entries: []admin.Card{
					{Title: "Access Token", Text: "Der zentrale Zugriff u.a. auf REST-APIs kann über globale API Access Tokens geregelt werden.", Target: c.tokenManagement.Pages.Tokens},
				},
			}
		})

		c.RootViewWithDecoration(c.tokenManagement.Pages.Tokens, func(wnd core.Window) core.View {
			return uitoken.PageCrud(wnd, c.tokenManagement.UseCases)
		})

	}

	return *c.tokenManagement, nil
}

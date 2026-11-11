// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"io/fs"
	"net/http"
	"slices"
	"strings"

	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/migration"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/token"
	uitoken "go.wdy.de/nago/application/token/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/layout"
)

// TokenManagement is a nago system(Token Management).
// It configures and provides the backend for managing API access tokens.
//
// Tokens are used to authenticate requests against REST APIs. They can carry
// groups, roles, permissions, and licenses, similar to regular users.
// This enables external applications or services to act as authenticated subjects.
//
// It is typically used together with cfghapi.Management to secure API endpoints with bearer tokens.
type TokenManagement struct {
	UseCases token.UseCases
	Pages    uitoken.Pages
}

func (c *Configurator) TokenManagement() (TokenManagement, error) {
	if c.tokenManagement == nil {
		tokenStore, err := c.EntityStore(string(token.Namespace))
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

		mg, err := c.Migrations()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get migrations: %w", err)
		}

		rdb, err := c.RDB()
		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get rdb: %w", err)
		}
		rdb.AddRule(rebac.Rule{Source: token.Namespace, Via: role.Namespace})
		rdb.AddRule(rebac.Rule{Source: token.Namespace, Via: group.Namespace})

		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   role.Namespace,
			Relation: rebac.Member,
			Target:   token.Namespace,
		})
		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   group.Namespace,
			Relation: rebac.Member,
			Target:   token.Namespace,
		})

		// we have permissions which are generated and registered any time later at runtime, which must be generally allowed to be assigned
		permission.OnPermissionRegistered(func(permission permission.Permission) {
			rdb.RegisterStaticRule(rebac.StaticRule{
				Source:   token.Namespace,
				Relation: rebac.Relation(permission.ID),
				Target:   rebac.Global,
			})
		})

		for perm := range permission.All() {
			rdb.RegisterStaticRule(rebac.StaticRule{
				Source:   token.Namespace,
				Relation: rebac.Relation(perm.ID),
				Target:   rebac.Global,
			})
		}

		// implementation note: it is important to first apply all user migrations, otherwise
		// we may risk data loss due to missing fields in current user entities
		if err := mg.Declare(newMigrateTokenPermsToReBAC(tokenStore, rdb), migration.Options{Immediate: true}); err != nil {
			return TokenManagement{}, fmt.Errorf("cannot declare migration: %w", err)
		}

		uc, err := token.NewUseCases(
			c.Context(),
			tokenRepo,
			users.UseCases.SubjectFromUser,
			groups.UseCases.FindByID,
			roles.UseCases.FindByID,
			users.UseCases.FindByID,
			users.UseCases.GetAnonUser,
			rdb,
		)

		if err != nil {
			return TokenManagement{}, fmt.Errorf("cannot get token usecases: %w", err)
		}

		rdb.RegisterResources(uc.Resources)

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
			var opts uitoken.PageCrudOptions
			if c.servesAPIDoc() {
				opts.APIDoc = apiDocPath
			}

			return layout.WithBackButton(wnd, uitoken.PageCrudWithOptions(wnd, c.tokenManagement.UseCases, opts))
		})

	}

	return *c.tokenManagement, nil
}

// apiDocPath is the documentation of the OpenAPI document of the application: a frontend like pkg/stoplight below
// /api/doc/, which shows the document /api/doc/spec.json of cfghapi.
const apiDocPath = "/api/doc"

// servesAPIDoc reports whether the application has the documentation of apiDocPath at this moment: the document
// /api/doc/spec.json, which cfghapi registers, and a page /api/doc/index.html of a served frontend. Without both, the
// button of the token page would open a page, which cannot load a document, or a 404.
func (c *Configurator) servesAPIDoc() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	spec := slices.ContainsFunc(c.rawEndpoint, func(e rawEndpoint) bool {
		return e.pattern == apiDocPath+"/spec.json" && (e.method == "" || e.method == http.MethodGet)
	})

	if !spec {
		return false
	}

	return slices.ContainsFunc(c.fsys, func(fsys fs.FS) bool {
		_, err := fs.Stat(fsys, strings.TrimPrefix(apiDocPath, "/")+"/index.html")
		return err == nil
	})
}

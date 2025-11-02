// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package admin

import (
	uibackup "go.wdy.de/nago/application/backup/ui"
	uibilling "go.wdy.de/nago/application/billing/ui"
	uigroup "go.wdy.de/nago/application/group/ui"
	uilicense "go.wdy.de/nago/application/license/ui"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/application/permission"
	uipermission "go.wdy.de/nago/application/permission/ui"
	"go.wdy.de/nago/application/role"
	uirole "go.wdy.de/nago/application/role/ui"
	uisecret "go.wdy.de/nago/application/secret/ui"
	uisession "go.wdy.de/nago/application/session/ui"
	uitemplate "go.wdy.de/nago/application/template/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type FindAllGroups func(subject auth.Subject) []Group
type QueryGroups func(subject auth.Subject, filterText string) []Group

type Pages struct {
	Mail          uimail.Pages
	Billing       uibilling.Pages
	Session       uisession.Pages
	UsersOverview core.NavigationPath
	Role          uirole.Pages
	Group         uigroup.Pages
	Permission    uipermission.Pages
	License       uilicense.Pages
	Dashboard     core.NavigationPath
	Backup        uibackup.Pages
	Secret        uisecret.Pages
	Template      uitemplate.Pages
}

type Card struct {
	Title        string
	Text         string
	ID           string
	Target       core.NavigationPath
	TargetParams core.Values
	Role         role.ID
	Permission   permission.ID
}

type Group struct {
	Title   string
	Entries []Card
}

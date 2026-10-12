// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
)

type Pages struct {
	Vault core.NavigationPath
	// CreateSecret lets the user choose the type of a new secret and then fill it in, without storing anything
	// before it is saved.
	CreateSecret core.NavigationPath
	EditSecret   core.NavigationPath
}

// UseCases are required by the secret pages.
type UseCases struct {
	Secrets       secret.UseCases
	FindGroupByID group.FindByID
	FindMyGroups  group.FindMyGroups
	// FindAllUsers is invoked as system user, because everyone who may share a secret may share it with everyone.
	FindAllUsers user.FindAll
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package app

import (
	"go.wdy.de/nago/pkg/data"
)

type ID string

type App struct {
	ID          ID
	Name        string
	Description string
}

func (a App) Identity() ID {
	return a.ID
}

func (a App) WithIdentity(id ID) App {
	a.ID = id
	return a
}

type Repository data.Repository[App, ID]

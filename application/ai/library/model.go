// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type Library struct {
	ID          ID
	Name        string
	Description string
	CreatedBy   user.ID
	CreatedAt   xtime.UnixMilliseconds
}

func (l Library) Identity() ID {
	return l.ID
}

type CreateOptions struct {
	Name        string
	Description string
}

type UpdateOptions struct {
	Name        string
	Description string
}

type Repository data.Repository[Library, ID]

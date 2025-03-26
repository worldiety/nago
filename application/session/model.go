// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"time"
)

type ID string

type Session struct {
	ID              ID                  `json:"id"`
	User            std.Option[user.ID] `json:"user,omitempty,omitzero"`
	CreatedAt       time.Time           `json:"createdAt,omitempty,omitzero"`
	AuthenticatedAt time.Time           `json:"authenticatedAt,omitempty,omitzero"`
	Values          map[string]string   `json:"values,omitempty,omitzero"`
}

func (s Session) Identity() ID {
	return s.ID
}

type Repository = data.Repository[Session, ID]

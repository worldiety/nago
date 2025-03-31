// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"encoding/hex"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
	"time"
)

type ID string

// Hash is the hex encoding of the token hash.
type Hash string

type Plaintext = user.Password

type Token struct {
	ID         ID     `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Desription string `json:"description,omitempty"`

	Algorithm user.HashAlgorithm `json:"algorithm,omitempty"`
	TokenHash []byte             `json:"tokenHash,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
	// ValidUntil can be set to zero for unlimited lifetime
	ValidUntil xtime.Date `json:"validTill,omitempty"`

	// Impersonation has priority thus if valid, other Groups, Roles, Permissions and Resources are ignored.
	Impersonation option.Opt[user.ID] `json:"impersonation"`

	// Other permissions rules

	Groups      []group.ID                        `json:"groups,omitempty"`
	Roles       []role.ID                         `json:"roles,omitempty"`
	Permissions []permission.ID                   `json:"permissions,omitempty"`
	Licenses    []license.ID                      `json:"licenses,omitempty"`
	Resources   map[user.Resource][]permission.ID `json:"resources,omitempty" json:"resources,omitempty"`
}

func HashString(hash []byte) Hash {
	return Hash(hex.EncodeToString(hash))
}

func (t Token) Identity() ID {
	return t.ID
}

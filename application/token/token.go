// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"encoding/hex"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

// Hash is the hex encoding of the token hash.
type Hash string

type Plaintext = user.Password

type Token struct {
	ID          ID     `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	Algorithm user.HashAlgorithm `json:"algorithm,omitempty"`
	TokenHash []byte             `json:"tokenHash,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
	// ValidUntil can be set to zero for unlimited lifetime
	ValidUntil xtime.Date `json:"validTill,omitempty"`

	// Impersonation has priority thus if valid, other Groups, Roles, Permissions and Resources are ignored.
	Impersonation option.Opt[user.ID] `json:"impersonation"`
}

func HashString(hash []byte) Hash {
	return Hash(hex.EncodeToString(hash))
}

func (t Token) Identity() ID {
	return t.ID
}

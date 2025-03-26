// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"time"
)

type ID string

type Credentials interface {
	GetName() string
	Credentials() bool // open sum type which can be extended by anyone
	IsZero() bool
}

type Secret struct {
	ID ID `json:"id"`
	// Owners can read and write to the secret.
	Owners []user.ID `json:"owners"`
	// Groups denotes all groups into which this secret is implicitly available. This does not mean it shall be publicly
	// made read or writeable. For example, a secret may be added to [group.System] and is therefore generally
	// accessible for all use cases which inspect that group, whatever that means.
	// E.g. the nago mail handler will inspect the system group to find its SMTP secret.
	Groups      []group.ID  `json:"groups,omitempty"`
	LastMod     time.Time   `json:"lastMod"` // the time, the secret has been updated the last time
	Credentials Credentials `json:"credentials"`
}

func (s Secret) Identity() ID {
	return s.ID
}

type Repository data.Repository[Secret, ID]

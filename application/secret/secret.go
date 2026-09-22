// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"slices"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
)

type ID string

// AccessDeniedErr is returned, if a subject neither owns a secret nor belongs to any of the groups
// the secret has been shared into. See also [Secret.HasAccess].
var AccessDeniedErr = std.NewLocalizedError("Zugriff verweigert", "Nur Besitzer des Secrets oder Mitglieder der zugeordneten Gruppen dürfen darauf zugreifen.")

type Credentials interface {
	GetName() string
	Credentials() bool // open sum type which can be extended by anyone
	IsZero() bool
}

type Secret struct {
	ID ID `json:"id"`
	// Owners can read and write to the secret.
	Owners []user.ID `json:"owners"`
	// Groups denotes all groups into which this secret is shared. Every member of such a group may read and write
	// the secret, exactly like an owner, see [Secret.HasAccess]. Beyond that, the secret is implicitly available
	// for all use cases which inspect that group, whatever that means. E.g. the nago mail handler will inspect
	// the system group to find its SMTP secret.
	// Be aware, that sharing a secret into a large group like [group.System] exposes it to every member of it.
	Groups      []group.ID  `json:"groups,omitempty"`
	LastMod     time.Time   `json:"lastMod"` // the time, the secret has been updated the last time
	Credentials Credentials `json:"credentials"`
}

func (s Secret) Identity() ID {
	return s.ID
}

// HasAccess returns true, if the given subject owns the secret or is a member of at least one group the secret
// has been shared into. Access always implies both reading and writing, thus there is no finer grained
// distinction. A secret without any groups is only accessible by its owners.
func (s Secret) HasAccess(subject auth.Subject) bool {
	if subject == nil || !subject.Valid() {
		return false
	}

	if slices.Contains(s.Owners, subject.ID()) {
		return true
	}

	for _, gid := range s.Groups {
		if subject.HasGroup(gid) {
			return true
		}
	}

	return false
}

// IsOwner returns true, if the given subject is an explicit owner of the secret. Note, that a subject may
// still have full access through a group membership, see [Secret.HasAccess].
func (s Secret) IsOwner(subject auth.Subject) bool {
	if subject == nil || !subject.Valid() {
		return false
	}

	return slices.Contains(s.Owners, subject.ID())
}

type Repository data.Repository[Secret, ID]

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"iter"
	"reflect"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

// FindMySecrets returns all entries which are accessible by the given subject, thus those which it owns and
// those which have been shared into a group the subject belongs to. See [Secret.HasAccess].
type FindMySecrets func(subject auth.Subject) iter.Seq2[Secret, error]

// CreateSecret creates a new secret with the given subject as owner.
type CreateSecret func(subject auth.Subject, credentials Credentials) (ID, error)

// FindMySecretByID returns the secret, if the subject has access to it, see [Secret.HasAccess].
// Otherwise [AccessDeniedErr] is returned.
type FindMySecretByID func(subject auth.Subject, id ID) (std.Option[Secret], error)

// UpdateMySecretGroups updates the secret with the given group set. The subject must have access to the secret,
// see [Secret.HasAccess]. It is only allowed to add groups in which the subject is also a member. Groups in which
// the subject is not a member are kept untouched, even if they are not contained in the given set, so that nobody
// silently revokes the access of others. A subject which only has access through a group membership cannot remove
// the last group which grants it access.
type UpdateMySecretGroups func(subject auth.Subject, id ID, groups []group.ID) error

// UpdateMySecretOwners updates the owners of the secret. The subject must have access to the secret,
// see [Secret.HasAccess]. An owner is always kept as an owner to avoid orphaned secrets, whereas a subject which
// only has access through a group membership never promotes itself to an owner.
type UpdateMySecretOwners func(subject auth.Subject, id ID, owners []user.ID) error

// UpdateMyCredentials updates the secret with the given credentials, if the subject has access to it,
// see [Secret.HasAccess].
type UpdateMyCredentials func(subject auth.Subject, id ID, credentials Credentials) error

// DeleteMySecretByID removes the secret, if the subject has access to it, see [Secret.HasAccess].
// Deleting is idempotent, thus removing an unknown secret is not an error.
type DeleteMySecretByID func(subject auth.Subject, id ID) error

// FindGroupSecrets returns all those secrets which are associated with the given group and only if
// the given subject also belongs to that group. It returns also those secrets, which are not owned by the subject,
// thus this may cause a secret exposure issue.
type FindGroupSecrets func(subject auth.Subject, gid group.ID) iter.Seq2[Secret, error]

type MatchOptions struct {
	Hint   string   // may be a name or identifier. Ignored if not found, but prioritized if available.
	Group  group.ID // may be empty, but if not set, this must be matched due to security constraints.
	Expect bool     // if expected, ([option.None],nil) is never returned but (None,[os.ErrNotExists]) instead.
}

// Match searches for the given credential type and applies the given MatchOptions to eventually return the best
// match to use.
type Match func(subject auth.Subject, typ reflect.Type, opts MatchOptions) (option.Opt[Credentials], error)

// FindGroupCredentialsForType asserts a distinct
// credentials type for type safe queries. See [FindGroupSecrets] for security notes.
func FindGroupCredentialsForType[T Credentials](subject auth.Subject, secrets FindGroupSecrets, gid group.ID) iter.Seq2[T, error] {
	var zero T
	return func(yield func(T, error) bool) {
		for secret, err := range secrets(subject, gid) {
			if err != nil {
				if !yield(zero, err) {
					return
				}

				continue
			}

			if t, ok := secret.Credentials.(T); ok {
				if !yield(t, nil) {
					return
				}
			}
		}
	}
}

type UseCases struct {
	FindMySecrets        FindMySecrets
	CreateSecret         CreateSecret
	UpdateMyCredentials  UpdateMyCredentials
	FindMySecretByID     FindMySecretByID
	DeleteMySecretByID   DeleteMySecretByID
	UpdateMySecretGroups UpdateMySecretGroups
	FindGroupSecrets     FindGroupSecrets
	UpdateMySecretOwners UpdateMySecretOwners
	Match                Match
}

func NewUseCases(bus events.Bus, repository Repository) UseCases {
	var globalLock sync.Mutex
	findMySecretsFn := NewFindMySecrets(repository)
	createSecretFn := NewCreateSecret(&globalLock, bus, repository)
	updateMyCredentialsFn := NewUpdateMyCredentials(&globalLock, bus, repository)
	findMySecretByIDFn := NewFindMySecretByID(repository)
	deleteMySecretByIDFn := NewDeleteMySecretByID(bus, repository)
	updateMySecretGroupsFn := NewUpdateMySecretGroups(&globalLock, bus, repository)
	findGroupSecretsFn := NewFindGroupSecrets(repository)
	updateMySecretOwnersFn := NewUpdateMySecretOwners(&globalLock, bus, repository)

	return UseCases{
		FindMySecrets:        findMySecretsFn,
		CreateSecret:         createSecretFn,
		UpdateMyCredentials:  updateMyCredentialsFn,
		FindMySecretByID:     findMySecretByIDFn,
		DeleteMySecretByID:   deleteMySecretByIDFn,
		UpdateMySecretGroups: updateMySecretGroupsFn,
		FindGroupSecrets:     findGroupSecretsFn,
		UpdateMySecretOwners: updateMySecretOwnersFn,
		Match:                NewMatch(repository),
	}
}

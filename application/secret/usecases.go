package secret

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
)

// FindMySecrets returns all entries which are owned by the given subject. Other sharing aspects are not relevant.
// Especially, the relationship between a group and a secret is not evaluated.
type FindMySecrets func(subject auth.Subject) iter.Seq2[Secret, error]

// CreateSecret creates a new secret with the given subject as owner.
type CreateSecret func(subject auth.Subject, credentials Credentials) (ID, error)

type FindMySecretByID func(subject auth.Subject, id ID) (std.Option[Secret], error)

// UpdateMySecretGroups updates the secret with the given group set. It is only allowed, to add
// groups, in which the subject is also a member. For sure, the subject must be also the owner.
type UpdateMySecretGroups func(subject auth.Subject, id ID, groups []group.ID) error

// UpdateMyCredentials updates the secret with the given credentials, if the subject is the owner.
type UpdateMyCredentials func(subject auth.Subject, id ID, credentials Credentials) error

type DeleteMySecretByID func(subject auth.Subject, id ID) error

// FindGroupSecrets returns all those secrets which are associated with the given group. This does
// not perform any other security checks, because it is usually relevant for system services or similar (domain) global
// use cases. So keep in mind, that this may be used to expose all secrets without any security checks, if used
// directly.
type FindGroupSecrets func(gid group.ID) iter.Seq2[Secret, error]

// FindGroupCredentialsForType asserts a distinct
// credentials type for type safe queries. See [FindGroupSecrets] for security notes.
func FindGroupCredentialsForType[T Credentials](secrets FindGroupSecrets, gid group.ID) iter.Seq2[T, error] {
	var zero T
	return func(yield func(T, error) bool) {
		for secret, err := range secrets(gid) {
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
}

func NewUseCases(repository Repository) UseCases {
	var globalLock sync.Mutex
	findMySecretsFn := NewFindMySecrets(repository)
	createSecretFn := NewCreateSecret(&globalLock, repository)
	updateMyCredentialsFn := NewUpdateMyCredentials(&globalLock, repository)
	findMySecretByIDFn := NewFindMySecretByID(repository)
	deleteMySecretByIDFn := NewDeleteMySecretByID(repository)
	updateMySecretGroupsFn := NewUpdateMySecretGroups(&globalLock, repository)

	return UseCases{
		FindMySecrets:        findMySecretsFn,
		CreateSecret:         createSecretFn,
		UpdateMyCredentials:  updateMyCredentialsFn,
		FindMySecretByID:     findMySecretByIDFn,
		DeleteMySecretByID:   deleteMySecretByIDFn,
		UpdateMySecretGroups: updateMySecretGroupsFn,
	}
}

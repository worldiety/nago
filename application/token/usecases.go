// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
	"iter"
	"sync"
	"time"
)

type CreationData struct {
	Name        string
	Description string          `label:"Beschreibung"`
	ValidUntil  xtime.Date      `label:"Gültig bis"`
	Groups      []group.ID      `label:"Gruppen" source:"nago.groups"`
	Roles       []role.ID       `label:"Rollen" source:"nago.roles"`
	Permissions []permission.ID `label:"Berechtigungen" source:"nago.permissions"`
	Licenses    []license.ID    `label:"Lizenzen" source:"nago.licenses.user"`
	Resources   map[user.Resource][]permission.ID
}

// Create allocates a new token and returns the generated Hash and Plaintext. The Plaintext is never stored.
type Create func(subject auth.Subject, data CreationData) (ID, Plaintext, error)

type UserCreationData struct {
	Name        string
	Description string
	ValidUntil  time.Time
	User        user.ID
}

// CreateUserToken creates tokens which are inherit and follow the permissions of the given user.
// A User can always create a Token based on his own permissions.
type CreateUserToken func(subject auth.Subject, data UserCreationData) (Hash, Plaintext, error)

// AuthenticateSubject returns always a [auth.Subject]. If the plaintext token is unknown or out of life, an invalid
// subject is returned. Errors are only returned, if the infrastructure fails.
type AuthenticateSubject func(plaintext Plaintext) (auth.Subject, error)

// Delete removes a token. A subject can always remove his tokens.
type Delete func(subject auth.Subject, id ID) error

// Rotate keeps all settings of the given token but assigns a new token.
type Rotate func(subject auth.Subject, id ID) (Plaintext, error)

// FindAll returns all those tokens, which are in the visible scope. At least, a valid user can always see his own
// tokens.
type FindAll func(subject auth.Subject) iter.Seq2[Token, error]

type FindByID func(subject auth.Subject, id ID) (option.Opt[Token], error)

type ResolvedTokenRights struct {
	Impersonated bool
	Groups       []group.Group
	Roles        []role.Role
	Permissions  []permission.Permission
	Licenses     []license.UserLicense
}

type ResolveTokenRights func(subject auth.Subject, id ID) (ResolvedTokenRights, error)

type Repository data.Repository[Token, ID]
type UseCases struct {
	Create              Create
	Delete              Delete
	AuthenticateSubject AuthenticateSubject
	FindAll             FindAll
	Rotate              Rotate
	ResolveTokenRights  ResolveTokenRights
	FindByID            FindByID
}

func NewUseCases(
	repository Repository,
	subjectFromUser user.SubjectFromUser,
	findGroupByID group.FindByID,
	findRoleByID role.FindByID,
	findUserByID user.FindByID,
	getAnonUser user.GetAnonUser,
	findLicenseByID license.FindUserLicenseByID,
) (UseCases, error) {
	var mutex sync.Mutex

	// the reverse lookup keeps all plaintext tokens in memory and makes an O(1) lookup for the token so that
	// a potential REST api can be as fast as possible and only the initial call is slow
	subjectLookup := &concurrent.RWMap[Plaintext, user.Subject]{}

	reverseHashLookup := &concurrent.RWMap[Hash, ID]{}

	// security note: we must ensure, that at no time we have a collision for the token.
	// let us initially build the reverse hash lookup table
	init := func() error {
		mutex.Lock()
		defer mutex.Unlock()

		for token, err := range repository.All() {
			if err != nil {
				return err
			}

			h := HashString(token.TokenHash)

			if id, ok := reverseHashLookup.Get(h); ok {
				return fmt.Errorf("ambigous api access token: %s is shared by %s and %s", h, id, token.ID)
			}

			reverseHashLookup.Put(h, token.ID)
		}

		return nil
	}

	if err := init(); err != nil {
		return UseCases{}, err
	}

	repo := data.NewNotifyRepository(nil, repository)

	repo.AddDeletedObserver(func(repository data.Repository[Token, ID], deleted data.Deleted[ID]) error {
		// note, that these clean up functions are all O(n), but at least it is in memory and probably
		// fast enough for anything a nago app will ever serve.
		subjectLookup.DeleteFunc(func(t Plaintext, token user.Subject) bool {
			return ID(token.ID()) == deleted.ID
		})

		reverseHashLookup.DeleteFunc(func(hash Hash, id ID) bool {
			return id == deleted.ID
		})

		return nil
	})

	const algo = user.Argon2IdMin

	return UseCases{
		Delete:              NewDelete(&mutex, repo),
		FindAll:             NewFindAll(repo),
		Create:              NewCreate(&mutex, repo, algo, reverseHashLookup),
		AuthenticateSubject: NewAuthenticateSubject(repo, algo, reverseHashLookup, subjectFromUser, subjectLookup, getAnonUser, findRoleByID),
		Rotate:              NewRotate(&mutex, repo, algo, reverseHashLookup),
		FindByID:            NewFindByID(repo),
		ResolveTokenRights: NewResolveTokenRights(
			repo,
			findGroupByID,
			findRoleByID,
			findUserByID,
			findLicenseByID,
		),
	}, nil
}

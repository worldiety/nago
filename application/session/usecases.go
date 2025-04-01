// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"sync"
	"time"
)

type FindByID func(id ID) (std.Option[Session], error)

// FindUserSessionByID always returns a UserSession, however it does not necessarily create
// a new session. This is intentional, because anonymous user sessions may be normal and massive, thus
// writing each one without any will cause unnecessary I/O, occupy disk space and therefore increase costs without
// any real value.
type FindUserSessionByID func(id ID) UserSession

// Login authenticates a user by mail/password combination. The session is either hijacked or created.
// Not sure, what this means security wise, but stealing a session by compromising the client
// would always work.
type Login func(id ID, login user.Email, password user.Password) (bool, error)

// LoginUser blindly accepts the given user id and marks the session as authenticated. There is no additional check,
// if the given user really exists or if it is a valid user at all. Afterward, the session is treated as authenticated
// and other mechanics apply, to keep up with the user state, see [user.SubjectFromUser] for details.
type LoginUser func(id ID, usr user.ID) error
type Logout func(id ID) (bool, error)

// Clear removes all entries from the session store and is only required for fixing session problems.
type Clear func() error

// UserSession represents a persistent view of an assigned client.
// The current implementation uses a single store for all sessions, thus all values are always read and written
// at once. This makes especially writing session values quite expensive, so use it only if you really have to.
//
// Remember, that this only represents a technical session and is not a domain resource. Putting things here
// is only valid, if you have to track things beyond the scope lifetime. For user based resources,
// create a proper domain model.
type UserSession interface {
	// ID is a unique identifier, which is assigned to a client using some sort of cookie mechanism. This is a
	// pure random string and belongs to a distinct client instance. It is shared across multiple windows on the client,
	// especially when using multiple tabs or activity windows. You may use this for authentication mechanics,
	// however be careful not to break external security concerns by never revisiting the actual user authentication
	// state.
	// It usually outlives a frontend process and e.g. is restored after a device restart.
	ID() ID

	// User returns the authenticated user id.
	User() std.Option[user.ID]

	// CreatedAt returns the time at which this session has been created the first time.
	CreatedAt() std.Option[time.Time]

	// AuthenticatedAt returns time when this User has been authenticated the last time. Usually, this is the time
	// of the last login.
	AuthenticatedAt() std.Option[time.Time]

	// PutString creates or updates a new string.
	PutString(key string, value string) error

	// GetString returns the value.
	GetString(key string) (string, bool)
}
type UseCases struct {
	FindSessionByID     FindByID
	FindUserSessionByID FindUserSessionByID
	Login               Login
	LoginUser           LoginUser
	Logout              Logout
	Clear               Clear
}

func NewUseCases(repo Repository, authByPwd user.AuthenticateByPassword) UseCases {
	sessionByIdFn := NewFindByID(repo)
	loginFn := NewLogin(repo, authByPwd)
	logoutFn := NewLogout(repo)
	findUserSessionByIDFn := NewFindUserSessionByID(repo)

	var mutex sync.Mutex
	return UseCases{
		FindSessionByID:     sessionByIdFn,
		Login:               loginFn,
		LoginUser:           NewLoginUser(repo),
		Logout:              logoutFn,
		FindUserSessionByID: findUserSessionByIDFn,
		Clear:               NewClear(&mutex, repo),
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"sync"
	"time"

	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
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

type NLSNonce string
type NLSNonceEntry struct {
	ID       NLSNonce `json:"id"`
	Session  ID       `json:"sid"`
	Redirect string   `json:"redirect"`
}

func (e NLSNonceEntry) Identity() NLSNonce {
	return e.ID
}

type NLSNonceRepository data.Repository[NLSNonceEntry, NLSNonce]

type NLSRefreshToken string

// StartNLSFlow allocates a nonce for the given session and returns the according URL to invoke the configured
// nago-login-service. That service will eventually authenticate our nonce over a REST api. The server can
// then create and authenticate the user and attach it to the session identified by the nonce. The redirect
// is the root of the app by default.
type StartNLSFlow func(id ID) (uri string, err error)

// RefreshNLS tries to refresh the given session. It uses the refresh token to issue a request to the NLS server
// which in turn updates the user.
type RefreshNLS func(id ID) error

// ExchangeNLS tries to exchange the nonce for a refresh token and stores that for the given session.
type ExchangeNLS func(id ID, nonce NLSNonce) (redirect string, err error)

type UseCases struct {
	FindSessionByID     FindByID
	FindUserSessionByID FindUserSessionByID
	Login               Login
	LoginUser           LoginUser
	Logout              Logout
	Clear               Clear
	StartNLSFlow        StartNLSFlow
	ExchangeNLS         ExchangeNLS
}

func NewUseCases(bus events.Bus, defaultNLSRedirectURL string, loadGlobal settings.LoadGlobal, mergeSSO user.MergeSingleSignOnUser, repo Repository, nonceRepo NLSNonceRepository, authByPwd user.AuthenticateByPassword) UseCases {
	var mutex sync.Mutex

	sessionByIdFn := NewFindByID(repo)
	loginFn := NewLogin(bus, repo, authByPwd)
	logoutFn := NewLogout(repo)
	refreshNLSFn := NewRefreshNLS(&mutex, bus, repo, loadGlobal, mergeSSO, logoutFn)
	findUserSessionByIDFn := NewFindUserSessionByID(repo, refreshNLSFn)

	return UseCases{
		FindSessionByID:     sessionByIdFn,
		Login:               loginFn,
		LoginUser:           NewLoginUser(bus, repo),
		Logout:              logoutFn,
		FindUserSessionByID: findUserSessionByIDFn,
		Clear:               NewClear(&mutex, repo),
		StartNLSFlow:        NewStartNLSFlow(&mutex, defaultNLSRedirectURL, nonceRepo, loadGlobal),
		ExchangeNLS:         NewExchangeNLS(&mutex, bus, nonceRepo, repo, loadGlobal, refreshNLSFn),
	}
}

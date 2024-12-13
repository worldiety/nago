package session

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
)

type FindByID func(id ID) (std.Option[Session], error)

// Login authenticates a user by mail/password combination. The session is either hijacked or created.
// Not sure, what this means security wise, but stealing a session by compromising the client
// would always work.
type Login func(id ID, login user.Email, password user.Password) (bool, error)
type Logout func(id ID) (bool, error)

type UseCases struct {
	FindSessionByID FindByID
	Login           Login
	Logout          Logout
}

func NewUseCases(repo Repository, authByPwd user.AuthenticateByPassword) UseCases {
	sessionByIdFn := NewFindByID(repo)
	loginFn := NewLogin(repo, authByPwd)
	logoutFn := NewLogout(repo)

	return UseCases{
		FindSessionByID: sessionByIdFn,
		Login:           loginFn,
		Logout:          logoutFn,
	}
}

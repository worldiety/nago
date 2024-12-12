package session

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type Subject func(id ID) (auth.Subject, error)

// Login authenticates a user by mail/password combination. The session is either hijacked or created.
// Not sure, what this means security wise, but stealing a session by compromising the client
// would always work.
type Login func(id ID, login user.Email, password user.Password) (bool, error)
type Logout func(id ID) (bool, error)

type UseCases struct {
	Subject Subject
	Login   Login
	Logout  Logout
}

func NewUseCases(repo Repository, findUserById user.FindByID, system user.System, viewOf user.ViewOf, authByPwd user.AuthenticateByPassword) UseCases {
	subjectFn := NewSubject(repo, findUserById, system, viewOf)
	loginFn := NewLogin(repo, authByPwd)
	logoutFn := NewLogout(repo)
	
	return UseCases{
		Subject: subjectFn,
		Login:   loginFn,
		Logout:  logoutFn,
	}
}

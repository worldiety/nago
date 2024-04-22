package usm

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"time"
)

type SessionRepository = data.Repository[session, core.SessionID]

type session struct {
	ID   core.SessionID
	User std.Option[AuthenticatedUser]
}

func (s session) Identity() core.SessionID {
	return s.ID
}

type AuthenticatedUser struct {
	ID              auth.UserID
	Login           auth.EMail
	Firstname       string
	Lastname        string
	Verification    Verification
	StaticRoles     []auth.RoleID
	AuthenticatedAt time.Time
}

func (u AuthenticatedUser) UserID() auth.UserID {
	return u.ID
}

func (u AuthenticatedUser) Verified() bool {
	return u.Verification != None
}

func (u AuthenticatedUser) Roles(yield func(auth.RoleID) bool) {
	for _, role := range u.StaticRoles {
		if !yield(role) {
			return
		}
	}
}

func (u AuthenticatedUser) Email() auth.EMail {
	return u.Login
}

func (u AuthenticatedUser) Name() string {
	return u.Firstname + " " + u.Lastname
}

func (u AuthenticatedUser) Valid() bool {
	return time.Now().Sub(u.AuthenticatedAt) < time.Hour*24*30
}

package core

import "go.wdy.de/nago/auth"

type invalidUser struct {
}

func (i invalidUser) UserID() auth.UserID {
	return ""
}

func (i invalidUser) Roles(yield func(auth.RoleID) bool) {
}

func (i invalidUser) SessionID() string {
	return ""
}

func (i invalidUser) Verified() bool {
	return false
}

func (i invalidUser) Email() auth.EMail {
	return ""
}

func (i invalidUser) Name() string {
	return ""
}

func (i invalidUser) Valid() bool {
	return false
}

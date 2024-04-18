package usm

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type UserRepository = data.Repository[User, auth.UserID]

type Verification string

const (
	None                      Verification = ""
	VerifiedByAdmin           Verification = "admin"
	VerifiedByMailDoubleOptIn Verification = "mail-double-opt-in"
)

type User struct {
	ID           auth.UserID
	Login        auth.EMail
	Firstname    string
	Lastname     string
	Salt         []byte
	PasswordHash []byte
	Verification Verification
	Disabled     bool
	StaticRoles  []auth.RoleID
}

func (u User) Identity() auth.UserID {
	return u.ID
}

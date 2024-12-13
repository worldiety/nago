package user

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/enum"
	"time"
)

func NewEnableBootstrapAdmin(repo Repository, system System, userByMail FindByMail) EnableBootstrapAdmin {
	return func(aliveUntil time.Time, password Password) (ID, error) {
		optUsr, err := userByMail(system(), "admin@localhost")
		if err != nil {
			return "", err
		}

		var usr User

		if optUsr.IsSome() {
			usr = optUsr.Unwrap()
		} else {
			usr.ID = data.RandIdent[ID]() // a random admin user id makes some attacks impossible
			usr.Email = "admin@localhost"
			usr.Contact.Firstname = "admin"
			usr.Contact.Lastname = "admin"
		}

		usr.Status = enum.Make[AccountStatus](EnabledUntil{ValidUntil: aliveUntil})

		// we are not allowed to have domain specific permissions, only those to bootstrap other users.
		// even admins must not see customers secret domain stuff.
		usr.Permissions = []permission.ID{
			PermCreate,
			PermFindByID,
			PermDelete,
			PermChangeOtherPassword,
			PermFindAll,
			PermFindByMail,
			PermUpdateOtherContact,
			PermChangeOtherPassword,
			PermUpdateOtherPermissions,
			PermUpdateOtherRoles,
			group.PermCreate,
			group.PermFindAll,
			group.PermDelete,
			group.PermFindByID,
			group.PermUpdate,
			role.PermCreate,
			role.PermFindByID,
			role.PermFindAll,
			role.PermDelete,
			role.PermUpdate,
			permission.PermFindAll,
		}

		hType := Argon2IdMin
		salt, hash, err := password.Hash(hType)
		if err != nil {
			return "", fmt.Errorf("hashing password: %w", err)
		}

		usr.LastPasswordChangedAt = time.Now()
		usr.EMailVerified = true
		usr.Salt = salt
		usr.PasswordHash = hash
		usr.Algorithm = hType

		if err := repo.Save(usr); err != nil {
			return "", fmt.Errorf("cannot save bootstrap user: %w", err)
		}

		return usr.ID, nil
	}
}

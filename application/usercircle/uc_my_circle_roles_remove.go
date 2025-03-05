package usercircle

import (
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"slices"
	"sync"
)

func NewMyCircleRolesRemove(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleRolesRemove {
	return func(subject auth.Subject, circleId ID, usrId user.ID, roles ...role.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		usr.Roles = slices.DeleteFunc(usr.Roles, func(id role.ID) bool {
			for _, rid := range roles {
				if rid == id {
					return true
				}
			}

			return false
		})

		return users.UpdateOtherRoles(user.SU(), usrId, usr.Roles)
	}
}

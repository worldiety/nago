package usercircle

import (
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"slices"
	"sync"
)

func NewMyCircleRolesAdd(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleRolesAdd {
	return func(admin user.ID, circleId ID, usrId user.ID, roles ...role.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, admin, circleId, usrId)
		if err != nil {
			return err
		}

		for _, rid := range roles {
			if !slices.Contains(usr.Roles, rid) {
				usr.Roles = append(usr.Roles, rid)
			}
		}

		return users.UpdateOtherRoles(user.SU(), usrId, usr.Roles)
	}
}

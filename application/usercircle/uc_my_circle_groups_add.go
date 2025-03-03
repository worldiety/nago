package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"slices"
	"sync"
)

func NewMyCircleGroupsAdd(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleGroupsAdd {
	return func(admin user.ID, circleId ID, usrId user.ID, groups ...group.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, admin, circleId, usrId)
		if err != nil {
			return err
		}

		for _, rid := range groups {
			if !slices.Contains(usr.Groups, rid) {
				usr.Groups = append(usr.Groups, rid)
			}
		}

		return users.UpdateOtherGroups(user.SU(), usrId, usr.Groups)
	}
}

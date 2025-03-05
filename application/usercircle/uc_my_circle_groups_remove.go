package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"slices"
	"sync"
)

func NewMyCircleGroupsRemove(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleGroupsRemove {
	return func(subject auth.Subject, circleId ID, usrId user.ID, groups ...group.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		usr.Groups = slices.DeleteFunc(usr.Groups, func(id group.ID) bool {
			for _, rid := range groups {
				if rid == id {
					return true
				}
			}

			return false
		})

		return users.UpdateOtherGroups(user.SU(), usrId, usr.Groups)
	}
}

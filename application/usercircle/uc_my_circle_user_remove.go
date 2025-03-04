package usercircle

import (
	"go.wdy.de/nago/application/user"
	"sync"
)

func NewMyCircleUserRemove(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleUserRemove {
	return func(admin user.ID, circleId ID, usrId user.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, admin, circleId, usrId)
		if err != nil {
			return err
		}

		return users.Delete(user.SU(), usr.ID)
	}
}

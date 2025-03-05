package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"sync"
)

func NewMyCircleUserRemove(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleUserRemove {
	return func(subject auth.Subject, circleId ID, usrId user.ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		return users.Delete(user.SU(), usr.ID)
	}
}

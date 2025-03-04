package usercircle

import (
	"go.wdy.de/nago/application/user"
	"sync"
)

func NewMyCircleUserVerified(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleUserVerified {
	return func(admin user.ID, circleId ID, usrId user.ID, verified bool) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, admin, circleId, usrId)
		if err != nil {
			return err
		}

		return users.UpdateVerification(user.SU(), usr.ID, verified)
	}
}

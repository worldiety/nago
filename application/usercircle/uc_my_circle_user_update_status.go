package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"sync"
)

func NewMyCircleUserUpdateStatus(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleUserUpdateStatus {
	return func(subject auth.Subject, circleId ID, usrId user.ID, status user.AccountStatus) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		return users.UpdateAccountStatus(user.SU(), usr.ID, status)
	}
}

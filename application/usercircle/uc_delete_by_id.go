package usercircle

import (
	"go.wdy.de/nago/auth"
	"sync"
)

func NewDeleteByID(mutex *sync.Mutex, repo Repository) DeleteByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteByID); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}

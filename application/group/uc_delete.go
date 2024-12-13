package group

import (
	"go.wdy.de/nago/application/permission"
	"sync"
)

func NewDelete(mutex *sync.Mutex, repo Repository) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}

package usercircle

import (
	"go.wdy.de/nago/auth"
	"os"
	"sync"
)

func NewUpdate(mutex *sync.Mutex, repo Repository) Update {
	return func(subject auth.Subject, circle Circle) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optCircle, err := repo.FindByID(circle.ID)
		if err != nil {
			return err
		}

		if optCircle.IsNone() {
			return os.ErrNotExist
		}

		return repo.Save(circle)
	}
}

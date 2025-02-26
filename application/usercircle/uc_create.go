package usercircle

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"os"
	"sync"
)

func NewCreate(mutex *sync.Mutex, repo Repository) Create {
	return func(subject auth.Subject, circle Circle) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if circle.ID == "" {
			circle.ID = data.RandIdent[ID]()
		}

		optCircle, err := repo.FindByID(circle.ID)
		if err != nil {
			return "", err
		}

		if optCircle.IsSome() {
			return "", os.ErrExist
		}

		return circle.ID, repo.Save(circle)
	}
}

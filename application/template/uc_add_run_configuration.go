package template

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"sync"
)

func NewAddRunConfiguration(mutex *sync.Mutex, repo Repository) AddRunConfiguration {
	return func(subject auth.Subject, pid ID, configuration RunConfiguration) error {
		if err := subject.AuditResource(repo.Name(), string(pid), PermAddRunConfiguration); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return nil
		}

		prj := optPrj.Unwrap()
		configuration.ID = data.RandIdent[string]()
		prj.RunConfigurations = append(prj.RunConfigurations, configuration)
		return repo.Save(prj)
	}
}

package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewStop(m *Manager) Stop {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermStop); err != nil {
			return err
		}

		return m.Stop(id)
	}
}

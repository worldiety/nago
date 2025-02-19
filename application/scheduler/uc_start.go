package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewStart(m *Manager) Start {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermStart); err != nil {
			return err
		}

		return m.Start(id)
	}
}

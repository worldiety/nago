package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewConfigure(m *Manager) Configure {
	return func(subject auth.Subject, opts Options) error {
		if err := subject.Audit(PermConfigure); err != nil {
			return err
		}

		return m.Configure(opts)
	}
}

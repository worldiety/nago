package scheduler

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
)

func NewListSchedulers(m *Manager) ListSchedulers {
	return func(subject auth.Subject) iter.Seq2[Options, error] {
		if err := subject.Audit(PermListSchedulers); err != nil {
			return xiter.WithError[Options](err)
		}

		tmp := m.Scheduler()
		return xslices.Values2[[]Options, Options, error](tmp)
	}
}

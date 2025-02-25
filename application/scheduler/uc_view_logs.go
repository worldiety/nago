package scheduler

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
)

func NewViewLogs(m *Manager) ViewLogs {
	return func(subject auth.Subject, id ID) iter.Seq2[LogEntry, error] {
		if err := subject.Audit(PermViewLogs); err != nil {
			return xiter.WithError[LogEntry](err)
		}

		return xslices.Values2[[]LogEntry, LogEntry, error](m.Logs(id))
	}
}

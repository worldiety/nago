package scheduler

import (
	"context"
	"errors"
	"go.wdy.de/nago/auth"
)

func NewExecuteNow(m *Manager) ExecuteNow {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermExecuteNow); err != nil {
			return err
		}

		err := m.ExecuteNow(id)
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	}
}

package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"time"
)

func NewLoginUser(sessions Repository) LoginUser {
	return func(id ID, usr user.ID) error {
		// first install the session
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return fmt.Errorf("sessions.FindByID failed: %w", err)
		}

		var session Session
		if optSession.IsNone() {
			session.ID = id
			session.CreatedAt = time.Now()
		} else {
			session = optSession.Unwrap()
		}

		session.User = std.Some(usr)
		session.AuthenticatedAt = time.Now()

		if err := sessions.Save(session); err != nil {
			return fmt.Errorf("sessions.Save failed: %w", err)
		}

		return nil
	}
}

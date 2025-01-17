package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"time"
)

func NewLogout(sessions Repository) Logout {
	return func(id ID) (bool, error) {
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return false, fmt.Errorf("sessions.FindByID failed: %v", err)
		}

		if optSession.IsNone() {
			// unknown, nothing to do
			return true, nil
		}

		session := optSession.Unwrap()
		session.User = std.None[user.ID]()
		session.AuthenticatedAt = time.Time{}
		if err := sessions.Save(session); err != nil {
			return false, fmt.Errorf("sessions.Save failed: %v", err)
		}

		return true, nil
	}
}

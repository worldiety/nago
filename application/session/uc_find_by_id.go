package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
	"time"
)

func NewFindByID(sessions Repository) FindByID {
	return func(id ID) (std.Option[Session], error) {
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return std.None[Session](), fmt.Errorf("failed to find session: %w", err)
		}

		if optSession.IsNone() {
			return std.None[Session](), nil
		}

		session := optSession.Unwrap()

		if session.User.IsNone() {
			return std.None[Session](), nil
		}

		const day = 24 * time.Hour
		if time.Now().Sub(session.AuthenticatedAt) > day*30*3 {
			slog.Error("session expired for user", "sessionID", session.ID, "user", session.User)
			session.User = std.None[user.ID]()
			session.AuthenticatedAt = time.Time{}
			if err := sessions.Save(session); err != nil {
				return std.None[Session](), fmt.Errorf("failed to save expired session: %w", err)
			}

			return std.None[Session](), nil
		}

		return std.Some(session), nil
	}
}

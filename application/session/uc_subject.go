package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
	"time"
)

func NewSubject(sessions Repository, userById user.FindByID, system user.System, viewOf user.ViewOf) Subject {
	return func(id ID) (auth.Subject, error) {
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return auth.InvalidSubject{}, fmt.Errorf("failed to find session: %w", err)
		}

		if optSession.IsNone() {
			return auth.InvalidSubject{}, nil
		}

		session := optSession.Unwrap()

		if session.User.IsNone() {
			return auth.InvalidSubject{}, nil
		}

		const day = 24 * time.Hour
		if time.Now().Sub(session.AuthenticatedAt) > day*30*3 {
			slog.Error("session expired for user", "sessionID", session.ID, "user", session.User)
			session.User = std.None[user.ID]()
			session.AuthenticatedAt = time.Time{}
			if err := sessions.Save(session); err != nil {
				return auth.InvalidSubject{}, fmt.Errorf("failed to save expired session: %w", err)
			}

			return auth.InvalidSubject{}, nil
		}

		optView, err := viewOf(system(), session.User.Unwrap())
		if err != nil {
			return auth.InvalidSubject{}, fmt.Errorf("failed to find user: %w", err)
		}

		if optView.IsNone() {
			// important, the user may have been deleted, but we are lazy, so we have them as stale references here
			return auth.InvalidSubject{}, nil
		}

		// note, that user.View handles already its valid status etc. itself, even if leaked for a long time in the domain
		return optView.Unwrap(), nil
	}
}

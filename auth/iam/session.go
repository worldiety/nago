package iam

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
	"time"
)

type SessionRepository = data.Repository[Session, core.SessionID]

type Session struct {
	ID              core.SessionID
	User            std.Option[auth.UID]
	CreatedAt       time.Time
	AuthenticatedAt time.Time
}

func (s Session) Identity() core.SessionID {
	return s.ID
}

// Subject returns always a subject, e.g. an authenticated user or an API user or just an invalid
// subject, if the session is not known or login date is too old.
func (s *Service) Subject(id core.SessionID) auth.Subject {
	optSession, err := s.sessions.FindByID(id)
	if err != nil {
		slog.Error("failed to find session", "err", err)
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	if !optSession.Valid {
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	session := optSession.Unwrap()

	if !session.User.Valid {
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	const day = 24 * time.Hour
	if time.Now().Sub(session.AuthenticatedAt) > day*30*3 {
		slog.Error("session expired for user", "sessionID", session.ID, "user", session.User)
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	optUsr, err := s.users.FindByID(session.User.Unwrap())
	if err != nil {
		slog.Error("failed to find user", "err", err)
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	if !optUsr.Valid {
		// important, the user may have been deleted, but we are lazy, so we have them as stale references here
		return auth.InvalidSubject{
			DeniedLog: s.logInvalidSubject,
		}
	}

	usr := optUsr.Unwrap()

	// TODO collect permissions from groups and roles
	tmp := map[string]struct{}{}
	for _, permission := range usr.Permissions {
		tmp[permission] = struct{}{}
	}

	return newNagoSubject(id, usr, tmp, func(session core.SessionID, usr User, permission string, granted bool) {
		slog.Info("audit-log", "usecase", permission, "granted", granted, "sid", session, "uid", usr.ID, "email", usr.Email)
	})
}

func (s *Service) logInvalidSubject(permission string) {
	slog.Info("audit-log", "usecase", permission, "granted", false)
}

// Logout removes the associated user from the given session.
func (s *Service) Logout(id core.SessionID) bool {
	optSession, err := s.sessions.FindByID(id)
	if err != nil {
		slog.Error("failed to find session", "err", err)
		return false
	}

	if !optSession.Valid {
		return true
	}

	session := optSession.Unwrap()
	session.User = std.None[auth.UID]()
	if err := s.sessions.Save(session); err != nil {
		slog.Error("failed to save logout session", "err", err)
		return false
	}

	return true
}

// Login installs eventually the newly authenticated user into the given session.
func (s *Service) Login(id core.SessionID, login, password string) bool {
	// first install the session
	optSession, err := s.sessions.FindByID(id)
	if err != nil {
		slog.Error("failed to find session", "err", err)
		return false
	}

	var session Session
	if !optSession.Valid {
		session.ID = id
		session.CreatedAt = time.Now()
	} else {
		session = optSession.Unwrap()
	}

	// try to authenticate
	usr, err := s.authenticates(login, password)
	if err != nil {
		slog.Error("failed to authenticate user", "err", err)

		// we failed:
		// this may be an attacker which tried to reauthenticate for a privileged session
		// or the user just had a typo. however, we will log him out of the session for security.
		session.User = std.None[auth.UID]()
		if err := s.sessions.Save(session); err != nil {
			slog.Error("failed to save downgraded session", "err", err)
		}

		return false
	}

	// authentication seems fine
	session.User = std.Some(usr.ID)
	session.AuthenticatedAt = time.Now()

	if err := s.sessions.Save(session); err != nil {
		slog.Error("failed to save session", "err", err)
		return false
	}

	return true
}

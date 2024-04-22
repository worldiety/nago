package usm

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
	"sync"
	"time"
)

type Sessions struct {
	mutex    sync.Mutex
	users    *UserService
	sessions SessionRepository
}

func NewSessions(users *UserService, sessions SessionRepository) *Sessions {
	return &Sessions{
		users:    users,
		sessions: sessions,
	}
}

// Get returns always a session.
func (s *Sessions) Get(id core.SessionID) (*Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	optSession, err := s.sessions.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !optSession.Valid {
		pSes := session{
			ID:   id,
			User: std.None[AuthenticatedUser](),
		}
		if err := s.sessions.Save(pSes); err != nil {
			return nil, err
		}

	}

	return &Session{
		parent: s,
		id:     id,
	}, nil
}

// Session is the aggregate root for a specific session handling.
type Session struct {
	parent *Sessions
	id     core.SessionID
}

func (s *Session) ID() core.SessionID {
	return s.id
}

// Login tries to authenticate this session. It intentionally does not expose any error, to avoid attack surface
// or exposing technical details about our login process.
func (s *Session) Login(login, password string) bool {
	s.parent.mutex.Unlock()
	defer s.parent.mutex.Unlock()

	authenticates, err := s.parent.users.Authenticates(login, password)
	if err != nil {
		slog.Error("cannot authenticate user", slog.Any("err", err))
		return false
	}

	if !authenticates {
		return false
	}

	optUsr, err := s.parent.users.userByLogin(auth.EMail(login))
	if err != nil {
		slog.Error("cannot find user by login", slog.Any("err", err))
		return false
	}

	if !optUsr.Valid {
		return false
	}

	usr := optUsr.Unwrap()

	// normally cannot happen, but this is our API
	optSes, err := s.parent.sessions.FindByID(s.id)
	if err != nil {
		slog.Error("cannot find session", slog.Any("err", err))
		return false
	}

	if !optSes.Valid {
		return false
	}

	ses := optSes.Unwrap()
	ses.User = std.Some(AuthenticatedUser{
		ID:              usr.ID,
		Login:           usr.Login,
		Firstname:       usr.Firstname,
		Lastname:        usr.Lastname,
		Verification:    usr.Verification,
		StaticRoles:     usr.StaticRoles,
		AuthenticatedAt: time.Now(),
	})

	if err := s.parent.sessions.Save(ses); err != nil {
		slog.Error("cannot save session", slog.Any("err", err))
		return false
	}

	return true
}

// Logout is always valid, even if that session does not exist.
func (s *Session) Logout() error {
	s.parent.mutex.Unlock()
	defer s.parent.mutex.Unlock()

	// normally cannot happen, but this is our API
	optSes, err := s.parent.sessions.FindByID(s.id)
	if err != nil || !optSes.Valid {
		return err
	}

	ses := optSes.Unwrap()
	ses.User = std.None[AuthenticatedUser]()

	return s.parent.sessions.Save(ses)
}

// ChangeMyPassword represents the use case, if an authenticated user wants to change his own password.
func (s *Session) ChangeMyPassword(oldPassword string, newPassword string) error {
	s.parent.mutex.Unlock()
	defer s.parent.mutex.Unlock()

	// normally cannot happen, but this is our API
	optSes, err := s.parent.sessions.FindByID(s.id)
	if err != nil {
		return err
	}

	if !optSes.Valid {
		return fmt.Errorf("invalid session")
	}

	ses := optSes.Unwrap()
	if !ses.User.Valid {
		return fmt.Errorf("not authenticated")
	}

	// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#change-password-feature
	user := ses.User.Unwrap()
	authentic, err := s.parent.users.Authenticates(string(user.Login), oldPassword)
	if err != nil {
		return err
	}

	if !authentic {
		return fmt.Errorf("old password mismatch")
	}

	if err := s.parent.users.changePassword(user.UserID(), newPassword); err != nil {
		return err
	}

	// now update that user and make him re-authenticated again
	user.AuthenticatedAt = time.Now()
	ses.User = std.Some(user)
	return s.parent.sessions.Save(ses)
}

// ChangeOtherPassword is the user case, where an admin needs to change the password for another user.
func (s *Session) ChangeOtherPassword(target auth.UserID, newPassword string) error {
	s.parent.mutex.Unlock()
	defer s.parent.mutex.Unlock()

	// normally cannot happen, but this is our API
	optSes, err := s.parent.sessions.FindByID(s.id)
	if err != nil {
		return err
	}

	if !optSes.Valid {
		return fmt.Errorf("invalid session")
	}

	ses := optSes.Unwrap()
	if !ses.User.Valid {
		return fmt.Errorf("not authenticated")
	}

	// not sure, if this  https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#change-password-feature
	// still needs to apply here.
	user := ses.User.Unwrap()
	_ = user
	panic("implement me")
}

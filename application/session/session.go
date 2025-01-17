package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/tick"
	"log/slog"
	"sync"
	"time"
)

type sessionImpl struct {
	id              ID
	repo            Repository
	session         Session // cache
	mutex           sync.RWMutex
	refreshInterval time.Duration
	lastRefreshedAt time.Time
}

func (s *sessionImpl) refresh() Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.refreshInterval == 0 {
		s.refreshInterval = 5 * time.Minute
	}

	now := tick.Now(tick.Minute)
	if now.Sub(s.lastRefreshedAt) >= s.refreshInterval {
		s.session = s.load()
	}

	return s.session
}

func (s *sessionImpl) load() Session {
	optSess, err := s.repo.FindByID(s.id)
	if err != nil {
		slog.Error("failed to find session by id", "err", err, "id", s.id)
		return Session{}
	}

	return optSess.UnwrapOr(Session{})
}

func (s *sessionImpl) ID() ID {
	return s.id
}

func (s *sessionImpl) User() std.Option[user.ID] {
	return s.refresh().User
}

func (s *sessionImpl) CreatedAt() std.Option[time.Time] {
	v := s.refresh().AuthenticatedAt

	if v.IsZero() {
		return std.None[time.Time]()
	}

	return std.Some(s.session.CreatedAt)
}

func (s *sessionImpl) AuthenticatedAt() std.Option[time.Time] {
	v := s.refresh().AuthenticatedAt

	if v.IsZero() {
		return std.None[time.Time]()
	}

	return std.Some(v)
}

func (s *sessionImpl) PutString(key string, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	optSession, err := s.repo.FindByID(s.id)
	if err != nil {
		return fmt.Errorf("session: failed to find session by id: %w", err)
	}

	var session Session
	if optSession.IsNone() {
		session.ID = s.id
	} else {
		session = optSession.Unwrap()
	}

	if session.Values == nil {
		session.Values = map[string]string{}
	}

	session.Values[key] = value

	return s.repo.Save(session)
}

func (s *sessionImpl) GetString(key string) (string, bool) {
	v, ok := s.refresh().Values[key]
	return v, ok
}

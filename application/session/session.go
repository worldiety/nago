// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/std"
)

// nowFunc is the clock of the single sign-on refresh interval. Tests replace it.
var nowFunc = time.Now

// nlsRefreshInterval is the time between two single sign-on refreshes of a session.
const nlsRefreshInterval = 5 * time.Minute

// nlsRetryInterval is the time after which a refresh is retried, which failed because the login service has been
// unavailable, see [ErrNLSUnavailable].
const nlsRetryInterval = time.Minute

// startNLSRefresh runs a single sign-on refresh. It runs in the background, because an access happens within the
// event loop of a window, which must not wait for the login service. Tests replace it.
var startNLSRefresh = func(refresh func()) { go refresh() }

// sessionImpl reads its record from the repository on every access. It deliberately keeps no copy: a login, a
// logout or a deletion must be visible immediately, also to HTTP handlers, and reading a record costs only a few
// microseconds. Only the single sign-on refresh, which calls the login service, keeps an interval.
type sessionImpl struct {
	id         ID
	repo       Repository
	refreshNLS RefreshNLS

	mutex            sync.Mutex // guards lastNLSRefreshAt and nlsRefreshing and serializes PutString
	lastNLSRefreshAt time.Time  // the start of the last refresh, or a time in the past to retry earlier
	nlsRefreshing    bool
}

func newSessionImpl(id ID, repo Repository, refresh RefreshNLS) *sessionImpl {
	return &sessionImpl{id: id, repo: repo, refreshNLS: refresh}
}

// current returns the stored record, without a user if it has expired. A due single sign-on refresh is started in
// the background, thus the access never waits for the login service: a refresh which fails and logs the user out
// takes effect for the next access and publishes [LoggedOut] for the open windows.
func (s *sessionImpl) current() Session {
	session := s.load()
	if s.nlsRefreshDue(session) {
		startNLSRefresh(s.refreshInBackground)
		session = s.load()
	}

	if expired(session, nowFunc()) {
		session.User = std.None[user.ID]()
		session.AuthenticatedAt = time.Time{}
	}

	return session
}

// nlsRefreshDue reports whether the single sign-on session must be refreshed now and reserves the refresh, so
// that concurrent accesses do not refresh twice. A refresh is due an interval after the last start or the last
// success, which may have happened in another process before a restart or by the sign-in itself.
func (s *sessionImpl) nlsRefreshDue(session Session) bool {
	if session.RefreshToken == "" {
		return false
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.nlsRefreshing {
		return false
	}

	now := nowFunc()
	last := s.lastNLSRefreshAt
	if session.NLSRefreshedAt.After(last) {
		last = session.NLSRefreshedAt
	}

	if !last.IsZero() && now.Sub(last) < nlsRefreshInterval {
		return false
	}

	s.lastNLSRefreshAt = now
	s.nlsRefreshing = true
	return true
}

func (s *sessionImpl) refreshInBackground() {
	err := s.refreshNLS(s.id)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.nlsRefreshing = false
	switch {
	case err == nil:
	case errors.Is(err, ErrNLSUnavailable):
		// the session is kept for a while, see NewRefreshNLS, so retry earlier than usual
		s.lastNLSRefreshAt = nowFunc().Add(nlsRetryInterval - nlsRefreshInterval)
		slog.Warn("cannot refresh NLS session, will retry", "session", logging.Secret(string(s.id)), "err", err.Error())
	default:
		slog.Error("failed to refresh NLS session", "session", logging.Secret(string(s.id)), "err", err.Error())
	}
}

func (s *sessionImpl) load() Session {
	optSess, err := s.repo.FindByID(s.id)
	if err != nil {
		slog.Error("failed to find session by id", "err", err, "id", logging.Secret(string(s.id)))
		return Session{}
	}

	return optSess.UnwrapOr(Session{})
}

func (s *sessionImpl) ID() ID {
	return s.id
}

func (s *sessionImpl) User() std.Option[user.ID] {
	return s.current().User
}

func (s *sessionImpl) CreatedAt() std.Option[time.Time] {
	session := s.current()
	if session.AuthenticatedAt.IsZero() {
		return std.None[time.Time]()
	}

	return std.Some(session.CreatedAt)
}

func (s *sessionImpl) AuthenticatedAt() std.Option[time.Time] {
	v := s.current().AuthenticatedAt

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
	v, ok := s.current().Values[key]
	return v, ok
}

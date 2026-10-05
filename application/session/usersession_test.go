// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/user"
	datamem "go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func newTestSessions(refresh RefreshNLS) (Repository, FindUserSessionByID, LoginUser, Logout) {
	repo := &datamem.Repository[Session, ID]{}
	if refresh == nil {
		refresh = func(id ID) error { return nil }
	}

	bus := events.NewEventBus()
	return repo, NewFindUserSessionByID(repo, refresh), NewLoginUser(bus, repo), NewLogout(repo)
}

// A logout and a login as another user in the same browser keep the session id. The user of the session must
// follow immediately, otherwise HTTP handlers serve the previous user.
func TestUserSessionFollowsLoginAndLogout(t *testing.T) {
	_, find, login, logout := newTestSessions(nil)
	const sid = ID("browser-session")

	if find(sid).User().IsSome() {
		t.Fatal("a fresh session has no user")
	}

	if err := login(sid, "alice"); err != nil {
		t.Fatal(err)
	}

	if got := find(sid).User(); got.UnwrapOr("") != "alice" {
		t.Fatalf("expected alice, got %v", got)
	}

	if _, err := logout(sid); err != nil {
		t.Fatal(err)
	}

	if got := find(sid).User(); got.IsSome() {
		t.Fatalf("a logged out session must have no user, got %v", got)
	}

	if err := login(sid, "bob"); err != nil {
		t.Fatal(err)
	}

	if got := find(sid).User(); got.UnwrapOr("") != "bob" {
		t.Fatalf("expected bob, got %v", got)
	}
}

// Deleting all sessions and writes outside of the use cases are visible immediately.
func TestUserSessionFollowsRepository(t *testing.T) {
	repo, find, login, _ := newTestSessions(nil)
	const sid = ID("browser-session")

	if err := login(sid, "alice"); err != nil {
		t.Fatal(err)
	}

	s := find(sid)
	if s.User().IsNone() {
		t.Fatal("expected alice")
	}

	if err := NewClear(&sync.Mutex{}, repo)(); err != nil {
		t.Fatal(err)
	}

	if s.User().IsSome() {
		t.Fatal("a cleared session must have no user")
	}

	if err := s.PutString("k", "v"); err != nil {
		t.Fatal(err)
	}

	if v, _ := find(sid).GetString("k"); v != "v" {
		t.Fatalf("expected the stored value, got %q", v)
	}
}

// An expired session has no user, also for HTTP handlers, which never call FindSessionByID.
func TestUserSessionExpires(t *testing.T) {
	repo, find, _, _ := newTestSessions(nil)
	const sid = ID("browser-session")

	if err := repo.Save(Session{ID: sid, User: std.Some[user.ID]("alice"), AuthenticatedAt: time.Now().Add(-91 * 24 * time.Hour)}); err != nil {
		t.Fatal(err)
	}

	s := find(sid)
	if s.User().IsSome() || s.AuthenticatedAt().IsSome() {
		t.Fatal("an expired session must have no user")
	}
}

// The single sign-on refresh keeps its interval, although the session is read on every access, and a failed
// refresh, which logs the user out, takes effect immediately.
func TestUserSessionRefreshesNLSOncePerInterval(t *testing.T) {
	var calls atomic.Int32
	var refresh func(id ID) error
	repo, find, _, _ := newTestSessions(func(id ID) error {
		calls.Add(1)
		return refresh(id)
	})

	refresh = func(id ID) error {
		// the refresh saves the session, like the real one does with the rotated token
		opt, _ := repo.FindByID(id)
		s := opt.Unwrap()
		s.RefreshToken = "rotated"
		return repo.Save(s)
	}

	const sid = ID("sso-session")
	if err := repo.Save(Session{ID: sid, User: std.Some[user.ID]("alice"), AuthenticatedAt: time.Now(), RefreshToken: "token"}); err != nil {
		t.Fatal(err)
	}

	clock := time.Now()
	nowFunc = func() time.Time { return clock }
	defer func() { nowFunc = time.Now }()

	s := find(sid)
	for range 5 {
		if s.User().IsNone() {
			t.Fatal("expected alice")
		}
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one refresh within the interval, got %d", got)
	}

	clock = clock.Add(6 * time.Minute)
	s.User()
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected a refresh after the interval, got %d", got)
	}

	// a failed refresh logs the user out, which this very access must already see
	refresh = func(id ID) error {
		_, err := NewLogout(repo)(id)
		return err
	}

	clock = clock.Add(6 * time.Minute)
	if s.User().IsSome() {
		t.Fatal("the user logged out by the failed refresh must be gone immediately")
	}
}

// Concurrent first accesses share one instance, so a scope never holds an orphan.
func TestFindUserSessionSharesInstance(t *testing.T) {
	_, find, _, _ := newTestSessions(nil)

	var wg sync.WaitGroup
	got := make([]UserSession, 16)
	for i := range got {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got[i] = find("same")
		}()
	}

	wg.Wait()
	for _, s := range got[1:] {
		if s != got[0] {
			t.Fatal("concurrent first accesses created several instances")
		}
	}
}

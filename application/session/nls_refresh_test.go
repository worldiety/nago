// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"net/http"
	"testing"
	"time"

	"go.wdy.de/nago/application/user"
	datamem "go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

// signedIn signs the victim in and returns a clock, which the refresh interval follows.
func (f nlsFixture) signedIn(t *testing.T) (us UserSession, advance func(d time.Duration)) {
	t.Helper()
	if _, err := f.uc.ExchangeNLS(victimSession, f.start(t, victimSession)); err != nil {
		t.Fatal(err)
	}

	clock := time.Now()
	nowFunc = func() time.Time { return clock }
	t.Cleanup(func() { nowFunc = time.Now })

	return f.uc.FindUserSessionByID(victimSession), func(d time.Duration) { clock = clock.Add(d) }
}

// An access within the event loop of a window must never wait for the login service.
func TestUserSessionRefreshesInTheBackground(t *testing.T) {
	startNLSRefresh = func(refresh func()) { go refresh() }
	defer func() { startNLSRefresh = func(refresh func()) { refresh() } }()

	release := make(chan struct{})
	done := make(chan struct{}, 1)
	repo := &datamem.Repository[Session, ID]{}
	find := NewFindUserSessionByID(repo, func(id ID) error {
		<-release
		done <- struct{}{}
		return nil
	})

	const sid = ID("sso-session")
	if err := repo.Save(Session{ID: sid, User: std.Some[user.ID]("alice"), AuthenticatedAt: time.Now(), RefreshToken: "token"}); err != nil {
		t.Fatal(err)
	}

	s := find(sid)
	returned := make(chan bool, 1)
	go func() { returned <- s.User().IsSome() }()

	select {
	case ok := <-returned:
		if !ok {
			t.Fatal("the stored user must be served during the refresh")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the access waits for the login service")
	}

	// a running refresh is not started twice
	s.User()
	close(release)
	<-done
	select {
	case <-done:
		t.Fatal("the refresh ran twice")
	case <-time.After(50 * time.Millisecond):
	}
}

// The sign-in has just asked the login service, so the next refresh is due an interval later.
func TestNLSRefreshIsDueAnIntervalAfterTheSignIn(t *testing.T) {
	f := newNLSFixture(t)
	us, advance := f.signedIn(t)

	if us.User().IsNone() || f.refreshes.Load() != 1 {
		t.Fatalf("expected only the refresh of the sign-in, got %d", f.refreshes.Load())
	}

	advance(nlsRefreshInterval + time.Second)
	us.User()
	if f.refreshes.Load() != 2 {
		t.Fatalf("expected a refresh after the interval, got %d", f.refreshes.Load())
	}
}

// A rejected token logs the user out. An unavailable login service keeps the session for a while and retries
// earlier, but not forever.
func TestNLSRefreshDistinguishesRejectionFromOutage(t *testing.T) {
	t.Run("rejected", func(t *testing.T) {
		f := newNLSFixture(t)
		us, advance := f.signedIn(t)
		f.refreshStatus.Store(http.StatusBadRequest)
		advance(nlsRefreshInterval + time.Second)
		if us.User().IsSome() || mustSession(t, f.sessions, victimSession).RefreshToken != "" {
			t.Fatal("a rejected token must log the user out")
		}
	})

	t.Run("outage within the grace period", func(t *testing.T) {
		f := newNLSFixture(t)
		us, advance := f.signedIn(t)
		f.refreshStatus.Store(http.StatusServiceUnavailable)
		advance(nlsRefreshInterval + time.Second)
		if us.User().IsNone() {
			t.Fatal("an outage must not log the user out at once")
		}

		before := f.refreshes.Load()
		advance(nlsRetryInterval + time.Second)
		us.User()
		if f.refreshes.Load() != before+1 {
			t.Fatal("expected a retry after the retry interval")
		}

		f.refreshStatus.Store(0)
		advance(nlsRetryInterval + time.Second)
		if us.User().IsNone() || mustSession(t, f.sessions, victimSession).NLSRefreshedAt.IsZero() {
			t.Fatal("the session must recover with the login service")
		}
	})

	t.Run("outage beyond the grace period", func(t *testing.T) {
		f := newNLSFixture(t)
		us, advance := f.signedIn(t)
		f.refreshStatus.Store(http.StatusBadGateway)
		advance(nlsGracePeriod + time.Second)
		if us.User().IsSome() {
			t.Fatal("a session must not survive an outage forever")
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		f := newNLSFixture(t)
		us, advance := f.signedIn(t)
		f.server.Close()
		advance(nlsRefreshInterval + time.Second)
		if us.User().IsNone() {
			t.Fatal("an unreachable login service must not log the user out at once")
		}
	})

	t.Run("session of an older version", func(t *testing.T) {
		f := newNLSFixture(t)
		us, advance := f.signedIn(t)
		record := mustSession(t, f.sessions, victimSession)
		record.NLSRefreshedAt = time.Time{}
		if err := f.sessions.Save(record); err != nil {
			t.Fatal(err)
		}

		f.refreshStatus.Store(http.StatusServiceUnavailable)
		advance(nlsRefreshInterval + time.Second)
		if us.User().IsSome() {
			t.Fatal("a session without a successful refresh must be logged out like before")
		}
	})
}

// A logout while a refresh asks the login service must not be undone by the refresh.
func TestNLSRefreshKeepsALogoutInTheMeantime(t *testing.T) {
	f := newNLSFixture(t)
	us, advance := f.signedIn(t)

	logout := func() {
		if _, err := f.uc.Logout(victimSession); err != nil {
			t.Error(err)
		}
	}
	f.onMerge.Store(&logout)

	advance(nlsRefreshInterval + time.Second)
	us.User()
	f.onMerge.Store(nil)

	if us.User().IsSome() {
		t.Fatal("the refresh brought back a logged out user")
	}
}

// The avatar is loaded at the sign-in and then once a day, not with every refresh.
func TestNLSRefreshLoadsTheAvatarOnceADay(t *testing.T) {
	f := newNLSFixture(t)
	us, advance := f.signedIn(t)
	if f.photos.Load() != 1 {
		t.Fatalf("expected the avatar of the sign-in, got %d", f.photos.Load())
	}

	for range 3 {
		advance(nlsRefreshInterval + time.Second)
		us.User()
	}

	if f.photos.Load() != 1 || f.refreshes.Load() != 4 {
		t.Fatalf("expected refreshes without avatars, got %d avatars and %d refreshes", f.photos.Load(), f.refreshes.Load())
	}

	advance(nlsAvatarInterval)
	us.User()
	if f.photos.Load() != 2 {
		t.Fatalf("expected the avatar again after a day, got %d", f.photos.Load())
	}
}

// The open windows of a session follow a logout, see core.Application.
func TestLogoutPublishesLoggedOut(t *testing.T) {
	f := newNLSFixture(t)
	got := make(chan ID, 1)
	defer events.SubscribeFor[LoggedOut](f.bus, func(evt LoggedOut) { got <- evt.Session })()

	if _, err := f.uc.Logout(victimSession); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-got:
		if id != victimSession {
			t.Fatalf("unexpected session %s", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no LoggedOut has been published")
	}
}

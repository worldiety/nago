// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	datamem "go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/events"
)

type nlsFixture struct {
	uc        UseCases
	sessions  *datamem.Repository[Session, ID]
	nonces    *datamem.Repository[NLSNonceEntry, NLSNonce]
	bus       events.EventBus
	server    *httptest.Server
	exchanges *atomic.Int32
	refreshes *atomic.Int32
	photos    *atomic.Int32

	// refreshStatus is the status of the refresh endpoint, 0 means success
	refreshStatus *atomic.Int32

	// onMerge is invoked by the merge of the user, which runs outside the mutex
	onMerge *atomic.Pointer[func()]
}

// newNLSFixture fakes the login service, which signs in whoever completes the flow as the victim.
func newNLSFixture(t *testing.T) nlsFixture {
	var exchanges, refreshes, photos, refreshStatus atomic.Int32
	var onMerge atomic.Pointer[func()]
	nls := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/nago/v1/exchange":
			exchanges.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]string{"refresh": "victim-refresh"})
		case "/api/nago/v1/refresh":
			refreshes.Add(1)
			if status := refreshStatus.Load(); status != 0 {
				w.WriteHeader(int(status))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"exp": 0, "user": map[string]string{"id": "v", "mail": "victim@example.com"}})
		case "/api/nago/v1/photo/me":
			photos.Add(1)
			_, _ = w.Write([]byte("png"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(nls.Close)

	load := settings.LoadGlobal(func(_ permission.Auditable, _ reflect.Type) (settings.GlobalSettings, error) {
		return user.Settings{SSONLSServer: nls.URL}, nil
	})
	merge := user.MergeSingleSignOnUser(func(u user.SingleSignOnUser, _ []byte) (user.ID, error) {
		if f := onMerge.Load(); f != nil {
			(*f)()
		}
		return user.ID("uid-" + string(u.Email)), nil
	})

	nonces := &datamem.Repository[NLSNonceEntry, NLSNonce]{}
	sessions := &datamem.Repository[Session, ID]{}
	bus := events.NewEventBus()
	uc := NewUseCases(bus, "/account/nls/authentication", load, merge, sessions, nonces, nil)
	return nlsFixture{uc: uc, sessions: sessions, nonces: nonces, bus: bus, server: nls, exchanges: &exchanges, refreshes: &refreshes, photos: &photos, refreshStatus: &refreshStatus, onMerge: &onMerge}
}

func (f nlsFixture) start(t *testing.T, sid ID) NLSNonce {
	t.Helper()
	uri, err := f.uc.StartNLSFlow(sid)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}

	return NLSNonce(u.Query().Get("nonce"))
}

const (
	attackerSession = ID("attacker-session-xxxxxxxxxxxxxxxxxxxx")
	victimSession   = ID("victim-session-xxxxxxxxxxxxxxxxxxxxxx")
)

// An attacker starts the flow and lets a victim complete it with the link of the login service. The victim's
// sign-in must not end up in the attacker's session.
func TestExchangeNLSRejectsAnotherSession(t *testing.T) {
	f := newNLSFixture(t)
	nonce := f.start(t, attackerSession)

	if _, err := f.uc.ExchangeNLS(victimSession, nonce); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("expected the exchange of another session to be rejected, got %v", err)
	}

	if got := f.uc.FindUserSessionByID(attackerSession).User(); got.IsSome() {
		t.Fatalf("the attacker's session must not be signed in, got %v", got)
	}

	if f.exchanges.Load() != 0 {
		t.Fatal("the login service must not be asked for a foreign nonce")
	}

	// the nonce is spent, the attacker cannot complete the flow either
	if _, err := f.uc.ExchangeNLS(attackerSession, nonce); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected the nonce to be spent, got %v", err)
	}
}

// The session which started the flow is signed in, but only once per nonce.
func TestExchangeNLSSignsInOnce(t *testing.T) {
	f := newNLSFixture(t)
	nonce := f.start(t, victimSession)

	if _, err := f.uc.ExchangeNLS(victimSession, nonce); err != nil {
		t.Fatal(err)
	}

	if got := f.uc.FindUserSessionByID(victimSession).User(); got.UnwrapOr("") != "uid-victim@example.com" {
		t.Fatalf("expected the victim to be signed in, got %v", got)
	}

	if _, err := f.uc.ExchangeNLS(victimSession, nonce); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a replayed nonce must be rejected, got %v", err)
	}

	if f.exchanges.Load() != 1 {
		t.Fatalf("expected one exchange, got %d", f.exchanges.Load())
	}
}

// A sign-in which has not been completed in time, and nonces of older versions without a time, are rejected.
func TestExchangeNLSRejectsExpiredNonces(t *testing.T) {
	f := newNLSFixture(t)

	for name, createdAt := range map[string]time.Time{
		"expired": time.Now().Add(-NLSNonceLifetime - time.Minute),
		"legacy":  {},
	} {
		nonce := NLSNonce("nonce-" + name)
		if err := f.nonces.Save(NLSNonceEntry{ID: nonce, Session: victimSession, CreatedAt: createdAt}); err != nil {
			t.Fatal(err)
		}

		if _, err := f.uc.ExchangeNLS(victimSession, nonce); !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("%s: expected the nonce to be rejected, got %v", name, err)
		}
	}

	if f.exchanges.Load() != 0 || f.uc.FindUserSessionByID(victimSession).User().IsSome() {
		t.Fatal("an expired nonce must not sign in")
	}
}

// A refresh every few minutes keeps the time of the sign-in, so a single sign-on session reaches its lifetime like
// any other session.
func TestNLSRefreshKeepsTheSignInTime(t *testing.T) {
	f := newNLSFixture(t)
	if _, err := f.uc.ExchangeNLS(victimSession, f.start(t, victimSession)); err != nil {
		t.Fatal(err)
	}

	signedIn := time.Now().Add(-sessionLifetime + 24*time.Hour)
	s, _ := f.sessions.FindByID(victimSession)
	record := s.Unwrap()
	record.AuthenticatedAt = signedIn
	if err := f.sessions.Save(record); err != nil {
		t.Fatal(err)
	}

	clock := time.Now()
	nowFunc = func() time.Time { return clock }
	defer func() { nowFunc = time.Now }()

	us := f.uc.FindUserSessionByID(victimSession)
	if us.User().IsNone() {
		t.Fatal("the session must be valid within its lifetime")
	}

	if got := mustSession(t, f.sessions, victimSession).AuthenticatedAt; !got.Equal(signedIn) {
		t.Fatalf("the refresh moved the sign-in time from %v to %v", signedIn, got)
	}

	// two days later, the lifetime is over, although refreshes keep succeeding
	clock = clock.Add(48 * time.Hour)
	if us.User().IsSome() {
		t.Fatal("the session must expire after its lifetime")
	}
}

func mustSession(t *testing.T, repo *datamem.Repository[Session, ID], id ID) Session {
	t.Helper()
	opt, err := repo.FindByID(id)
	if err != nil || opt.IsNone() {
		t.Fatalf("session %s not found: %v", id, err)
	}

	return opt.Unwrap()
}

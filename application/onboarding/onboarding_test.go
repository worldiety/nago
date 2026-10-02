// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package onboarding

import (
	"errors"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
)

type fakeUsers struct {
	mutex    sync.Mutex
	users    map[user.ID]user.User
	roles    map[user.ID][]role.ID
	logins   map[session.ID]user.ID
	codes    []string
	events   []any
	countErr error
}

func (f *fakeUsers) deps() Deps {
	return Deps{
		CountUsers: func() (int, error) {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			return len(f.users), f.countErr
		},
		Create: func(subject permission.Auditable, model user.ShortRegistrationUser) (user.User, error) {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			if model.Password == "" || model.Password != model.PasswordRepeated {
				return user.User{}, errors.New("bad password")
			}
			usr := user.User{ID: user.ID("u" + string(rune('0'+len(f.users)))), Email: model.Email, EMailVerified: model.Verified}
			f.users[usr.ID] = usr
			return usr, nil
		},
		UpdateVerification: func(subject permission.Auditable, id user.ID, verified bool) error { return nil },
		UpdateOtherRoles: func(subject user.AuditableUser, id user.ID, roles []role.ID) error {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			f.roles[id] = roles
			return nil
		},
		UpdateOtherGroups:      func(subject user.AuditableUser, id user.ID, groups []group.ID) error { return nil },
		UpdateOtherPermissions: func(subject user.AuditableUser, id user.ID, permissions []permission.ID) error { return nil },
		Delete: func(subject permission.Auditable, id user.ID) error {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			delete(f.users, id)
			return nil
		},
		LoginUser: func(id session.ID, usr user.ID) error {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			f.logins[id] = usr
			return nil
		},
		Deliver: func(to user.Email, code string, validUntil time.Time) error {
			f.mutex.Lock()
			defer f.mutex.Unlock()
			f.codes = append(f.codes, code)
			return nil
		},
	}
}

func (f *fakeUsers) lastCode() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.codes[len(f.codes)-1]
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time      { return c.t }
func (c *clock) add(d time.Duration) { c.t = c.t.Add(d) }

func newTestSetup(t *testing.T, opts Options) (*setup, *fakeUsers, *clock) {
	t.Helper()
	if opts.Email == "" {
		opts.Email = " Torben@Example.com "
	}

	f := &fakeUsers{users: map[user.ID]user.User{}, roles: map[user.ID][]role.ID{}, logins: map[session.ID]user.ID{}}
	s := newSetup(opts, f.deps())
	c := &clock{t: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	s.now = c.now
	return s, f, c
}

var profile = Profile{Firstname: "Torben", Lastname: "Schinke", Password: "secret!", PasswordRepeated: "secret!"}

func TestState(t *testing.T) {
	for raw, want := range map[string]Problem{"": NoProblem, "  ": NoEmail, "no-mail": InvalidEmail} {
		s, _, _ := newTestSetup(t, Options{Email: raw})
		if raw == "" {
			s, _, _ = newTestSetup(t, Options{}) // the default address
		} else {
			s.email, s.issue = parseEmail(raw)
		}

		st := s.state()
		if !st.Pending || st.Problem != want {
			t.Fatalf("%q: unexpected state %+v", raw, st)
		}

		if want == NoProblem && (st.Email != "torben@example.com" || st.MaskedEmail() != "t***@example.com") {
			t.Fatalf("unexpected address %q %q", st.Email, st.MaskedEmail())
		}

		if want != NoProblem {
			if _, err := s.requestCode("s1"); !errors.Is(err, ErrNoEmail) {
				t.Fatalf("%q: expected no email, got %v", raw, err)
			}
		}
	}
}

func TestCompleteFlow(t *testing.T) {
	s, f, _ := newTestSetup(t, Options{Roles: []role.ID{"admin"}})

	if _, err := s.complete("s1", profile); !errors.Is(err, ErrNotVerified) {
		t.Fatalf("expected not verified, got %v", err)
	}

	if err := s.verifyCode("s1", "000000"); !errors.Is(err, ErrNoCode) {
		t.Fatalf("expected no code, got %v", err)
	}

	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	code := f.lastCode()
	if len(code) != 6 {
		t.Fatalf("unexpected code %q", code)
	}

	if err := s.verifyCode("s1", wrong(code)); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected an invalid code, got %v", err)
	}

	if c, ok := s.currentCode(); !ok || c.AttemptsLeft != 4 {
		t.Fatalf("unexpected code info %+v", c)
	}

	if err := s.verifyCode("s1", code[:3]+" "+code[3:]); err != nil {
		t.Fatalf("a code with a space must be accepted: %v", err)
	}

	// a code confirms one session only
	if err := s.verifyCode("s2", code); !errors.Is(err, ErrNoCode) {
		t.Fatalf("the code has been used twice: %v", err)
	}

	usr, err := s.complete("s1", profile)
	if err != nil {
		t.Fatal(err)
	}

	if usr.Email != "torben@example.com" || !usr.EMailVerified || f.roles[usr.ID][0] != "admin" || f.logins["s1"] != usr.ID {
		t.Fatalf("the user has not been set up: %+v roles=%v logins=%v", usr, f.roles, f.logins)
	}

	if s.state().Pending {
		t.Fatal("the setup must be done")
	}

	if _, err := s.complete("s1", profile); !errors.Is(err, ErrNotPending) {
		t.Fatalf("expected not pending, got %v", err)
	}
}

func TestCodeLimits(t *testing.T) {
	s, f, c := newTestSetup(t, Options{MaxCodesPerHour: 3})

	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	var cooldown CooldownError
	if _, err := s.requestCode("s2"); !errors.As(err, &cooldown) || !cooldown.RetryAt.Equal(c.t.Add(time.Minute)) {
		t.Fatalf("expected the cooldown, got %v", err)
	}

	// wrong codes lock the code
	code := f.lastCode()
	for i := 0; i < 4; i++ {
		if err := s.verifyCode("s1", wrong(code)); !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("attempt %d: expected an invalid code, got %v", i, err)
		}
	}

	if err := s.verifyCode("s1", wrong(code)); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected the lock, got %v", err)
	}

	if err := s.verifyCode("s1", code); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("a locked code must not be accepted, got %v", err)
	}

	// a new code after the cooldown replaces the locked one
	c.add(time.Minute)
	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	c.add(time.Minute)
	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	// the hourly limit
	c.add(time.Minute)
	if _, err := s.requestCode("s1"); !errors.As(err, &cooldown) {
		t.Fatalf("expected the hourly limit, got %v", err)
	}

	// expiry
	if err := s.verifyCode("s1", f.lastCode()); err != nil {
		t.Fatal(err)
	}

	c.add(time.Hour)
	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	c.add(16 * time.Minute)
	if err := s.verifyCode("s1", f.lastCode()); !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("expected an expired code, got %v", err)
	}

	// the confirmation of a session expires as well
	if s.isVerified("s1") {
		t.Fatal("the confirmation must have expired")
	}
}

func TestFailedConfigurationRemovesUser(t *testing.T) {
	s, f, _ := newTestSetup(t, Options{ConfigurePermissions: func(usr user.User) error { return errors.New("boom") }})
	if _, err := s.requestCode("s1"); err != nil {
		t.Fatal(err)
	}

	if err := s.verifyCode("s1", f.lastCode()); err != nil {
		t.Fatal(err)
	}

	if _, err := s.complete("s1", profile); err == nil {
		t.Fatal("expected the configuration error")
	}

	if len(f.users) != 0 || !s.state().Pending {
		t.Fatal("the half configured user must be removed and the setup must stay pending")
	}

	// a wrong profile keeps the confirmation, so the visitor can correct it
	if _, err := s.complete("s1", Profile{Password: "a", PasswordRepeated: "b"}); err == nil {
		t.Fatal("expected the password error")
	}

	if !s.isVerified("s1") {
		t.Fatal("the confirmation must survive a wrong profile")
	}
}

func TestNoSetupIfUsersCannotBeCounted(t *testing.T) {
	s, f, _ := newTestSetup(t, Options{})
	f.countErr = errors.New("broken store")

	if s.state().Pending {
		t.Fatal("the setup must fail closed")
	}

	f.countErr = nil
	if !s.state().Pending {
		t.Fatal("the setup must be offered, once the users can be counted")
	}
}

func wrong(code string) string {
	if code[0] == '9' {
		return "0" + code[1:]
	}

	return string(code[0]+1) + code[1:]
}

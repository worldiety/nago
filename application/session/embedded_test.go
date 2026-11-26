// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"errors"
	"fmt"
	"testing"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
)

// An embedded id is recognized as such, a plain one or an id, which only contains the marker, is not.
func TestEmbeddedID(t *testing.T) {
	a, b := NewEmbeddedID(), NewEmbeddedID()
	for _, row := range []struct {
		name string
		id   ID
		want bool
	}{
		{name: "new embedded id", id: a, want: true},
		{name: "another new embedded id", id: b, want: true},
		{name: "plain id", id: ID("ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12"), want: false},
		{name: "marker inside", id: ID("x-" + string(a)), want: false},
		{name: "empty", id: "", want: false},
	} {
		if got := row.id.Embedded(); got != row.want {
			t.Errorf("%s: %q.Embedded() = %v, want %v", row.name, row.id, got, row.want)
		}
	}

	if a == b || len(a) < len(embeddedPrefix)+64 {
		t.Errorf("embedded ids must be random and as long as a plain one: %q, %q", a, b)
	}
}

// A session of an embedded page never logs in: a login by password, by user id and the start of a single sign-on
// fail with ErrEmbedded and store nothing for the session, while the same calls succeed for a plain session.
func TestEmbeddedSessionCannotLogIn(t *testing.T) {
	f := newNLSFixture(t)
	uc := f.uc
	uc.Login = NewLogin(f.bus, f.sessions, func(login user.Email, password user.Password) (std.Option[user.User], error) {
		return std.Some(user.User{ID: "alice"}), nil
	})

	// stored reports whether the session has a stored user or a pending single sign-on
	stored := func(id ID) bool {
		optSession, err := f.sessions.FindByID(id)
		if err != nil {
			t.Fatal(err)
		}

		if optSession.IsSome() && optSession.Unwrap().User.IsSome() {
			return true
		}

		for entry, err := range f.nonces.All() {
			if err != nil {
				t.Fatal(err)
			}

			if entry.Session == id {
				return true
			}
		}

		return false
	}

	attempts := []struct {
		name  string
		login func(id ID) error
	}{
		{name: "login by password", login: func(id ID) error {
			ok, err := uc.Login(id, "alice@example.com", "secret")
			if err == nil && !ok {
				return errors.New("not authenticated")
			}
			return err
		}},
		{name: "login by user id", login: func(id ID) error { return uc.LoginUser(id, "alice") }},
		{name: "start of a single sign-on", login: func(id ID) error {
			_, err := uc.StartNLSFlow(id)
			return err
		}},
	}

	for i, a := range attempts {
		// fresh sessions per attempt, thus no attempt sees the effect of another one
		plain := ID(fmt.Sprintf("plain-session-%d-xxxxxxxxxxxxxxxxxxxxxx", i))
		embedded := NewEmbeddedID()
		for _, row := range []struct {
			name    string
			id      ID
			refused bool
		}{
			{name: a.name + ", plain session", id: plain, refused: false},
			{name: a.name + ", embedded session", id: embedded, refused: true},
		} {
			err := a.login(row.id)
			switch {
			case row.refused && !errors.Is(err, ErrEmbedded):
				t.Errorf("%s: want ErrEmbedded, got %v", row.name, err)
			case !row.refused && err != nil:
				t.Errorf("%s: want success, got %v", row.name, err)
			}

			if got := stored(row.id); got == row.refused {
				t.Errorf("%s: stored a user or a sign-on %v, want %v", row.name, got, !row.refused)
			}

			if got := uc.FindUserSessionByID(row.id).User(); row.refused && got.IsSome() {
				t.Errorf("%s: the session has the user %v", row.name, got)
			}
		}
	}
}

// A user, which has been stored for a session of an embedded page by another way than a login, is not reported,
// while the same record of a plain session is.
func TestEmbeddedSessionHidesAStoredUser(t *testing.T) {
	repo, find, _, _ := newTestSessions(nil)
	for _, row := range []struct {
		name string
		id   ID
		want std.Option[user.ID]
	}{
		{name: "plain session", id: ID("plain-session-xxxxxxxxxxxxxxxxxxxxxxxx"), want: std.Some[user.ID]("alice")},
		{name: "embedded session", id: NewEmbeddedID(), want: std.None[user.ID]()},
	} {
		if err := repo.Save(Session{ID: row.id, User: std.Some[user.ID]("alice"), AuthenticatedAt: nowFunc()}); err != nil {
			t.Fatal(err)
		}

		s := find(row.id)
		if got := s.User(); got != row.want {
			t.Errorf("%s: user %v, want %v", row.name, got, row.want)
		}

		if got := s.AuthenticatedAt(); got.IsSome() != row.want.IsSome() {
			t.Errorf("%s: authenticated at %v, want a time %v", row.name, got, row.want.IsSome())
		}
	}
}

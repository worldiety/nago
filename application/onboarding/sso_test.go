// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package onboarding

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sync"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"golang.org/x/text/language"
)

func TestParseAccessEntries(t *testing.T) {
	tests := []struct {
		text    string
		want    string
		invalid string
	}{
		{"anna@firma.de, bob@firma.de", "[anna@firma.de bob@firma.de]", "[]"},
		// pasted from a mail program, with names and odd separators
		{"Anna Muster <Anna@Firma.de>; Bob (bob@x.de)\n\tcarl@y.de.", "[anna@firma.de bob@x.de carl@y.de]", "[]"},
		{"firma.de\n@zweite-firma.com   www.dritte.de https://vierte.de/kontakt", "[@firma.de @zweite-firma.com @dritte.de @vierte.de]", "[]"},
		{"mailto:c@d.de a@b.de A@B.DE a@b.de", "[c@d.de a@b.de]", "[]"},
		{"und alle von firma.de bitte", "[@firma.de]", "[]"},
		{"anna@ firma@ anna@firma @-x.de", "[]", "[{anna@ false} {firma@ false} {anna@firma false} {@-x.de false}]"},
		// a public provider must not be allowed as a whole, but its single addresses
		{"gmail.com @web.de max@gmail.com", "[max@gmail.com]", "[{gmail.com true} {@web.de true}]"},
		{"", "[]", "[]"},
	}

	for _, tt := range tests {
		entries, invalid := ParseAccessEntries(tt.text)
		if got := fmt.Sprint(entries); got != tt.want {
			t.Errorf("%q: got %s, want %s", tt.text, got, tt.want)
		}

		if got := fmt.Sprint(invalid); got != tt.invalid {
			t.Errorf("%q: got invalid %s, want %s", tt.text, got, tt.invalid)
		}
	}
}

func TestAllowList(t *testing.T) {
	admin := user.Email("torben@example.com")
	selected, _ := ParseAccessEntries("anna@firma.de, firma2.de, torben@example.com")
	tests := []struct {
		access  SSOAccess
		admin   user.Email
		allowed []string
		denied  []string
	}{
		{SSOAccess{Kind: SSOOnlyAdmin}, admin, []string{"torben@example.com"}, []string{"anna@example.com", "xtorben@example.com"}},
		{SSOAccess{Kind: SSODomain}, admin, []string{"torben@example.com", "anna@example.com"}, []string{"anna@sub.example.com", "anna@example.com.evil.de", "anna@exampleXcom"}},
		// the domain of a public provider is never allowed as a whole
		{SSOAccess{Kind: SSODomain}, "torben@gmail.com", []string{"torben@gmail.com"}, []string{"anna@gmail.com"}},
		{SSOAccess{Kind: SSOSelected, Entries: selected}, admin, []string{"torben@example.com", "anna@firma.de", "bob@firma2.de"}, []string{"bob@firma.de", "anna@example.com"}},
		{SSOAccess{Kind: SSOEveryone}, admin, nil, nil},
	}

	for i, tt := range tests {
		list := tt.access.AllowList(tt.admin)
		if tt.access.Kind == SSOEveryone && len(list) != 0 {
			t.Fatalf("%d: everybody needs an empty list: %v", i, list)
		}

		match := func(mail string) bool {
			for _, p := range list {
				if regexp.MustCompile(p).MatchString(mail) {
					return true
				}
			}

			return false
		}

		for _, mail := range tt.allowed {
			if !match(mail) {
				t.Errorf("%d: %s must be allowed by %v", i, mail, list)
			}
		}

		for _, mail := range tt.denied {
			if match(mail) {
				t.Errorf("%d: %s must be denied by %v", i, mail, list)
			}
		}
	}

	if list := (SSOAccess{Kind: SSOSelected, Entries: selected}).AllowList(admin); len(list) != 3 {
		t.Fatalf("the address of the first user must not be listed twice: %v", list)
	}
}

type fakeSSO struct {
	mutex    sync.Mutex
	global   map[reflect.Type]settings.GlobalSettings
	flows    []session.ID
	startErr error
}

func newSSOSetup(t *testing.T, opts Options) (*setup, *fakeUsers, *fakeSSO) {
	t.Helper()
	original := opts
	s, f, _ := newTestSetup(t, opts)
	if original.Email == "" {
		original.Email = s.opts.Email
	}

	sso := &fakeSSO{global: map[reflect.Type]settings.GlobalSettings{}}
	deps := f.deps()
	deps.LoadGlobal = func(subject permission.Auditable, t reflect.Type) (settings.GlobalSettings, error) {
		sso.mutex.Lock()
		defer sso.mutex.Unlock()
		if v, ok := sso.global[t]; ok {
			return v, nil
		}

		return reflect.New(t).Elem().Interface().(settings.GlobalSettings), nil
	}
	deps.StoreGlobal = func(subject permission.Auditable, v settings.GlobalSettings) error {
		sso.mutex.Lock()
		defer sso.mutex.Unlock()
		sso.global[reflect.TypeOf(v)] = v
		return nil
	}
	deps.StartNLSFlow = func(id session.ID) (string, error) {
		sso.mutex.Lock()
		defer sso.mutex.Unlock()
		sso.flows = append(sso.flows, id)
		return "https://nls.example/?nonce=n1", sso.startErr
	}
	deps.PendingSSO = json.NewSloppyJSONRepository[PendingSSO, PendingSSOID](mem.NewBlobStore("pending"))

	s2 := newSetup(original, deps)
	s2.now = s.now
	return s2, f, sso
}

func (sso *fakeSSO) userSettings() user.Settings {
	sso.mutex.Lock()
	defer sso.mutex.Unlock()
	v, _ := sso.global[reflect.TypeFor[user.Settings]()].(user.Settings)
	return v
}

func verified(t *testing.T, s *setup, f *fakeUsers, sid session.ID) {
	t.Helper()
	if _, err := s.requestCode(sid); err != nil {
		t.Fatal(err)
	}

	if err := s.verifyCode(sid, f.lastCode()); err != nil {
		t.Fatal(err)
	}
}

// The first user signs in by the login service instead of a password. Until it has signed in, nobody else may.
func TestCompleteWithSSO(t *testing.T) {
	s, f, sso := newSSOSetup(t, Options{Roles: []role.ID{"admin"}})
	if st := s.state(); st.SSOServer != DefaultNLSServer {
		t.Fatalf("expected the default login service, got %q", st.SSOServer)
	}

	if _, err := s.completeWithSSO("s1", SSOAccess{Kind: SSODomain}, language.English); !errors.Is(err, ErrNotVerified) {
		t.Fatalf("expected not verified, got %v", err)
	}

	verified(t, s, f, "s1")
	loginURL, err := s.completeWithSSO("s1", SSOAccess{Kind: SSODomain}, language.English)
	if err != nil || loginURL != "https://nls.example/?nonce=n1" || !slices.Equal(sso.flows, []session.ID{"s1"}) {
		t.Fatalf("expected the login service for s1, got %q %v %v", loginURL, sso.flows, err)
	}

	usr := f.users["u0"]
	if usr.Email != "torben@example.com" || !usr.RequirePasswordChange || !slices.Equal(f.roles["u0"], []role.ID{"admin"}) {
		t.Fatalf("unexpected first user %+v %v", usr, f.roles)
	}

	if len(f.logins) != 0 {
		t.Fatalf("the session must sign in by the login service: %v", f.logins)
	}

	gate := []string{`^torben@example\.com$`}
	if st := sso.userSettings(); st.SSONLSServer != DefaultNLSServer || !slices.Equal(st.SSOAllowList, gate) {
		t.Fatalf("unexpected settings %+v", st)
	}

	if st := s.state(); st.Pending || st.SSOServer != "" {
		t.Fatalf("the setup must be done: %+v", st)
	}

	// another user, which has signed in somehow, does not open the single sign-on
	if err := s.applySSOAccess("other"); err != nil || !slices.Equal(sso.userSettings().SSOAllowList, gate) {
		t.Fatalf("unexpected access %v %v", sso.userSettings().SSOAllowList, err)
	}

	if err := s.applySSOAccess("u0"); err != nil {
		t.Fatal(err)
	}

	if list := sso.userSettings().SSOAllowList; !slices.Equal(list, []string{`^[^@]+@example\.com$`}) {
		t.Fatalf("expected the domain, got %v", list)
	}

	if opt, _ := s.deps.PendingSSO.FindByID(PendingSSOKey); opt.IsSome() {
		t.Fatal("the pending access must be gone")
	}

	if err := s.applySSOAccess("u0"); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteWithSSOVariants(t *testing.T) {
	t.Run("a configured login service takes precedence", func(t *testing.T) {
		s, f, sso := newSSOSetup(t, Options{})
		sso.global[reflect.TypeFor[user.Settings]()] = user.Settings{SSONLSServer: "https://nls.custom", DefaultSSORoles: []role.ID{"r"}}
		if st := s.state(); st.SSOServer != "https://nls.custom" {
			t.Fatalf("got %q", st.SSOServer)
		}

		verified(t, s, f, "s1")
		if _, err := s.completeWithSSO("s1", SSOAccess{Kind: SSOOnlyAdmin}, language.German); err != nil {
			t.Fatal(err)
		}

		if st := sso.userSettings(); st.SSONLSServer != "https://nls.custom" || !slices.Equal(st.DefaultSSORoles, []role.ID{"r"}) {
			t.Fatalf("the other settings must be kept: %+v", st)
		}

		// only the first user, thus nothing is pending
		if opt, _ := s.deps.PendingSSO.FindByID(PendingSSOKey); opt.IsSome() {
			t.Fatal("nothing must be pending")
		}
	})

	t.Run("everybody", func(t *testing.T) {
		s, f, sso := newSSOSetup(t, Options{})
		verified(t, s, f, "s1")
		if _, err := s.completeWithSSO("s1", SSOAccess{Kind: SSOEveryone}, language.German); err != nil {
			t.Fatal(err)
		}

		if err := s.applySSOAccess("u0"); err != nil || len(sso.userSettings().SSOAllowList) != 0 {
			t.Fatalf("expected an empty allow list: %v %v", sso.userSettings().SSOAllowList, err)
		}
	})

	t.Run("an allow list of an administrator wins", func(t *testing.T) {
		s, f, sso := newSSOSetup(t, Options{})
		verified(t, s, f, "s1")
		if _, err := s.completeWithSSO("s1", SSOAccess{Kind: SSOEveryone}, language.German); err != nil {
			t.Fatal(err)
		}

		changed := sso.userSettings()
		changed.SSOAllowList = []string{"^x@y\\.de$"}
		sso.global[reflect.TypeFor[user.Settings]()] = changed
		if err := s.applySSOAccess("u0"); err != nil || !slices.Equal(sso.userSettings().SSOAllowList, changed.SSOAllowList) {
			t.Fatalf("the list of the administrator must be kept: %v %v", sso.userSettings().SSOAllowList, err)
		}

		if opt, _ := s.deps.PendingSSO.FindByID(PendingSSOKey); opt.IsSome() {
			t.Fatal("the pending access must be gone")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		s, f, _ := newSSOSetup(t, Options{DisableSSO: true})
		if st := s.state(); st.SSOServer != "" {
			t.Fatalf("got %q", st.SSOServer)
		}

		verified(t, s, f, "s1")
		if _, err := s.completeWithSSO("s1", SSOAccess{}, language.German); !errors.Is(err, ErrNoSSO) {
			t.Fatalf("expected no sso, got %v", err)
		}

		// without the dependencies, there is no single sign-on as well
		s, _, _ = newTestSetup(t, Options{})
		if st := s.state(); st.SSOServer != "" {
			t.Fatalf("got %q", st.SSOServer)
		}
	})

	t.Run("the login service cannot be started", func(t *testing.T) {
		s, f, sso := newSSOSetup(t, Options{})
		sso.startErr = errors.New("nonce store down")
		verified(t, s, f, "s1")
		if _, err := s.completeWithSSO("s1", SSOAccess{}, language.German); err == nil {
			t.Fatal("expected an error")
		}

		// the account exists and the login page offers the single sign-on
		if st := s.state(); st.Pending || len(f.users) != 1 || sso.userSettings().SSONLSServer == "" {
			t.Fatalf("expected the account and the login service: %+v %v", st, f.users)
		}
	})
}

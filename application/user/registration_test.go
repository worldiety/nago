// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

type sentLink struct {
	to   Email
	id   RegistrationID
	code string
}

type registrationFixture struct {
	reg        *registration
	repo       RegistrationRepository
	settings   Settings
	used       map[Email]bool
	links      []sentLink
	registered []Email
	created    []ShortRegistrationUser
	createErr  error
	deliverErr error
	now        time.Time
}

func newRegistrationFixture(t *testing.T) *registrationFixture {
	t.Helper()

	f := &registrationFixture{
		repo:     RegistrationRepository(json.NewSloppyJSONRepository[RegistrationRequest, RegistrationID](mem.NewBlobStore("registration"))),
		settings: Settings{SelfRegistration: true},
		used:     map[Email]bool{},
		now:      time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC),
	}

	f.reg = newRegistration(RegistrationOptions{}, RegistrationDeps{
		Repository: f.repo,
		LoadGlobal: func(subject permission.Auditable, t reflect.Type) (settings.GlobalSettings, error) {
			return f.settings, nil
		},
		EMailUsed: func(email Email) (bool, error) {
			return f.used[email], nil
		},
		Create: func(subject permission.Auditable, model ShortRegistrationUser) (User, error) {
			if f.createErr != nil {
				return User{}, f.createErr
			}
			f.created = append(f.created, model)
			f.used[model.Email] = true
			return User{ID: ID("u" + string(model.Email)), Email: model.Email}, nil
		},
		Deliver: func(to Email, id RegistrationID, code string, validUntil time.Time) error {
			if f.deliverErr != nil {
				return f.deliverErr
			}
			f.links = append(f.links, sentLink{to: to, id: id, code: code})
			return nil
		},
		DeliverRegistered: func(to Email) error {
			if f.deliverErr != nil {
				return f.deliverErr
			}
			f.registered = append(f.registered, to)
			return nil
		},
	})
	f.reg.now = func() time.Time { return f.now }

	return f
}

func (f *registrationFixture) count(t *testing.T) int {
	t.Helper()
	n := 0
	for _, err := range f.repo.All() {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}

	return n
}

func TestRegistration_RequestFindComplete(t *testing.T) {
	f := newRegistrationFixture(t)
	uc := f.reg.useCases()

	if err := uc.Request("session-1", " Erika@Example.com "); err != nil {
		t.Fatal(err)
	}

	if len(f.links) != 1 || f.links[0].to != "erika@example.com" {
		t.Fatalf("expected one link to the normalized address, got %v", f.links)
	}

	link := f.links[0]
	email, err := uc.Find(link.id, link.code)
	if err != nil || email != "erika@example.com" {
		t.Fatalf("find: %v %v", email, err)
	}

	_, err = uc.Complete(link.id, link.code, ShortRegistrationUser{
		Firstname: "Erika",
		Email:     "attacker@example.com", // must be ignored
		Verified:  false,                  // must be overridden
	})
	if err != nil {
		t.Fatal(err)
	}

	got := f.created[0]
	if got.Email != "erika@example.com" || !got.Verified || !got.SelfRegistered || got.NotifyUser {
		t.Fatalf("unexpected model: %+v", got)
	}

	if _, err := uc.Find(link.id, link.code); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("a used link must be invalid, got %v", err)
	}

	if f.count(t) != 0 {
		t.Fatal("the request must be removed")
	}
}

func TestRegistration_CodeIsNotStored(t *testing.T) {
	f := newRegistrationFixture(t)
	if err := f.reg.request("", "erika@example.com"); err != nil {
		t.Fatal(err)
	}

	optReq, err := f.repo.FindByID(f.links[0].id)
	if err != nil || optReq.IsNone() {
		t.Fatal("request not stored", err)
	}

	req := optReq.Unwrap()
	if string(req.CodeHash) == f.links[0].code || len(req.CodeHash) != 32 {
		t.Fatal("the code must only be stored as hash")
	}
}

func TestRegistration_RegisteredAddressGetsNotice(t *testing.T) {
	f := newRegistrationFixture(t)
	f.used["erika@example.com"] = true

	if err := f.reg.request("", "erika@example.com"); err != nil {
		t.Fatalf("a registered address must look like any other, got %v", err)
	}

	if len(f.links) != 0 || len(f.registered) != 1 || f.count(t) != 0 {
		t.Fatalf("expected only a notice: links=%v notices=%v", f.links, f.registered)
	}

	// the cooldown applies the same way
	var cooldown RegistrationCooldownError
	if err := f.reg.request("", "erika@example.com"); !errors.As(err, &cooldown) {
		t.Fatalf("expected cooldown, got %v", err)
	}
}

func TestRegistration_NewLinkReplacesOld(t *testing.T) {
	f := newRegistrationFixture(t)
	_ = f.reg.request("", "erika@example.com")
	f.now = f.now.Add(2 * time.Minute)
	if err := f.reg.request("", "erika@example.com"); err != nil {
		t.Fatal(err)
	}

	if _, err := f.reg.find(f.links[0].id, f.links[0].code); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("the old link must be invalid, got %v", err)
	}

	if _, err := f.reg.find(f.links[1].id, f.links[1].code); err != nil {
		t.Fatal(err)
	}

	if f.count(t) != 1 {
		t.Fatal("expected exactly one request")
	}
}

func TestRegistration_Expired(t *testing.T) {
	f := newRegistrationFixture(t)
	_ = f.reg.request("", "erika@example.com")
	f.now = f.now.Add(DefaultRegistrationLifeTime)

	if _, err := f.reg.find(f.links[0].id, f.links[0].code); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("expected expired, got %v", err)
	}

	if f.count(t) != 0 {
		t.Fatal("an expired request must be removed")
	}
}

func TestRegistration_WrongCodes(t *testing.T) {
	f := newRegistrationFixture(t)
	_ = f.reg.request("", "erika@example.com")
	link := f.links[0]

	for range 5 {
		if _, err := f.reg.find(link.id, "wrong"); !errors.Is(err, ErrRegistrationInvalid) {
			t.Fatalf("expected invalid, got %v", err)
		}
	}

	if _, err := f.reg.find(link.id, link.code); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("too many wrong codes must remove the request, got %v", err)
	}
}

func TestRegistration_RateLimit(t *testing.T) {
	f := newRegistrationFixture(t)

	for i := range 5 {
		if err := f.reg.request("", "erika@example.com"); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		f.now = f.now.Add(2 * time.Minute)
	}

	var cooldown RegistrationCooldownError
	if err := f.reg.request("", "erika@example.com"); !errors.As(err, &cooldown) {
		t.Fatalf("expected the hourly limit, got %v", err)
	}

	if err := f.reg.request("", "max@example.com"); err != nil {
		t.Fatalf("other addresses are not affected: %v", err)
	}

	f.now = f.now.Add(time.Hour)
	if err := f.reg.request("", "erika@example.com"); err != nil {
		t.Fatalf("the limit must end after an hour: %v", err)
	}
}

func TestRegistration_Settings(t *testing.T) {
	f := newRegistrationFixture(t)

	f.settings.SelfRegistration = false
	if err := f.reg.request("", "erika@example.com"); !errors.Is(err, ErrRegistrationDisabled) {
		t.Fatalf("expected disabled, got %v", err)
	}

	f.settings.SelfRegistration = true
	f.settings.AllowedDomains = []string{"@worldiety.de"}
	if err := f.reg.request("", "erika@example.com"); err == nil {
		t.Fatal("expected a not allowed domain to fail")
	}

	if err := f.reg.request("", "not a mail"); err == nil {
		t.Fatal("expected an invalid address to fail")
	}

	if err := f.reg.request("", "erika@worldiety.de"); err != nil {
		t.Fatal(err)
	}

	// the settings may change between the request and the completion
	f.settings.SelfRegistration = false
	if _, err := f.reg.complete(f.links[0].id, f.links[0].code, ShortRegistrationUser{}); !errors.Is(err, ErrRegistrationDisabled) {
		t.Fatalf("expected disabled, got %v", err)
	}
}

func TestRegistration_FailedCreateKeepsLink(t *testing.T) {
	f := newRegistrationFixture(t)
	_ = f.reg.request("", "erika@example.com")
	link := f.links[0]

	f.createErr = errors.New("weak password")
	if _, err := f.reg.complete(link.id, link.code, ShortRegistrationUser{}); err == nil {
		t.Fatal("expected the error of create")
	}

	// the visitor corrects the input and tries again with the same link
	f.createErr = nil
	if _, err := f.reg.complete(link.id, link.code, ShortRegistrationUser{}); err != nil {
		t.Fatal(err)
	}

	if _, err := f.reg.complete(link.id, link.code, ShortRegistrationUser{}); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("a used link must be invalid, got %v", err)
	}
}

func TestRegistration_ClientLimit(t *testing.T) {
	f := newRegistrationFixture(t)

	for i := range 10 {
		if err := f.reg.request("session-1", Email(fmt.Sprintf("visitor%d@example.com", i))); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}

	var cooldown RegistrationCooldownError
	if err := f.reg.request("session-1", "another@example.com"); !errors.As(err, &cooldown) {
		t.Fatalf("expected the client limit, got %v", err)
	}

	if err := f.reg.request("session-2", "another@example.com"); err != nil {
		t.Fatalf("other clients are not affected: %v", err)
	}
}

func TestRegistration_DeliveryErrorLooksTheSame(t *testing.T) {
	f := newRegistrationFixture(t)
	f.used["known@example.com"] = true
	f.deliverErr = errors.New("smtp down")

	errNew := f.reg.request("", "new@example.com")
	errKnown := f.reg.request("", "known@example.com")
	if !errors.Is(errNew, ErrRegistrationDelivery) || !errors.Is(errKnown, ErrRegistrationDelivery) || errNew.Error() != errKnown.Error() {
		t.Fatalf("a registered address must not be told apart: %v / %v", errNew, errKnown)
	}

	if f.count(t) != 0 {
		t.Fatal("a request without a sent link must be removed")
	}
}

func TestRegistration_RemovesExpiredAtStart(t *testing.T) {
	f := newRegistrationFixture(t)
	_ = f.reg.request("", "erika@example.com")

	f.now = f.now.Add(DefaultRegistrationLifeTime)
	f.reg.removeExpired()
	if f.count(t) != 0 {
		t.Fatal("expected the expired request to be removed")
	}
}

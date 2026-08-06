// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"sync"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/events"
)

type testSubject struct {
	id    ID
	perms []permission.ID
}

func (s testSubject) ID() ID {
	return s.id
}

func (s testSubject) Valid() bool {
	return s.id != ""
}

func (s testSubject) HasPermission(p permission.ID) bool {
	for _, id := range s.perms {
		if id == p {
			return true
		}
	}

	return false
}

func (s testSubject) Audit(p permission.ID) error {
	if !s.Valid() {
		return InvalidSubjectErr
	}

	if !s.HasPermission(p) {
		return PermissionDeniedErr
	}

	return nil
}

// syncBus is a minimal synchronous event bus, so that tests do not have to deal with the
// goroutine spawning of the default async bus.
type syncBus struct {
	mutex  sync.Mutex
	events []any
}

func (b *syncBus) Publish(evt any) {
	if evt == nil {
		return
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.events = append(b.events, evt)
}

func (b *syncBus) Subscribe(fn func(evt any), opts ...events.SubscriberOption) (close func()) {
	panic("not required for these tests")
}

func mailChangedEvents(b *syncBus) []EMailChanged {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var res []EMailChanged
	for _, evt := range b.events {
		if e, ok := evt.(EMailChanged); ok {
			res = append(res, e)
		}
	}

	return res
}

func newChangeOtherEmailFixture(t *testing.T, users ...User) (ChangeOtherEmail, *mem.Repository[User, ID], *syncBus) {
	t.Helper()

	repo := &mem.Repository[User, ID]{}
	for _, usr := range users {
		if err := repo.Save(usr); err != nil {
			t.Fatal(err)
		}
	}

	notifyRepo := data.NewNotifyRepository[User, ID](nil, repo)

	var mutex sync.Mutex
	bus := &syncBus{}

	return NewChangeOtherEmail(&mutex, bus, notifyRepo, NewMailIndex(notifyRepo)), repo, bus
}

func TestChangeOtherEmail_PermissionDenied(t *testing.T) {
	uc, repo, _ := newChangeOtherEmailFixture(t, User{ID: "1", Email: "old@example.com", EMailVerified: true})

	err := uc(testSubject{id: "42"}, "1", "new@example.com", false)
	if !errors.Is(err, PermissionDeniedErr) {
		t.Fatalf("want permission denied, got %v", err)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "old@example.com" {
		t.Fatalf("mail must not have been changed, got %s", usr.Email)
	}
}

func TestChangeOtherEmail_InvalidMail(t *testing.T) {
	uc, _, _ := newChangeOtherEmailFixture(t, User{ID: "1", Email: "old@example.com"})

	err := uc(testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail}}, "1", "not-a-mail", false)
	if !errors.Is(err, InvalidEMailErr) {
		t.Fatalf("want invalid mail, got %v", err)
	}
}

func TestChangeOtherEmail_AlreadyInUse(t *testing.T) {
	uc, repo, _ := newChangeOtherEmailFixture(t,
		User{ID: "1", Email: "old@example.com"},
		User{ID: "2", Email: "taken@example.com"},
	)

	// note the different casing, mail addresses are matched case-insensitive
	err := uc(testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail}}, "1", "TAKEN@example.com", false)
	if !errors.Is(err, EMailAlreadyInUseErr) {
		t.Fatalf("want already in use, got %v", err)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "old@example.com" {
		t.Fatalf("mail must not have been changed, got %s", usr.Email)
	}
}

func TestChangeOtherEmail_Success(t *testing.T) {
	uc, repo, bus := newChangeOtherEmailFixture(t, User{
		ID:                  "1",
		Email:               "old@example.com",
		EMailVerified:       true,
		VerificationCode:    Code{Value: "abcdef"},
		PasswordRequestCode: Code{Value: "123456"},
	})

	if err := uc(testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail}}, "1", "new@example.com", true); err != nil {
		t.Fatal(err)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "new@example.com" {
		t.Fatalf("want new@example.com, got %s", usr.Email)
	}

	if usr.EMailVerified {
		t.Fatal("new mail must not be verified")
	}

	if !usr.VerificationCode.IsZero() || !usr.PasswordRequestCode.IsZero() {
		t.Fatal("codes of the old mail address must have been invalidated")
	}

	evts := mailChangedEvents(bus)
	if len(evts) != 1 {
		t.Fatalf("want exactly 1 event, got %d", len(evts))
	}

	evt := evts[0]
	if evt.ID != "1" || evt.OldEMail != "old@example.com" || evt.NewEMail != "new@example.com" || !evt.NotifyUser {
		t.Fatalf("unexpected event: %+v", evt)
	}
}

func TestChangeOtherEmail_SameMailIsNoop(t *testing.T) {
	uc, repo, bus := newChangeOtherEmailFixture(t, User{ID: "1", Email: "old@example.com", EMailVerified: true})

	if err := uc(testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail}}, "1", "OLD@example.com", true); err != nil {
		t.Fatal(err)
	}

	usr, _ := repo.Load("1")
	if !usr.EMailVerified {
		t.Fatal("a noop must not invalidate the verification")
	}

	if evts := mailChangedEvents(bus); len(evts) != 0 {
		t.Fatalf("want no event, got %d", len(evts))
	}
}

// TestChangeOtherEmail_SSOUser ensures that an admin can fix the mail address of an SSO managed user by hand,
// which is required if the address has been changed within the identity provider.
func TestChangeOtherEmail_SSOUser(t *testing.T) {
	uc, repo, _ := newChangeOtherEmailFixture(t, User{ID: "1", Email: "old@example.com", NLSManagedUser: true})

	if err := uc(testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail}}, "1", "new@example.com", false); err != nil {
		t.Fatal(err)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "new@example.com" {
		t.Fatalf("want new@example.com, got %s", usr.Email)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

func newMergeFixture(t *testing.T, users ...User) (MergeSingleSignOnUser, *datamem.Repository[User, ID], *UserIndex, *syncBus) {
	t.Helper()

	repo := &datamem.Repository[User, ID]{}
	for _, usr := range users {
		if err := repo.Save(usr); err != nil {
			t.Fatal(err)
		}
	}

	notifyRepo := data.NewNotifyRepository[User, ID](nil, repo)
	idx := NewUserIndex(notifyRepo)
	bus := &syncBus{}

	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	loadGlobal := settings.LoadGlobal(func(subject permission.Auditable, typ reflect.Type) (settings.GlobalSettings, error) {
		return Settings{}, nil
	})

	var mutex sync.Mutex
	merge := NewMergeSingleSignOnUser(&mutex, bus, notifyRepo, idx, loadGlobal, nil, rdb)

	return merge, repo, idx, bus
}

func TestMergeSingleSignOnUser_CreatesUserWithNLSUserID(t *testing.T) {
	merge, repo, idx, _ := newMergeFixture(t)

	uid, err := merge(SingleSignOnUser{ID: "entra-1", Email: "a@example.com", Name: "Anna Admin"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	usr, _ := repo.Load(uid)
	if usr.NLSUserID != "entra-1" {
		t.Fatalf("want entra-1, got %s", usr.NLSUserID)
	}

	if !usr.NLSManagedUser || !usr.EMailVerified {
		t.Fatal("an sso user must be managed and verified")
	}

	assertLookupNLS(t, idx, "entra-1", uid)
}

// TestMergeSingleSignOnUser_BackfillsNLSUserID covers accounts which have been merged before the subject id
// existed: they are still found by mail and get the id stamped on.
func TestMergeSingleSignOnUser_BackfillsNLSUserID(t *testing.T) {
	merge, repo, idx, _ := newMergeFixture(t, User{ID: "1", Email: "a@example.com", NLSManagedUser: true})

	uid, err := merge(SingleSignOnUser{ID: "entra-1", Email: "a@example.com"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "1" {
		t.Fatalf("want the existing user, got %s", uid)
	}

	usr, _ := repo.Load("1")
	if usr.NLSUserID != "entra-1" {
		t.Fatalf("want entra-1, got %s", usr.NLSUserID)
	}

	assertLookupNLS(t, idx, "entra-1", "1")
}

// TestMergeSingleSignOnUser_FollowsMailChange is the actual feature: the mail has been changed within the
// identity provider, so the local account must follow without locking the user out.
func TestMergeSingleSignOnUser_FollowsMailChange(t *testing.T) {
	merge, repo, idx, bus := newMergeFixture(t, User{
		ID:             "1",
		Email:          "old@example.com",
		NLSManagedUser: true,
		NLSUserID:      "entra-1",
		EMailVerified:  true,
	})

	uid, err := merge(SingleSignOnUser{ID: "entra-1", Email: "new@example.com"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "1" {
		t.Fatalf("want the existing user, got %s", uid)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "new@example.com" {
		t.Fatalf("want new@example.com, got %s", usr.Email)
	}

	if !usr.EMailVerified {
		t.Fatal("the address is verified by the identity provider, otherwise the user would be locked out")
	}

	assertLookup(t, idx, "new@example.com", "1")
	assertLookup(t, idx, "old@example.com", "")

	evts := mailChangedEvents(bus)
	if len(evts) != 1 {
		t.Fatalf("want exactly 1 event, got %d", len(evts))
	}

	if evts[0].NotifyUser {
		t.Fatal("the user must not be asked to confirm an address which the identity provider verified")
	}
}

// TestMergeSingleSignOnUser_RefusesForeignIdentity makes sure that the mail address alone cannot be used to
// take over an account which already belongs to a different external identity.
func TestMergeSingleSignOnUser_RefusesForeignIdentity(t *testing.T) {
	merge, repo, _, _ := newMergeFixture(t, User{
		ID:             "1",
		Email:          "a@example.com",
		NLSManagedUser: true,
		NLSUserID:      "entra-1",
	})

	_, err := merge(SingleSignOnUser{ID: "entra-2", Email: "a@example.com"}, nil)
	if err == nil {
		t.Fatal("the login must be refused")
	}

	if !errors.Is(err, os.ErrPermission) {
		t.Fatalf("unexpected error: %v", err)
	}

	usr, _ := repo.Load("1")
	if usr.NLSUserID != "entra-1" {
		t.Fatalf("the known identity must not be overwritten, got %s", usr.NLSUserID)
	}
}

// TestMergeSingleSignOnUser_RefusesMailOfAnotherUser covers the case where the new address is already taken
// by an unrelated local account.
func TestMergeSingleSignOnUser_RefusesMailOfAnotherUser(t *testing.T) {
	merge, repo, _, _ := newMergeFixture(t,
		User{ID: "1", Email: "old@example.com", NLSManagedUser: true, NLSUserID: "entra-1", EMailVerified: true},
		User{ID: "2", Email: "taken@example.com"},
	)

	_, err := merge(SingleSignOnUser{ID: "entra-1", Email: "taken@example.com"}, nil)
	if err == nil {
		t.Fatal("the merge must fail")
	}

	if !strings.Contains(err.Error(), "cannot follow mail change") {
		t.Fatalf("unexpected error: %v", err)
	}

	usr, _ := repo.Load("1")
	if usr.Email != "old@example.com" {
		t.Fatalf("the mail must not have been changed, got %s", usr.Email)
	}
}

// TestMergeSingleSignOnUser_WithoutNLSUserID keeps the behaviour of older NLS versions which do not report a
// subject id: pure mail based matching and no stamping.
func TestMergeSingleSignOnUser_WithoutNLSUserID(t *testing.T) {
	merge, repo, _, _ := newMergeFixture(t, User{ID: "1", Email: "a@example.com", NLSUserID: "entra-1"})

	uid, err := merge(SingleSignOnUser{Email: "a@example.com"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "1" {
		t.Fatalf("want the existing user, got %s", uid)
	}

	usr, _ := repo.Load("1")
	if usr.NLSUserID != "entra-1" {
		t.Fatalf("an existing identity must never be cleared, got %s", usr.NLSUserID)
	}
}

func TestMergeSingleSignOnUser_MatchesByIDDespiteOtherMail(t *testing.T) {
	merge, _, _, _ := newMergeFixture(t,
		User{ID: "1", Email: "old@example.com", NLSManagedUser: true, NLSUserID: "entra-1", EMailVerified: true},
	)

	uid, err := merge(SingleSignOnUser{ID: "entra-1", Email: "OLD@Example.com"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "1" {
		t.Fatalf("want the existing user, got %s", uid)
	}
}

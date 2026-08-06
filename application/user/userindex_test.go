// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"sync"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/mem"
)

func newUserIndexFixture(t *testing.T, users ...User) (data.NotifyRepository[User, ID], *UserIndex) {
	t.Helper()

	repo := &mem.Repository[User, ID]{}
	for _, usr := range users {
		if err := repo.Save(usr); err != nil {
			t.Fatal(err)
		}
	}

	notifyRepo := data.NewNotifyRepository[User, ID](nil, repo)

	return notifyRepo, NewUserIndex(notifyRepo)
}

func assertLookup(t *testing.T, idx *UserIndex, mail Email, want ID) {
	t.Helper()

	id, ok, err := idx.LookupMail(mail)
	if err != nil {
		t.Fatal(err)
	}

	if want == "" {
		if ok {
			t.Fatalf("%s should not be indexed, got %s", mail, id)
		}

		return
	}

	if !ok {
		t.Fatalf("%s should be indexed", mail)
	}

	if id != want {
		t.Fatalf("want %s for %s, got %s", want, mail, id)
	}
}

func TestUserIndex_InitialScan(t *testing.T) {
	_, idx := newUserIndexFixture(t,
		User{ID: "1", Email: "a@example.com"},
		User{ID: "2", Email: "b@example.com"},
	)

	assertLookup(t, idx, "a@example.com", "1")
	assertLookup(t, idx, "b@example.com", "2")
	assertLookup(t, idx, "c@example.com", "")
}

func TestUserIndex_CaseInsensitive(t *testing.T) {
	// note the mixed case within the repository, which the old FindByMail implementation did not find
	_, idx := newUserIndexFixture(t, User{ID: "1", Email: "MiXeD@Example.com"})

	assertLookup(t, idx, "mixed@example.com", "1")
	assertLookup(t, idx, "MIXED@EXAMPLE.COM", "1")
	assertLookup(t, idx, " mixed@example.com ", "1")
}

func TestUserIndex_SaveUpdatesIndex(t *testing.T) {
	repo, idx := newUserIndexFixture(t, User{ID: "1", Email: "old@example.com"})

	// force the initial scan, so that we test the observer and not the scan
	assertLookup(t, idx, "old@example.com", "1")

	if err := repo.Save(User{ID: "1", Email: "new@example.com"}); err != nil {
		t.Fatal(err)
	}

	assertLookup(t, idx, "new@example.com", "1")
	assertLookup(t, idx, "old@example.com", "")
}

func TestUserIndex_SaveBeforeInitialScan(t *testing.T) {
	repo, idx := newUserIndexFixture(t, User{ID: "1", Email: "old@example.com"})

	if err := repo.Save(User{ID: "2", Email: "other@example.com"}); err != nil {
		t.Fatal(err)
	}

	assertLookup(t, idx, "old@example.com", "1")
	assertLookup(t, idx, "other@example.com", "2")
}

func TestUserIndex_DeleteUpdatesIndex(t *testing.T) {
	repo, idx := newUserIndexFixture(t, User{ID: "1", Email: "a@example.com"})

	assertLookup(t, idx, "a@example.com", "1")

	if err := repo.DeleteByID("1"); err != nil {
		t.Fatal(err)
	}

	assertLookup(t, idx, "a@example.com", "")
}

func assertLookupNLS(t *testing.T, idx *UserIndex, nlsUserID NLSUserID, want ID) {
	t.Helper()

	id, ok, err := idx.LookupNLSUserID(nlsUserID)
	if err != nil {
		t.Fatal(err)
	}

	if want == "" {
		if ok {
			t.Fatalf("%s should not be indexed, got %s", nlsUserID, id)
		}

		return
	}

	if !ok {
		t.Fatalf("%s should be indexed", nlsUserID)
	}

	if id != want {
		t.Fatalf("want %s for %s, got %s", want, nlsUserID, id)
	}
}

func TestUserIndex_NLSUserID(t *testing.T) {
	repo, idx := newUserIndexFixture(t,
		User{ID: "1", Email: "a@example.com", NLSUserID: "entra-1"},
		User{ID: "2", Email: "b@example.com"},
	)

	assertLookupNLS(t, idx, "entra-1", "1")
	assertLookupNLS(t, idx, "entra-2", "")
	// an empty id must never match a user without an external identity
	assertLookupNLS(t, idx, "", "")

	if err := repo.Save(User{ID: "2", Email: "b@example.com", NLSUserID: "entra-2"}); err != nil {
		t.Fatal(err)
	}

	assertLookupNLS(t, idx, "entra-2", "2")
}

// TestUserIndex_NLSUserIDSurvivesMailChange is the core of the SSO matching: the mail address may change,
// the external identity must stay resolvable.
func TestUserIndex_NLSUserIDSurvivesMailChange(t *testing.T) {
	repo, idx := newUserIndexFixture(t, User{ID: "1", Email: "old@example.com", NLSUserID: "entra-1"})

	assertLookupNLS(t, idx, "entra-1", "1")

	if err := repo.Save(User{ID: "1", Email: "new@example.com", NLSUserID: "entra-1"}); err != nil {
		t.Fatal(err)
	}

	assertLookupNLS(t, idx, "entra-1", "1")
	assertLookup(t, idx, "new@example.com", "1")
	assertLookup(t, idx, "old@example.com", "")
}

func TestUserIndex_DeleteDropsNLSUserID(t *testing.T) {
	repo, idx := newUserIndexFixture(t, User{ID: "1", Email: "a@example.com", NLSUserID: "entra-1"})

	assertLookupNLS(t, idx, "entra-1", "1")

	if err := repo.DeleteByID("1"); err != nil {
		t.Fatal(err)
	}

	assertLookupNLS(t, idx, "entra-1", "")
	assertLookup(t, idx, "a@example.com", "")
}

func TestUserIndex_UsedByOther(t *testing.T) {
	_, idx := newUserIndexFixture(t,
		User{ID: "1", Email: "a@example.com"},
		User{ID: "2", Email: "b@example.com"},
	)

	assertUsedByOther(t, idx, "b@example.com", "1", true)
	// the own address must not count as a conflict
	assertUsedByOther(t, idx, "A@example.com", "1", false)
	assertUsedByOther(t, idx, "free@example.com", "1", false)
}

func assertUsedByOther(t *testing.T, idx *UserIndex, mail Email, except ID, want bool) {
	t.Helper()

	got, err := idx.UsedByOther(mail, except)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("UsedByOther(%s, %s): want %v, got %v", mail, except, want, got)
	}
}

// TestMailIndex_ChangeOtherEmailKeepsIndexInSync is the regression for the interaction of the use case and the
// index: after the change the user must be resolvable by the new address only.
func TestUserIndex_ChangeOtherEmailKeepsIndexInSync(t *testing.T) {
	repo := &mem.Repository[User, ID]{}
	if err := repo.Save(User{ID: "1", Email: "old@example.com"}); err != nil {
		t.Fatal(err)
	}

	notifyRepo := data.NewNotifyRepository[User, ID](nil, repo)
	idx := NewUserIndex(notifyRepo)

	var mutex sync.Mutex
	changeMail := NewChangeOtherEmail(&mutex, &syncBus{}, notifyRepo, idx)
	findByMail := NewFindByMail(notifyRepo, idx)
	mailUsed := NewEMailUsed(idx)

	admin := testSubject{id: "42", perms: []permission.ID{PermChangeOtherEmail, PermFindByMail}}
	if err := changeMail(admin, "1", "new@example.com", false); err != nil {
		t.Fatal(err)
	}

	optUsr, err := findByMail(admin, "new@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if optUsr.IsNone() || optUsr.Unwrap().ID != "1" {
		t.Fatal("user must be findable by the new mail address")
	}

	optUsr, err = findByMail(admin, "old@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if optUsr.IsSome() {
		t.Fatal("user must not be findable by the old mail address anymore")
	}

	used, err := mailUsed("old@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if used {
		t.Fatal("old mail address must be free again")
	}
}

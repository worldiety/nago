// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"testing"
	"time"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/std"
)

// countingRepository counts the lookups, which actually reach the repository.
type countingRepository struct {
	data.NotifyRepository[User, ID]
	finds int
}

func (r *countingRepository) FindByID(id ID) (std.Option[User], error) {
	r.finds++
	return r.NotifyRepository.FindByID(id)
}

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func newDisplayNameFixture(t *testing.T, users ...User) (*countingRepository, *fakeClock, DisplayName) {
	t.Helper()

	repo := &mem.Repository[User, ID]{}
	for _, usr := range users {
		if err := repo.Save(usr); err != nil {
			t.Fatal(err)
		}
	}

	counting := &countingRepository{NotifyRepository: data.NewNotifyRepository[User, ID](nil, repo)}
	clock := &fakeClock{now: time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)}

	return counting, clock, newDisplayName(counting, 5*time.Minute, clock.Now)
}

func TestDisplayName_CachesWithinRefreshInterval(t *testing.T) {
	repo, clock, displayName := newDisplayNameFixture(t, User{ID: "1", Contact: Contact{Firstname: "Erika", Lastname: "Mustermann"}, Status: Enabled{}})

	for range 3 {
		if got := displayName("1").Displayname; got != "Erika Mustermann" {
			t.Fatalf("unexpected display name %q", got)
		}
	}

	if repo.finds != 1 {
		t.Fatalf("expected a single repository lookup, got %d", repo.finds)
	}

	clock.now = clock.now.Add(4 * time.Minute)
	displayName("1")
	if repo.finds != 1 {
		t.Fatalf("expected the cache to be used within the refresh interval, got %d lookups", repo.finds)
	}

	clock.now = clock.now.Add(2 * time.Minute)
	displayName("1")
	if repo.finds != 2 {
		t.Fatalf("expected a new lookup after the refresh interval, got %d lookups", repo.finds)
	}

	displayName("1")
	if repo.finds != 2 {
		t.Fatalf("expected the refreshed cache to be used again, got %d lookups", repo.finds)
	}
}

func TestDisplayName_SaveEvictsAtOnce(t *testing.T) {
	repo, _, displayName := newDisplayNameFixture(t,
		User{ID: "1", Contact: Contact{Firstname: "Erika"}, Status: Enabled{}},
		User{ID: "2", Contact: Contact{Firstname: "Max"}, Status: Enabled{}},
	)

	displayName("1")
	displayName("2")

	if err := repo.Save(User{ID: "1", Contact: Contact{Firstname: "Erika", Lastname: "Musterfrau"}, Status: Disabled{}}); err != nil {
		t.Fatal(err)
	}

	got := displayName("1")
	if got.Displayname != "Erika Musterfrau" {
		t.Fatalf("expected the changed name at once, got %q", got.Displayname)
	}

	if got.Valid {
		t.Fatal("expected the disabled account at once")
	}

	// the other user is still cached
	displayName("2")
	if repo.finds != 3 {
		t.Fatalf("expected only the saved user to be evicted, got %d lookups", repo.finds)
	}
}

func TestDisplayName_DeleteEvictsAtOnce(t *testing.T) {
	repo, _, displayName := newDisplayNameFixture(t, User{ID: "1", Contact: Contact{Firstname: "Erika"}, Status: Enabled{}})

	if !displayName("1").Valid {
		t.Fatal("expected a valid user")
	}

	if err := repo.DeleteByID("1"); err != nil {
		t.Fatal(err)
	}

	if got := displayName("1"); got.ID != "" {
		t.Fatalf("expected the zero value for a deleted user, got %+v", got)
	}
}

func TestDisplayName_EnabledUntilExpires(t *testing.T) {
	repo, _, displayName := newDisplayNameFixture(t, User{ID: "1", Status: EnabledUntil{ValidUntil: time.Now().Add(50 * time.Millisecond)}})

	if !displayName("1").Valid {
		t.Fatal("expected a valid user before the limit")
	}

	displayName("1")
	if repo.finds != 1 {
		t.Fatalf("expected the cache before the limit, got %d lookups", repo.finds)
	}

	time.Sleep(60 * time.Millisecond)

	if displayName("1").Valid {
		t.Fatal("expected an invalid user after the limit")
	}
}

func TestDisplayName_PlainRepositoryStillWorks(t *testing.T) {
	repo := &mem.Repository[User, ID]{}
	if err := repo.Save(User{ID: "1", Contact: Contact{Firstname: "Erika"}}); err != nil {
		t.Fatal(err)
	}

	displayName := NewDisplayName(repo, 5*time.Minute)
	if got := displayName("1").Displayname; got != "Erika" {
		t.Fatalf("unexpected display name %q", got)
	}
}

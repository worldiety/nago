// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role_test

import (
	"context"
	"io"
	"iter"
	"slices"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	. "go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
)

// countingStore wraps a blob store and counts the mutating operations. It exists because the central promise
// of UpsertPermissions - that an already satisfied role costs no write - is invisible from the outside: the
// resulting permission set is identical either way, so only the number of writes can tell the two apart.
type countingStore struct {
	blob.Store
	writes  int
	deletes int
}

func (s *countingStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	s.writes++
	return s.Store.NewWriter(ctx, key)
}

func (s *countingStore) Delete(ctx context.Context, key string) error {
	s.deletes++
	return s.Store.Delete(ctx, key)
}

func (s *countingStore) NewReader(ctx context.Context, key string) (option.Opt[io.ReadCloser], error) {
	return s.Store.NewReader(ctx, key)
}

func (s *countingStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return s.Store.List(ctx, opts)
}

// newCountingUseCases is newTestUseCases with an observable ReBAC store.
func newCountingUseCases(t *testing.T) (UseCases, *countingStore) {
	t.Helper()

	store := &countingStore{Store: mem.NewBlobStore("rebac")}

	rdb, err := rebac.NewDB(store)
	if err != nil {
		t.Fatalf("cannot create rebac db: %v", err)
	}

	for perm := range permission.All() {
		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   Namespace,
			Relation: rebac.Relation(perm.ID),
			Target:   rebac.Global,
		})
	}

	repo := Repository(json.NewSloppyJSONRepository[Role, ID](mem.NewBlobStore(string(Namespace))))

	return NewUseCases(repo, events.NewEventBus(), rdb), store
}

// listPermissions reads the permissions of a role through the use case.
func listPermissions(t *testing.T, uc UseCases, id ID) []permission.ID {
	t.Helper()

	var got []permission.ID
	for pid, err := range uc.ListPermissions(user.SU(), id) {
		if err != nil {
			t.Fatalf("cannot list permissions: %v", err)
		}
		got = append(got, pid)
	}

	slices.Sort(got)
	return got
}

// TestUpsertPermissionsGrantsMissing is the base case: an empty role receives exactly what was asked for.
func TestUpsertPermissionsGrantsMissing(t *testing.T) {
	uc, _ := newCountingUseCases(t)

	const rid ID = "test.role.upsert"
	want := []permission.ID{PermFindAll, PermFindByID}

	if err := uc.UpsertPermissions(user.SU(), rid, want); err != nil {
		t.Fatalf("cannot upsert permissions: %v", err)
	}

	got := listPermissions(t, uc, rid)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestUpsertPermissionsIsWriteFreeWhenSatisfied is the property that makes the use case safe to call on every
// application start.
func TestUpsertPermissionsIsWriteFreeWhenSatisfied(t *testing.T) {
	uc, store := newCountingUseCases(t)

	const rid ID = "test.role.idempotent"
	perms := []permission.ID{PermFindAll, PermFindByID}

	if err := uc.UpsertPermissions(user.SU(), rid, perms); err != nil {
		t.Fatalf("cannot upsert permissions: %v", err)
	}

	if store.writes == 0 {
		t.Fatal("precondition failed: the first upsert should have written something")
	}

	store.writes = 0
	store.deletes = 0

	// Repeat with the same set, and once more with the set in a different order and with a duplicate, since
	// neither may be mistaken for a change.
	if err := uc.UpsertPermissions(user.SU(), rid, perms); err != nil {
		t.Fatalf("cannot repeat upsert: %v", err)
	}

	if err := uc.UpsertPermissions(user.SU(), rid, []permission.ID{PermFindByID, PermFindAll, PermFindAll}); err != nil {
		t.Fatalf("cannot repeat upsert: %v", err)
	}

	if store.writes != 0 || store.deletes != 0 {
		t.Errorf("an already satisfied upsert performed %d writes and %d deletes, want none", store.writes, store.deletes)
	}
}

// TestUpsertPermissionsKeepsForeignGrants distinguishes the use case from UpdatePermissions: an operator may
// grant a system role more than the declaring module asked for, and a restart must not take it away.
func TestUpsertPermissionsKeepsForeignGrants(t *testing.T) {
	uc, _ := newCountingUseCases(t)

	const rid ID = "test.role.additive"

	if err := uc.UpdatePermissions(user.SU(), rid, []permission.ID{PermDelete}); err != nil {
		t.Fatalf("cannot set initial permissions: %v", err)
	}

	if err := uc.UpsertPermissions(user.SU(), rid, []permission.ID{PermFindAll}); err != nil {
		t.Fatalf("cannot upsert permissions: %v", err)
	}

	want := []permission.ID{PermDelete, PermFindAll}
	slices.Sort(want)

	if got := listPermissions(t, uc, rid); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v - upsert must never revoke", got, want)
	}
}

// deniedSubject is a minimal auditable that holds nothing at all.
type deniedSubject struct{}

func (deniedSubject) Audit(permission.ID) error        { return user.PermissionDeniedErr }
func (deniedSubject) HasPermission(permission.ID) bool { return false }

// TestUpsertPermissionsRequiresUpdatePermission keeps the audit in place; a caller that may not change a role
// must not be able to widen it through the additive path.
func TestUpsertPermissionsRequiresUpdatePermission(t *testing.T) {
	uc, store := newCountingUseCases(t)

	if err := uc.UpsertPermissions(deniedSubject{}, "test.role.denied", []permission.ID{PermFindAll}); err == nil {
		t.Error("a subject without nago.role.update was allowed to grant permissions")
	}

	if store.writes != 0 {
		t.Errorf("the denied upsert still performed %d writes", store.writes)
	}
}

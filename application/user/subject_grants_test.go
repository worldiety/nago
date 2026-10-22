// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

const (
	grantsPermByRole  permission.ID = "de.worldiety.subjectgrants.by_role"
	grantsPermByGroup permission.ID = "de.worldiety.subjectgrants.by_group"
	grantsPermMissing permission.ID = "de.worldiety.subjectgrants.missing"
)

// newGrantsTestSubject builds a real viewImpl for a user, who is member of one role and one group, wired with the same
// resolvers as the application. The role carries a realistic amount of other permissions.
func newGrantsTestSubject(tb testing.TB) *viewImpl {
	tb.Helper()

	repo := Repository(json.NewSloppyJSONRepository[User, ID](mem.NewBlobStore("user")))

	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		tb.Fatalf("cannot create rebac db: %v", err)
	}

	rdb.ValidateStaticRules(false)
	rdb.AddMemberResolver(Namespace, role.Namespace)
	rdb.AddMemberResolver(Namespace, group.Namespace)

	usr := User{ID: "42", EMailVerified: true, Status: Enabled{}}
	if err := repo.Save(usr); err != nil {
		tb.Fatalf("cannot save user: %v", err)
	}

	global := rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances}
	member := func(ns rebac.Namespace, id rebac.Instance) rebac.Triple {
		return rebac.Triple{
			Source:   rebac.Entity{Namespace: ns, Instance: id},
			Relation: rebac.Member,
			Target:   rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(usr.ID)},
		}
	}

	triples := []rebac.Triple{
		member(role.Namespace, "editor"),
		member(group.Namespace, "team"),
		{Source: rebac.Entity{Namespace: role.Namespace, Instance: "editor"}, Relation: rebac.Relation(grantsPermByRole), Target: global},
		{Source: rebac.Entity{Namespace: group.Namespace, Instance: "team"}, Relation: rebac.Relation(grantsPermByGroup), Target: global},
	}

	for i := range 100 {
		triples = append(triples, rebac.Triple{
			Source:   rebac.Entity{Namespace: role.Namespace, Instance: "editor"},
			Relation: rebac.Relation(fmt.Sprintf("de.worldiety.subjectgrants.other%03d", i)),
			Target:   global,
		})
	}

	// the missing permission is known to the database, otherwise the lookup would already fail at interning
	triples = append(triples, rebac.Triple{
		Source:   rebac.Entity{Namespace: role.Namespace, Instance: "admin"},
		Relation: rebac.Relation(grantsPermMissing),
		Target:   global,
	})

	for _, triple := range triples {
		if err := rdb.Put(triple); err != nil {
			tb.Fatalf("cannot put triple: %v", err)
		}
	}

	return newViewImpl(context.Background, rdb, repo, usr)
}

func TestGlobalGrantsMatchesHasPermission(t *testing.T) {
	v := newGrantsTestSubject(t)

	grants, ok := v.GlobalGrants()
	if !ok {
		t.Fatal("grants must be enumerable")
	}

	for perm, want := range map[permission.ID]bool{
		grantsPermByRole:  true,
		grantsPermByGroup: true,
		grantsPermMissing: false,
	} {
		if got := v.HasPermission(perm); got != want {
			t.Errorf("%s: got %v, want %v", perm, got, want)
		}

		if got := grants.Has(rebac.Relation(perm)); got != want {
			t.Errorf("%s: grants got %v, want %v", perm, got, want)
		}
	}

	// an invalid user has no permissions at all, like HasPermission
	v.user.Status = Disabled{}
	v.lastRefreshedAt = time.Now()
	grants, ok = v.GlobalGrants()
	if !ok || grants.Has(rebac.Relation(grantsPermByRole)) {
		t.Error("a disabled user must not have any grants")
	}
}

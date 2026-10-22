// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"fmt"
	"testing"

	"go.wdy.de/nago/pkg/blob/mem"
)

// TestGlobalGrantsMatchesResolve ensures that the snapshot decides exactly like resolving each relation individually.
func TestGlobalGrantsMatchesResolve(t *testing.T) {
	const (
		usr   Namespace = "nago.iam.user"
		role  Namespace = "nago.iam.role"
		group Namespace = "nago.iam.group"
		other Namespace = "nago.other"
		doc   Namespace = "nago.doc"
	)

	db, err := NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	db.ValidateStaticRules(false)
	db.AddMemberResolver(usr, role)
	db.AddMemberResolver(usr, group)

	global := Entity{Global, AllInstances}
	put := func(src Entity, rel Relation, tgt Entity) {
		t.Helper()
		if err := db.Put(Triple{Source: src, Relation: rel, Target: tgt}); err != nil {
			t.Fatal(err)
		}
	}

	var relations []Relation
	for i := range 50 {
		relations = append(relations, Relation(fmt.Sprintf("perm%02d", i)))
	}

	// intern half of the relations before and half after "member", so that the role has relations on both sides
	// of its membership block
	for _, rel := range relations[:25] {
		put(Entity{role, "r1"}, rel, global)
	}

	for i := range 300 {
		put(Entity{role, "r1"}, Member, Entity{usr, Instance(fmt.Sprintf("crowd%03d", i))})
	}

	put(Entity{role, "r1"}, Member, Entity{usr, "u1"})
	for _, rel := range relations[25:40] {
		put(Entity{role, "r1"}, rel, global)
	}

	put(Entity{role, "r2"}, Member, Entity{usr, "u2"})
	put(Entity{role, "r2"}, relations[40], global)
	put(Entity{group, "g1"}, Member, Entity{usr, "u2"})
	put(Entity{group, "g1"}, relations[41], global)
	// a grant on another target is not global
	put(Entity{group, "g1"}, relations[42], Entity{doc, "d1"})
	put(Entity{group, "g1"}, relations[43], Entity{Global, "something"})
	// a membership in a namespace without a member resolver does not grant anything
	put(Entity{other, "o1"}, Member, Entity{usr, "u2"})
	put(Entity{other, "o1"}, relations[44], global)
	// direct grants and resource grants of a user
	put(Entity{usr, "u3"}, relations[45], global)
	put(Entity{usr, "u3"}, relations[46], Entity{doc, "d1"})
	// a nonsense membership of a user, see the inverted AddUserToGroup bug
	put(Entity{usr, "u3"}, Member, Entity{group, "g1"})

	candidates := append(relations, Member, "never-interned")
	users := []Instance{"u1", "u2", "u3", "crowd007", "unknown"}

	granted := 0
	for _, uid := range users {
		grants, ok := db.GlobalGrants(Entity{usr, uid})
		if !ok {
			t.Fatalf("%s: grants must be enumerable", uid)
		}

		for _, rel := range candidates {
			want, err := db.Resolve(Triple{Source: Entity{usr, uid}, Relation: rel, Target: global})
			if err != nil {
				t.Fatal(err)
			}

			if got := grants.Has(rel); got != want {
				t.Errorf("%s %s: got %v, want %v", uid, rel, got, want)
			}

			if want {
				granted++
			}
		}
	}

	// u1 and crowd007 through r1, u2 through r2 and g1, u3 directly
	if want := 40 + 40 + 2 + 1; granted != want {
		t.Errorf("granted %d relations, want %d", granted, want)
	}

	if _, ok := db.GlobalGrants(Entity{usr, ""}); ok {
		t.Error("an incomplete source must not be enumerable")
	}

	db.AddResolver(func(db *DB, triple Triple) (bool, error) {
		return false, nil
	})

	if _, ok := db.GlobalGrants(Entity{usr, "u1"}); ok {
		t.Error("an opaque resolver must not be enumerable")
	}
}

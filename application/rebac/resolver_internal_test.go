// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"testing"

	"go.wdy.de/nago/pkg/blob/mem"
)

// TestContainsViaMemberMatchesQuery ensures that the binary fast path of the member resolver decides exactly like
// the generic implementation based on Query and Contains.
func TestContainsViaMemberMatchesQuery(t *testing.T) {
	db, err := NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	db.ValidateStaticRules(false)

	const (
		usr   Namespace = "nago.iam.user"
		role  Namespace = "nago.iam.role"
		group Namespace = "nago.iam.group"
		other Namespace = "nago.other"
		doc   Namespace = "nago.doc"
	)

	member := func(ns Namespace, id Instance, uid Instance) Triple {
		return Triple{Source: Entity{ns, id}, Relation: Member, Target: Entity{usr, uid}}
	}

	grant := func(ns Namespace, id Instance, rel Relation, target Entity) Triple {
		return Triple{Source: Entity{ns, id}, Relation: rel, Target: target}
	}

	for _, triple := range []Triple{
		member(role, "r1", "u1"),
		member(role, "r2", "u1"),
		member(role, "r2", "u2"),
		member(group, "g1", "u2"),
		member(other, "o1", "u1"),
		grant(role, "r1", "read", Entity{Global, AllInstances}),
		grant(role, "r2", "write", Entity{doc, "d1"}),
		grant(role, "r2", "delete", Entity{doc, AllInstances}),
		grant(group, "g1", "read", Entity{doc, "d2"}),
		grant(other, "o1", "admin", Entity{Global, AllInstances}),
		// a grant of a role, which nobody is member of
		grant(role, "r3", "admin", Entity{Global, AllInstances}),
		// a direct grant is not the business of the member resolver
		grant(usr, "u3", "read", Entity{Global, AllInstances}),
	} {
		if err := db.Put(triple); err != nil {
			t.Fatal(err)
		}
	}

	users := []Instance{"u1", "u2", "u3", "unknown"}
	relations := []Relation{"read", "write", "delete", "admin", "unknown"}
	targets := []Entity{{Global, AllInstances}, {doc, "d1"}, {doc, "d2"}, {doc, "never-seen"}, {doc, AllInstances}, {"nago.unknown", "x"}}
	memberNamespaces := []Namespace{role, group, other, "nago.unknown"}

	granted := 0
	for _, uid := range users {
		for _, rel := range relations {
			for _, target := range targets {
				for _, ns := range memberNamespaces {
					triple := Triple{Source: Entity{usr, uid}, Relation: rel, Target: target}

					want, err := containsViaMemberQuery(db, ns, triple)
					if err != nil {
						t.Fatal(err)
					}

					got, applicable := db.containsViaMember(ns, triple)
					if !applicable {
						t.Fatalf("%v via %s: fast path must be applicable", triple, ns)
					}

					if got != want {
						t.Errorf("%v via %s: got %v, want %v", triple, ns, got, want)
					}

					if got {
						granted++
					}
				}
			}
		}
	}

	// guard against a vacuous comparison
	if granted == 0 {
		t.Fatal("no combination has been granted at all")
	}

	// the empty string means "anything" in a query, which only the generic implementation supports
	if _, applicable := db.containsViaMember(role, Triple{Source: Entity{usr, ""}, Relation: "read", Target: Entity{Global, AllInstances}}); applicable {
		t.Error("fast path must not be applicable for an empty source instance")
	}
}

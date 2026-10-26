// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac_test

import (
	"fmt"
	"slices"
	"testing"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/blob/mem"
)

const (
	tUser   rebac.Namespace = "nago.iam.user"
	tRole   rebac.Namespace = "nago.iam.role"
	tGroup  rebac.Namespace = "nago.iam.group"
	tOther  rebac.Namespace = "nago.other"
	tDoc    rebac.Namespace = "nago.doc"
	tFolder rebac.Namespace = "nago.folder"
)

var tGlobal = rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances}

func triple(src rebac.Entity, rel rebac.Relation, tgt rebac.Entity) rebac.Triple {
	return rebac.Triple{Source: src, Relation: rel, Target: tgt}
}

func entity(ns rebac.Namespace, inst rebac.Instance) rebac.Entity {
	return rebac.Entity{Namespace: ns, Instance: inst}
}

// newRuleTestDB creates a database with direct grants, wildcards, role and group memberships, a membership chain,
// a cyclic membership and a parent hierarchy with a custom membership relation.
func newRuleTestDB(tb testing.TB) *rebac.DB {
	tb.Helper()

	db, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		tb.Fatal(err)
	}

	db.ValidateStaticRules(false)
	db.AddRule(rebac.Rule{Source: tUser, Via: tRole})
	db.AddRule(rebac.Rule{Source: tUser, Via: tGroup})
	// a group may be member of a role, thus the chain user -> group -> role must be followed
	db.AddRule(rebac.Rule{Source: tGroup, Via: tRole})
	// roles may include roles, which also allows cycles
	db.AddRule(rebac.Rule{Source: tRole, Via: tRole})
	// a document inherits the relations of its parent folder, with another membership relation than Member
	db.AddRule(rebac.Rule{Source: tDoc, Via: tFolder, Membership: rebac.Parent})

	for _, t := range []rebac.Triple{
		// memberships
		triple(entity(tRole, "r1"), rebac.Member, entity(tUser, "u1")),
		triple(entity(tRole, "r2"), rebac.Member, entity(tUser, "u1")),
		triple(entity(tRole, "r2"), rebac.Member, entity(tUser, "u2")),
		triple(entity(tGroup, "g1"), rebac.Member, entity(tUser, "u2")),
		triple(entity(tGroup, "g2"), rebac.Member, entity(tUser, "u4")),
		triple(entity(tOther, "o1"), rebac.Member, entity(tUser, "u1")),
		// chain: u4 -> g2 -> r3
		triple(entity(tRole, "r3"), rebac.Member, entity(tGroup, "g2")),
		// cycle: rc1 <-> rc2, u5 is member of rc1
		triple(entity(tRole, "rc1"), rebac.Member, entity(tRole, "rc2")),
		triple(entity(tRole, "rc2"), rebac.Member, entity(tRole, "rc1")),
		triple(entity(tRole, "rc1"), rebac.Member, entity(tUser, "u5")),
		// a nonsense membership of a user, see the inverted AddUserToGroup bug
		triple(entity(tUser, "u3"), rebac.Member, entity(tGroup, "g1")),
		// hierarchy: folder f1 is the parent of d1
		triple(entity(tFolder, "f1"), rebac.Parent, entity(tDoc, "d1")),

		// grants
		triple(entity(tRole, "r1"), "read", tGlobal),
		triple(entity(tRole, "r2"), "write", entity(tDoc, "d1")),
		triple(entity(tRole, "r2"), "delete", entity(tDoc, rebac.AllInstances)),
		triple(entity(tRole, "r3"), "chained", tGlobal),
		triple(entity(tRole, "rc2"), "cyclic", entity(tDoc, "d2")),
		triple(entity(tGroup, "g1"), "read", entity(tDoc, "d2")),
		triple(entity(tOther, "o1"), "admin", tGlobal),
		// a grant of a role, which nobody is member of
		triple(entity(tRole, "r9"), "admin", tGlobal),
		// direct grants, also on a wildcard
		triple(entity(tUser, "u3"), "read", tGlobal),
		triple(entity(tUser, "u3"), "write", entity(tDoc, rebac.AllInstances)),
		triple(entity(tUser, "u3"), "share", entity(tDoc, "d1")),
		// a folder grant, which d1 inherits
		triple(entity(tFolder, "f1"), "inherited", entity(tUser, "u1")),
	} {
		if err := db.Put(t); err != nil {
			tb.Fatal(err)
		}
	}

	return db
}

// reference decides like [rebac.DB.Resolve], but is written naively with Query and Contains on the string level,
// so that the binary engine is checked against an independent and readable implementation.
func reference(db *rebac.DB, t rebac.Triple, depth int) bool {
	if t.Source.Namespace == "" || t.Source.Instance == "" || t.Relation == "" || t.Target.Namespace == "" || t.Target.Instance == "" {
		return false
	}

	if ok, _ := db.Contains(t); ok {
		return true
	}

	if t.Target.Instance != rebac.AllInstances {
		w := t
		w.Target.Instance = rebac.AllInstances
		if ok, _ := db.Contains(w); ok {
			return true
		}
	}

	if depth >= rebac.MaxRuleDepth {
		return false
	}

	for rule := range db.Rules() {
		if rule.Source != t.Source.Namespace {
			continue
		}

		q := rebac.Select().
			Where().Source().IsNamespace(rule.Via).
			Where().Relation().Has(rule.Membership).
			Where().Target().Set(t.Source)

		for membership, err := range db.Query(q) {
			if err != nil {
				panic(err)
			}

			sub := t
			sub.Source = membership.Source
			if reference(db, sub, depth+1) {
				return true
			}
		}
	}

	return false
}

type ruleCase struct {
	src  rebac.Entity
	rel  rebac.Relation
	tgt  rebac.Entity
	want bool
}

var ruleCases = []ruleCase{
	{entity(tUser, "u1"), "read", tGlobal, true},                      // by role r1
	{entity(tUser, "u1"), "write", entity(tDoc, "d1"), true},          // by role r2
	{entity(tUser, "u1"), "write", entity(tDoc, "d2"), false},         // r2 writes only d1
	{entity(tUser, "u1"), "delete", entity(tDoc, "d1"), true},         // by role r2 wildcard
	{entity(tUser, "u1"), "delete", entity(tDoc, "never-seen"), true}, // wildcard covers unknown instances
	{entity(tUser, "u1"), "admin", tGlobal, false},                    // o1 has no rule
	{entity(tUser, "u2"), "read", entity(tDoc, "d2"), true},           // by group g1
	{entity(tUser, "u2"), "read", tGlobal, false},                     // g1 reads only d2
	{entity(tUser, "u3"), "read", tGlobal, true},                      // direct
	{entity(tUser, "u3"), "write", entity(tDoc, "d1"), true},          // direct wildcard
	{entity(tUser, "u3"), "share", entity(tDoc, "d2"), false},         // direct, other instance
	{entity(tUser, "u3"), "read", entity(tDoc, "d2"), false},          // the nonsense membership grants nothing
	{entity(tUser, "u4"), "chained", tGlobal, true},                   // u4 -> g2 -> r3
	{entity(tUser, "u5"), "cyclic", entity(tDoc, "d2"), true},         // u5 -> rc1 -> rc2, terminates
	{entity(tUser, "u5"), "unknown", entity(tDoc, "d2"), false},       // the cycle terminates without a match
	{entity(tDoc, "d1"), "inherited", entity(tUser, "u1"), true},      // by parent folder
	{entity(tDoc, "d2"), "inherited", entity(tUser, "u1"), false},     // d2 has no parent
	{entity(tRole, "r1"), rebac.Member, entity(tUser, "u1"), true},    // stored memberships resolve as well
	{entity(tUser, "unknown"), "read", tGlobal, false},                // never interned
	{entity(tUser, "u1"), "", tGlobal, false},                         // incomplete
	{entity(tUser, ""), "read", tGlobal, false},                       // incomplete
}

func TestResolveRules(t *testing.T) {
	db := newRuleTestDB(t)

	for _, c := range ruleCases {
		got, err := db.Resolve(triple(c.src, c.rel, c.tgt))
		if err != nil {
			t.Fatal(err)
		}

		if got != c.want {
			t.Errorf("%s %s %s: got %v, want %v", c.src, c.rel, c.tgt, got, c.want)
		}
	}

	// Contains stays exact: neither the wildcard nor a rule applies
	for _, tr := range []rebac.Triple{
		triple(entity(tUser, "u3"), "write", entity(tDoc, "d1")),
		triple(entity(tUser, "u1"), "read", tGlobal),
	} {
		if ok, _ := db.Contains(tr); ok {
			t.Errorf("%v: Contains must be exact", tr)
		}
	}
}

// TestResolveMatchesReference compares the binary engine with the naive reference implementation for every
// combination of the known sources, relations and targets.
func TestResolveMatchesReference(t *testing.T) {
	db := newRuleTestDB(t)

	var sources []rebac.Entity
	for _, ns := range []rebac.Namespace{tUser, tRole, tGroup, tDoc, tOther, "nago.unknown"} {
		for _, inst := range []rebac.Instance{"u1", "u2", "u3", "u4", "u5", "r1", "r2", "rc1", "g1", "g2", "d1", "d2", "o1", "unknown", ""} {
			sources = append(sources, entity(ns, inst))
		}
	}

	relations := []rebac.Relation{"read", "write", "delete", "share", "admin", "chained", "cyclic", "inherited", rebac.Member, rebac.Parent, "unknown", ""}
	targets := []rebac.Entity{
		tGlobal, entity(tDoc, "d1"), entity(tDoc, "d2"), entity(tDoc, "never-seen"), entity(tDoc, rebac.AllInstances),
		entity(tUser, "u1"), entity(tGroup, "g1"), entity(tRole, "rc1"), entity(rebac.Global, "something"), entity("nago.unknown", "x"), entity(tDoc, ""),
	}

	granted := 0
	for _, src := range sources {
		for _, rel := range relations {
			for _, tgt := range targets {
				tr := triple(src, rel, tgt)
				want := reference(db, tr, 0)
				got, err := db.Resolve(tr)
				if err != nil {
					t.Fatal(err)
				}

				if got != want {
					t.Errorf("%v: got %v, want %v", tr, got, want)
				}

				if got {
					granted++
				}
			}
		}
	}

	// guard against a vacuous comparison
	if granted < 20 {
		t.Fatalf("only %d combinations have been granted", granted)
	}
}

// TestRelationsMatchesResolve ensures that the enumeration decides exactly like resolving each relation
// individually, for global and resource targets, through roles with many members and through chains.
func TestRelationsMatchesResolve(t *testing.T) {
	db := newRuleTestDB(t)

	// a role with relations on both sides of a large membership block, so that the block must be skipped
	var relations []rebac.Relation
	for i := range 50 {
		relations = append(relations, rebac.Relation(fmt.Sprintf("perm%02d", i)))
	}

	put := func(tr rebac.Triple) {
		t.Helper()
		if err := db.Put(tr); err != nil {
			t.Fatal(err)
		}
	}

	for _, rel := range relations[:25] {
		put(triple(entity(tRole, "big"), rel, tGlobal))
	}

	for i := range 300 {
		put(triple(entity(tRole, "big"), rebac.Member, entity(tUser, rebac.Instance(fmt.Sprintf("crowd%03d", i)))))
	}

	put(triple(entity(tRole, "big"), rebac.Member, entity(tUser, "u1")))
	for _, rel := range relations[25:40] {
		put(triple(entity(tRole, "big"), rel, tGlobal))
	}

	// grants on other targets of the same relation block
	put(triple(entity(tRole, "big"), relations[40], entity(tDoc, "d1")))
	put(triple(entity(tRole, "big"), relations[41], entity(tDoc, rebac.AllInstances)))
	put(triple(entity(tRole, "big"), relations[42], entity(rebac.Global, "something")))
	for i := range 20 {
		put(triple(entity(tRole, "big"), relations[43], entity(tDoc, rebac.Instance(fmt.Sprintf("many%02d", i)))))
	}

	candidates := append(slices.Clone(relations), "read", "write", "delete", "share", "admin", "chained", "cyclic", "inherited", rebac.Member, rebac.Parent, "unknown", "")
	sources := []rebac.Entity{
		entity(tUser, "u1"), entity(tUser, "u2"), entity(tUser, "u3"), entity(tUser, "u4"), entity(tUser, "u5"), entity(tUser, "crowd007"),
		entity(tRole, "big"), entity(tRole, "rc1"), entity(tGroup, "g2"), entity(tDoc, "d1"), entity(tUser, "unknown"), entity(tUser, ""),
	}
	targets := []rebac.Entity{
		tGlobal, entity(tDoc, "d1"), entity(tDoc, "d2"), entity(tDoc, "many07"), entity(tDoc, "never-seen"), entity(tDoc, rebac.AllInstances),
		entity(tUser, "u1"), entity(tUser, "crowd007"), entity(tRole, "rc1"), entity(rebac.Global, "something"), entity(tDoc, ""),
	}

	granted := 0
	for _, src := range sources {
		for _, tgt := range targets {
			set := db.Relations(src, tgt)
			all := slices.Collect(set.All())
			if len(all) != set.Len() {
				t.Errorf("%s %s: All yields %d relations, Len is %d", src, tgt, len(all), set.Len())
			}

			for _, rel := range candidates {
				want, err := db.Resolve(triple(src, rel, tgt))
				if err != nil {
					t.Fatal(err)
				}

				if got := set.Has(rel); got != want {
					t.Errorf("%s %s %s: got %v, want %v", src, rel, tgt, got, want)
				}

				if want {
					granted++
					if !slices.Contains(all, rel) {
						t.Errorf("%s %s %s: All must yield the relation", src, rel, tgt)
					}
				}
			}
		}
	}

	if granted < 100 {
		t.Fatalf("only %d combinations have been granted", granted)
	}

	if set := db.Relations(entity(tUser, ""), tGlobal); set.Len() != 0 {
		t.Error("an incomplete source has no relations")
	}

	var zero rebac.RelationSet
	if zero.Has("read") || zero.Len() != 0 || len(slices.Collect(zero.All())) != 0 {
		t.Error("the zero value contains no relation")
	}
}

func TestAddRule(t *testing.T) {
	db, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	db.AddRule(rebac.Rule{Source: tUser, Via: tRole})
	db.AddRule(rebac.Rule{Source: tUser, Via: tRole, Membership: rebac.Member}) // the same rule
	db.AddRule(rebac.Rule{Source: tUser, Via: tGroup})
	db.AddRule(rebac.Rule{Source: "", Via: tGroup}) // ignored
	db.AddRule(rebac.Rule{Source: tUser, Via: ""})  // ignored

	want := []rebac.Rule{
		{Source: tUser, Via: tRole, Membership: rebac.Member},
		{Source: tUser, Via: tGroup, Membership: rebac.Member},
	}

	if got := slices.Collect(db.Rules()); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// newBenchDB builds a database like a real application: the user is member of one role and one group, the role
// carries a realistic amount of permissions and a crowd of other members.
func newBenchDB(b *testing.B) *rebac.DB {
	b.Helper()

	db, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		b.Fatal(err)
	}

	db.ValidateStaticRules(false)
	db.AddRule(rebac.Rule{Source: tUser, Via: tRole})
	db.AddRule(rebac.Rule{Source: tUser, Via: tGroup})

	put := func(tr rebac.Triple) {
		if err := db.Put(tr); err != nil {
			b.Fatal(err)
		}
	}

	put(triple(entity(tRole, "editor"), rebac.Member, entity(tUser, "u1")))
	put(triple(entity(tGroup, "team"), rebac.Member, entity(tUser, "u1")))
	put(triple(entity(tRole, "editor"), "by-role", tGlobal))
	put(triple(entity(tGroup, "team"), "by-group", tGlobal))
	put(triple(entity(tRole, "admin"), "missing", tGlobal))
	for i := range 100 {
		put(triple(entity(tRole, "editor"), rebac.Relation(fmt.Sprintf("other%03d", i)), tGlobal))
	}

	for i := range 300 {
		put(triple(entity(tRole, "editor"), rebac.Member, entity(tUser, rebac.Instance(fmt.Sprintf("crowd%03d", i)))))
	}

	return db
}

func BenchmarkResolve(b *testing.B) {
	db := newBenchDB(b)

	for name, rel := range map[string]rebac.Relation{"granted-by-role": "by-role", "granted-by-group": "by-group", "denied": "missing"} {
		b.Run(name, func(b *testing.B) {
			tr := triple(entity(tUser, "u1"), rel, tGlobal)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := db.Resolve(tr); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRelations(b *testing.B) {
	db := newBenchDB(b)
	src := entity(tUser, "u1")

	b.ReportAllocs()
	for b.Loop() {
		if db.Relations(src, tGlobal).Len() != 102 {
			b.Fatal("unexpected relations")
		}
	}
}

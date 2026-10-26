// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"math/rand/v2"
	"slices"
	"testing"

	"go.wdy.de/nago/pkg/blob/mem"
)

func TestPlan(t *testing.T) {
	for _, c := range []struct {
		name    string
		pattern binaryTriple
		prefix  int
		reverse bool
	}{
		{"nothing", binaryTriple{}, 0, false},
		{"source", binaryTriple{A: bEntity{1, 2}}, 8, false},
		{"source namespace", binaryTriple{A: bEntity{Namespace: 1}}, 4, false},
		{"target", binaryTriple{B: bEntity{1, 2}}, 8, true},
		{"target namespace", binaryTriple{B: bEntity{Namespace: 1}}, 4, true},
		{"source namespace and target", binaryTriple{A: bEntity{Namespace: 1}, B: bEntity{1, 2}}, 8, true},
		{"source and target", binaryTriple{A: bEntity{1, 2}, B: bEntity{3, 4}}, 8, false},
		{"source, relation and target namespace", binaryTriple{A: bEntity{1, 2}, Relation: 3, B: bEntity{Namespace: 4}}, 16, false},
		{"all memberships of a user", binaryTriple{A: bEntity{Namespace: 1}, Relation: 3, B: bEntity{4, 5}}, 16, true},
		{"exact", binaryTriple{A: bEntity{1, 2}, Relation: 3, B: bEntity{4, 5}}, 20, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			pivot, prefix, reverse := plan(c.pattern)
			if prefix != c.prefix || reverse != c.reverse {
				t.Errorf("got prefix %d reverse %v, want %d %v", prefix, reverse, c.prefix, c.reverse)
			}

			want := c.pattern
			if reverse {
				want = want.reverse()
			}

			if pivot != want.Binary() {
				t.Errorf("pivot %v does not match the pattern", pivot)
			}
		})
	}
}

// TestScanMatchesFullScan ensures that the planned range scan yields exactly the triples, which a full scan with
// the pattern filter yields, for random patterns with arbitrary holes.
func TestScanMatchesFullScan(t *testing.T) {
	db, err := NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	db.ValidateStaticRules(false)
	rnd := rand.New(rand.NewPCG(3, 4))
	id := func(n int) uint32 { return uint32(rnd.IntN(n) + 1) }

	// few distinct ids, so that patterns match a lot
	var all []binaryTriple
	for range 2000 {
		bt := binaryTriple{A: bEntity{bNamespace(id(3)), bInstance(id(20))}, Relation: bRelation(id(5)), B: bEntity{bNamespace(id(3)), bInstance(id(20))}}
		db.forward.Set(bt.Binary())
		db.backward.Set(bt.reverse().Binary())
		all = append(all, bt)
	}

	less := func(a, b binaryTriple) int {
		x, y := a.Binary(), b.Binary()
		return slices.Compare(x[:], y[:])
	}

	for range 500 {
		hole := func(v uint32) uint32 {
			if rnd.IntN(2) == 0 {
				return 0
			}

			return v
		}

		pattern := binaryTriple{
			A:        bEntity{bNamespace(hole(id(3))), bInstance(hole(id(20)))},
			Relation: bRelation(hole(id(5))),
			B:        bEntity{bNamespace(hole(id(3))), bInstance(hole(id(20)))},
		}

		var want []binaryTriple
		for _, bt := range all {
			if bt.matches(pattern) {
				want = append(want, bt)
			}
		}

		var got []binaryTriple
		db.scan(pattern, func(bt binaryTriple) bool {
			got = append(got, bt)
			return true
		})

		slices.SortFunc(want, less)
		want = slices.Compact(want)
		slices.SortFunc(got, less)

		if !slices.Equal(got, want) {
			t.Fatalf("pattern %+v: got %d triples, want %d", pattern, len(got), len(want))
		}
	}

	// the scan stops when the consumer is done
	n := 0
	db.scan(binaryTriple{}, func(binaryTriple) bool {
		n++
		return n < 3
	})

	if n != 3 {
		t.Errorf("scan visited %d triples after the consumer stopped", n)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/xslices"
)

func TestDB2_Put(t *testing.T) {
	t.Log("allocating test set")
	testSet := createTestSet()
	t.Logf("test set allocated %d entries, inserting...", len(testSet))
	store := mem.NewBlobStore("test")
	db := option.Must(rebac.NewDB(store))
	option.MustZero(db.PutAll(xslices.ValuesWithError(testSet, nil)))
	t.Log("checking inserted tuples")
	start := time.Now()
	for range 5 {

		for _, tuple := range testSet {
			if !option.Must(db.Contains(tuple)) {
				t.Fatal("tuple not found in db", tuple)
			}

			tuple.Target.Instance = ""
			if option.Must(db.Contains(tuple)) {
				t.Fatal("tuple found in db with empty instance", tuple)
			}
		}
	}

	t.Log("read done in", time.Since(start))

	// check replay
	db = option.Must(rebac.NewDB(store))
	runtime.GC()
	for _, tuple := range testSet {
		if !option.Must(db.Contains(tuple)) {
			t.Fatal("tuple not found in db", tuple)
		}

		tuple.Target.Instance = ""
		if option.Must(db.Contains(tuple)) {
			t.Fatal("tuple found in db with empty instance", tuple)
		}
	}

	// check range
	relCount := 0
	for tuple := range db.Query(rebac.Select().Where().Source().Is("subj-ns-10", "subj-inst-10")) {
		if tuple.Source.Namespace != "subj-ns-10" || tuple.Source.Instance != "subj-inst-10" {
			t.Fatal("tuple not found in db", tuple)
		}
		relCount++
	}
	if relCount != 10 {
		t.Fatal("expected 10 relations for subj-ns-10 subj-inst-10, got", relCount)
	}

	// check range with rel
	relCount = 0
	for tuple := range db.Query(rebac.Select().Where().Source().Is("subj-ns-10", "subj-inst-10").Where().Relation().Has("rel-2")) {
		if tuple.Source.Namespace != "subj-ns-10" || tuple.Source.Instance != "subj-inst-10" {
			t.Fatal("tuple not found in db", tuple)
		}
		relCount++
	}

	if relCount != 1 {
		t.Fatal("expected 10 relations for subj-ns-10 subj-inst-10, got", relCount)
	}

	if i := option.Must(db.Count()); int(i) != len(testSet) {
		t.Fatal("expected count to be equal to number of tuples", i, len(testSet))
	}

	// delete one
	option.MustZero(db.Delete(testSet[0]))
	if i := option.Must(db.Count()); int(i) != len(testSet)-1 {
		t.Fatal("expected count to be equal to number of tuples", i, len(testSet)-1)
	}

	// delete all
	option.MustZero(db.DeleteByQuery(rebac.Select()))
	if c := option.Must(db.Count()); c != 0 {
		t.Fatal("expected count to be zero but got", c)
	}
}

func createTestSet() []rebac.Triple {
	nsSize := 100
	instSize := 1000
	relSize := 10
	res := make([]rebac.Triple, 0, nsSize*instSize*relSize)
	for ns := 0; ns < nsSize; ns++ {
		for in := 0; in < instSize; in++ {
			for rel := 0; rel < relSize; rel++ {
				res = append(res, rebac.Triple{
					Source: rebac.Entity{
						Namespace: rebac.Namespace(fmt.Sprintf("subj-ns-%d", ns)),
						Instance:  rebac.Instance(fmt.Sprintf("subj-inst-%d", in)),
					},
					Relation: rebac.Relation(fmt.Sprintf("rel-%d", rel)),
					Target: rebac.Entity{
						Namespace: rebac.Namespace(fmt.Sprintf("obj-ns-%d", ns)),
						Instance:  rebac.Instance(fmt.Sprintf("obj-inst-%d", in)),
					},
				})
			}

		}
	}

	return res
}

const (
	usrNS  = "nago.user"
	roleNS = "nago.role"
)

func TestDB2_Resolve(t *testing.T) {
	store := mem.NewBlobStore("test")
	db := option.Must(rebac.NewDB(store))
	option.MustZero(db.PutAll(xslices.ValuesWithError([]rebac.Triple{
		// user torben [is a] member [of] admin role
		{
			Source: rebac.Entity{
				Namespace: roleNS,
				Instance:  "admin",
			},
			Relation: rebac.Member,
			Target: rebac.Entity{
				Namespace: usrNS,
				Instance:  "torben",
			},
		},

		// user torben [is a] member [of] some group
		{
			Source: rebac.Entity{
				Namespace: "nago.group",
				Instance:  "some group",
			},
			Relation: rebac.Member,
			Target: rebac.Entity{
				Namespace: usrNS,
				Instance:  "torben",
			},
		},

		// admin role [allows to] create_users [in] nago.resource test
		{
			Source: rebac.Entity{
				Namespace: roleNS,
				Instance:  "admin",
			},
			Relation: "create_user",
			Target: rebac.Entity{
				Namespace: "nago.resource",
				Instance:  "test",
			},
		},

		// admin role [allows to] delete [all] nago.resource
		{
			Source: rebac.Entity{
				Namespace: roleNS,
				Instance:  "admin",
			},
			Relation: "delete",
			Target: rebac.Entity{
				Namespace: "nago.resource",
				Instance:  rebac.AllInstances,
			},
		},
	}, nil)))

	// query is if torben can create_user in nago.resource test
	q1 := rebac.Triple{
		Source: rebac.Entity{
			Namespace: usrNS,
			Instance:  "torben",
		},
		Relation: "create_user",
		Target: rebac.Entity{
			Namespace: "nago.resource",
			Instance:  "test",
		},
	}

	// without resolver must fail
	if ok := option.Must(db.Resolve(q1)); ok {
		t.Fatal("expected resolve to be !ok")
	}

	// with resolver must ok
	db.AddResolver(rebac.NewSourceMemberResolver(usrNS, roleNS))
	if ok := option.Must(db.Resolve(q1)); !ok {
		t.Fatal("expected resolve to be ok")
	}

	// query 2
	q2 := rebac.Triple{
		Source: rebac.Entity{
			Namespace: usrNS,
			Instance:  "torben",
		},
		Relation: "delete",
		Target: rebac.Entity{
			Namespace: "nago.resource",
			Instance:  "something specific but admin allows wildcard",
		},
	}

	if ok := option.Must(db.Resolve(q2)); !ok {
		t.Fatal("expected resolve to be ok")
	}

	// query 3
	allRolesOfUser := rebac.Select().
		Where().Target().Is("nago.user", "torben").
		Where().Relation().Has("member").
		Where().Source().IsNamespace("nago.role")

	roles := option.Must(xslices.Collect2(db.Query(allRolesOfUser)))
	if len(roles) != 1 {
		t.Fatal("expected 1 role")
	}

	// query 4
	allRolesOfUser2 := rebac.Select().
		Where().Target().Is("nago.user", "torben").
		// relation==?
		Where().Source().IsNamespace("nago.role")

	roles = option.Must(xslices.Collect2(db.Query(allRolesOfUser2)))
	for _, role := range roles {
		if role.Target.Namespace != "nago.user" {
			t.Fatal("expected role to be from nago.user namespace")
		}

		if role.Target.Instance != "torben" {
			t.Fatal("expected role to be for torben")
		}

		if role.Source.Namespace != "nago.role" {
			t.Fatal("expected role to be from nago.role namespace but found", role.Source.Namespace)
		}

	}
	if len(roles) != 1 {
		t.Fatal("expected 1 role but found", len(roles))
	}

}

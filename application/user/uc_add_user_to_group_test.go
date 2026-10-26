// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"slices"
	"testing"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

// TestAddUserToGroup ensures that the membership is stored in the same direction as UpdateOtherGroups writes it and
// the static rule of the group management allows it: group has-member user.
func TestAddUserToGroup(t *testing.T) {
	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	// the only allowed membership direction, see Configurator.GroupManagement
	rdb.RegisterStaticRule(rebac.StaticRule{Source: group.Namespace, Relation: rebac.Member, Target: Namespace})

	repo := Repository(json.NewSloppyJSONRepository[User, ID](mem.NewBlobStore("user")))
	usr := User{ID: "u1", EMailVerified: true, Status: Enabled{}}
	if err := repo.Save(usr); err != nil {
		t.Fatal(err)
	}

	const gid group.ID = "g1"
	if err := NewAddUserToGroup(rdb)(SU(), usr.ID, gid); err != nil {
		t.Fatalf("cannot add user to group: %v", err)
	}

	membership := rebac.Triple{
		Source:   rebac.Entity{Namespace: group.Namespace, Instance: rebac.Instance(gid)},
		Relation: rebac.Member,
		Target:   rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(usr.ID)},
	}

	if ok, _ := rdb.Contains(membership); !ok {
		t.Fatal("membership must be stored as group:has-member:user")
	}

	subject := newViewImpl(context.Background, rdb, repo, usr)
	if !subject.HasGroup(gid) {
		t.Error("the subject must be in the group")
	}

	if got := slices.Collect(subject.Groups()); !slices.Equal(got, []group.ID{gid}) {
		t.Errorf("Groups yields %v, want [%s]", got, gid)
	}
}

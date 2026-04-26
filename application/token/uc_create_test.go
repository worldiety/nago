// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"context"
	"sync"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func TestCreatePersistsResourcePermissions(t *testing.T) {
	tokenRepo := json.NewSloppyJSONRepository[Token](mem.NewBlobStore("tokens"))
	rdb := option.Must(rebac.NewDB(mem.NewBlobStore("rebac")))
	perm := permission.ID("test.token.resource")
	resource := user.Resource{Name: "test.resource", ID: "resource-1"}

	rdb.RegisterStaticRule(rebac.StaticRule{Source: Namespace, Relation: rebac.Relation(perm), Target: rebac.Namespace(resource.Name)})

	create := NewCreate(&sync.Mutex{}, tokenRepo, user.Argon2IdMin, &concurrent.RWMap[Hash, ID]{}, rdb)
	id, _, err := create(user.SU(), CreationData{
		Name:      "upload token",
		Plaintext: "0123456789abcdef",
		Resources: map[user.Resource][]permission.ID{
			resource: {perm},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	subject := newSubject(context.Background(), nil, tokenRepo, option.Must(tokenRepo.FindByID(id)).Unwrap(), rdb)
	if !subject.HasResourcePermission(rebac.Namespace(resource.Name), rebac.Instance(resource.ID), perm) {
		t.Fatal("expected token subject to have resource permission")
	}
}

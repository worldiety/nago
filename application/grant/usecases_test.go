// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package grant

import (
	"go.wdy.de/nago/application/user"
	"testing"
)

func TestID_Split(t *testing.T) {
	id0 := NewID(user.Resource{
		Name: "orgs",
		ID:   "1234",
	}, "567")

	if id0 != "orgs/1234/567" {
		t.Fatalf("want orgs/1234/567, got %s", id0)
	}

	if !id0.Valid() {
		t.Fatalf("id0 should be valid")
	}

	res, uid := id0.Split()
	if res.Name != "orgs" {
		t.Fatalf("want orgs, got %s", res.Name)
	}

	if res.ID != "1234" {
		t.Fatalf("want 1234, got %s", res.ID)
	}

	if uid != "567" {
		t.Fatalf("want 567, got %s", uid)
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
)

func TestAccessSummary(t *testing.T) {
	me := user.SU()
	names := func(_ permission.Auditable, id group.ID) (option.Opt[group.Group], error) {
		return option.Some(group.Group{ID: id, Name: "Name of " + string(id)}), nil
	}

	tests := []struct {
		scr  secret.Secret
		want string
	}{
		{secret.Secret{Owners: []user.ID{me.ID()}}, "Nur Sie"},
		{secret.Secret{Owners: []user.ID{me.ID(), "a"}}, "Geteilt mit 1 weiteren Person"},
		{secret.Secret{Owners: []user.ID{me.ID(), "a", "b"}, Groups: []group.ID{"g1", "g2"}}, "Geteilt mit 2 weiteren Personen und den Gruppen Name of g1 und Name of g2"},
		{secret.Secret{Owners: []user.ID{me.ID()}, Groups: []group.ID{"g1"}}, "Geteilt mit der Gruppe Name of g1"},
	}

	for _, tt := range tests {
		if got := accessSummary(me, names, tt.scr); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestJoinAnd(t *testing.T) {
	if joinAnd([]string{"a"}) != "a" || joinAnd([]string{"a", "b", "c"}) != "a, b und c" {
		t.Fatal("unexpected join")
	}
}

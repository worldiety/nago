// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"reflect"
	"testing"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

func TestMatch(t *testing.T) {
	repo := json.NewSloppyJSONRepository[Secret, ID](mem.NewBlobStore("secrets"))
	for _, s := range []Secret{
		// saved first on purpose: a secret of another group must not end the search
		{ID: "a-foreign", Owners: []user.ID{stranger}, Groups: []group.ID{grpForeign}, Credentials: testCredentials{Name: "foreign"}},
		{ID: "b-personal", Owners: []user.ID{owner}, Credentials: testCredentials{Name: "personal"}},
		{ID: "c-shared", Owners: []user.ID{stranger}, Groups: []group.ID{grpShared}, Credentials: testCredentials{Name: "shared"}},
		{ID: "d-shared2", Owners: []user.ID{stranger}, Groups: []group.ID{grpShared}, Credentials: testCredentials{Name: "preferred"}},
	} {
		if err := repo.Save(s); err != nil {
			t.Fatal(err)
		}
	}

	match := NewMatch(repo)
	typ := reflect.TypeFor[testCredentials]()

	tests := []struct {
		name    string
		subject testSubject
		opts    MatchOptions
		want    string // empty means none
		wantErr bool
	}{
		{name: "group member finds group secret", subject: sub(member, grpShared), opts: MatchOptions{Group: grpShared}, want: "shared"},
		{name: "hint prefers named secret", subject: sub(member, grpShared), opts: MatchOptions{Group: grpShared, Hint: "preferred"}, want: "preferred"},
		{name: "owner finds personal secret", subject: sub(owner), want: "personal"},
		{name: "owner has no access to group", subject: sub(owner), opts: MatchOptions{Group: grpShared}},
		{name: "stranger to group finds nothing", subject: sub(member, grpOther), opts: MatchOptions{Group: grpShared}},
		{name: "expect fails with group filter", subject: sub(member, grpOther), opts: MatchOptions{Group: grpShared, Expect: true}, wantErr: true},
		{name: "invalid subject finds nothing", subject: testSubject{id: owner}, opts: MatchOptions{Expect: true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := match(tt.subject, typ, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error %v", err)
			}

			var name string
			if got.IsSome() {
				name = got.Unwrap().GetName()
			}

			if name != tt.want {
				t.Fatalf("want %q, got %q", tt.want, name)
			}
		})
	}
}

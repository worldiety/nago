// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"errors"
	"slices"
	"testing"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
)

type writeOnlyCredentials struct {
	Name     string `json:"name"`
	Password string `json:"password" style:"secret" credentialWriteOnly:"true"`
	Hint     string `json:"hint" style:"secret"` // secret style alone is not write-only
}

var _ = enum.Variant[Credentials, writeOnlyCredentials](enum.Rename[writeOnlyCredentials]("nago.secret.test.writeonly"))

func (c writeOnlyCredentials) GetName() string   { return c.Name }
func (c writeOnlyCredentials) Credentials() bool { return true }
func (c writeOnlyCredentials) IsZero() bool      { return c == writeOnlyCredentials{} }

func TestWithoutWriteOnly(t *testing.T) {
	c := writeOnlyCredentials{Name: "smtp", Password: "pw", Hint: "hint"}
	got := WithoutWriteOnly(c).(writeOnlyCredentials)
	if got != (writeOnlyCredentials{Name: "smtp", Hint: "hint"}) {
		t.Fatalf("expected only the write-only field to be cleared, got %+v", got)
	}

	if c.Password != "pw" {
		t.Fatal("the original must not be changed")
	}

	if fields := WriteOnlyFields(c); len(fields) != 1 || !fields["Password"] {
		t.Fatalf("unexpected write-only fields %v", fields)
	}

	if fields := WriteOnlyFields(writeOnlyCredentials{}); fields["Password"] {
		t.Fatal("expected no stored password")
	}
}

// A form which has never received the stored password must not delete it, but it can replace it.
func TestUpdateKeepsEmptyWriteOnlyFields(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	id, err := uc.CreateSecret(sub(owner), writeOnlyCredentials{Name: "smtp", Password: "pw", Hint: "hint"})
	if err != nil {
		t.Fatal(err)
	}

	if err := uc.UpdateMyCredentials(sub(owner), id, writeOnlyCredentials{Name: "renamed"}); err != nil {
		t.Fatal(err)
	}

	stored := std.Must(repo.FindByID(id)).Unwrap().Credentials.(writeOnlyCredentials)
	if stored != (writeOnlyCredentials{Name: "renamed", Password: "pw"}) {
		t.Fatalf("expected the password to be kept and the hint to be cleared, got %+v", stored)
	}

	if err := uc.UpdateMyCredentials(sub(owner), id, writeOnlyCredentials{Name: "renamed", Password: "new"}); err != nil {
		t.Fatal(err)
	}

	stored = std.Must(repo.FindByID(id)).Unwrap().Credentials.(writeOnlyCredentials)
	if stored.Password != "new" {
		t.Fatalf("expected the password to be replaced, got %+v", stored)
	}
}

// A new secret is shared at once, so that it is never stored half way.
func TestCreateSecretShares(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	id, err := uc.CreateSecret(sub(owner, grpShared), testCredentials{Name: "new"}, CreateOptions{
		Groups: []group.ID{grpShared},
		Owners: []user.ID{member},
	})
	if err != nil {
		t.Fatal(err)
	}

	scr := std.Must(repo.FindByID(id)).Unwrap()
	if !slices.Equal(scr.Groups, []group.ID{grpShared}) || !slices.Equal(scr.Owners, []user.ID{owner, member}) {
		t.Fatalf("unexpected sharing %v %v", scr.Groups, scr.Owners)
	}
}

func TestCreateSecretRejectsForeignGroups(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	before := std.Must(repo.Count())
	_, err := uc.CreateSecret(sub(owner, grpShared), testCredentials{Name: "new"}, CreateOptions{Groups: []group.ID{grpForeign}})
	if err == nil {
		t.Fatal("expected a foreign group to be rejected")
	}

	_, err = uc.CreateSecret(sub(owner, grpShared).without(PermUpdateMySecretGroups), testCredentials{Name: "new"}, CreateOptions{Groups: []group.ID{grpShared}})
	if !errors.Is(err, user.PermissionDeniedErr) {
		t.Fatalf("expected the permission to share to be required, got %v", err)
	}

	if std.Must(repo.Count()) != before {
		t.Fatal("expected no secret to be stored")
	}
}

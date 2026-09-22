// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"context"
	"errors"
	"iter"
	"slices"
	"testing"

	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"golang.org/x/text/language"
)

// testCredentials is a minimal credential variant, so that the repository can marshal and unmarshal the
// open sum type [Credentials].
type testCredentials struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

var _ = enum.Variant[Credentials, testCredentials](enum.Rename[testCredentials]("nago.secret.test.credentials"))

func (c testCredentials) GetName() string   { return c.Name }
func (c testCredentials) Credentials() bool { return true }
func (c testCredentials) IsZero() bool      { return c == testCredentials{} }

const (
	owner    user.ID = "owner"
	member   user.ID = "member"
	stranger user.ID = "stranger"

	grpShared  group.ID = "shared"
	grpForeign group.ID = "foreign"
	grpOther   group.ID = "other"
)

// testSubject is an in-memory subject. It grants every declared secret permission unless explicitly denied,
// so that the tests isolate the access rules from the permission audit.
type testSubject struct {
	id     user.ID
	groups []group.ID
	denied []permission.ID
	valid  bool
}

func (s testSubject) ID() user.ID { return s.id }

func (s testSubject) HasPermission(p permission.ID) bool {
	return s.valid && !slices.Contains(s.denied, p)
}

func (s testSubject) Audit(p permission.ID) error {
	if s.HasPermission(p) {
		return nil
	}
	return user.PermissionDeniedErr
}

func (s testSubject) HasResourcePermission(rebac.Namespace, rebac.Instance, permission.ID) bool {
	return false
}

func (s testSubject) AuditResource(rebac.Namespace, rebac.Instance, permission.ID) error {
	return user.PermissionDeniedErr
}

func (s testSubject) HasGroup(id group.ID) bool { return slices.Contains(s.groups, id) }

func (s testSubject) Groups() iter.Seq[group.ID] { return slices.Values(s.groups) }

func (s testSubject) Roles() iter.Seq[role.ID] { return func(yield func(role.ID) bool) {} }

func (s testSubject) HasRole(role.ID) bool     { return false }
func (s testSubject) Valid() bool              { return s.valid }
func (s testSubject) Name() string             { return string(s.id) }
func (s testSubject) Firstname() string        { return "" }
func (s testSubject) Lastname() string         { return "" }
func (s testSubject) Email() string            { return "" }
func (s testSubject) Avatar() string           { return "" }
func (s testSubject) Language() language.Tag   { return language.German }
func (s testSubject) Bundle() *i18n.Bundle     { return nil }
func (s testSubject) Context() context.Context { return context.Background() }

func sub(id user.ID, groups ...group.ID) testSubject {
	return testSubject{id: id, groups: groups, valid: true}
}

func (s testSubject) without(p permission.ID) testSubject {
	s.denied = append(slices.Clone(s.denied), p)
	return s
}

// newTestUseCases wires the real use cases against an in-memory repository holding a single secret which is
// owned by [owner] and shared into grpShared and grpForeign.
func newTestUseCases(t *testing.T) (UseCases, Repository, ID) {
	t.Helper()

	repo := json.NewSloppyJSONRepository[Secret, ID](mem.NewBlobStore("secrets"))
	uc := NewUseCases(events.NewEventBus(), repo)

	const id ID = "the-secret"
	err := repo.Save(Secret{
		ID:          id,
		Owners:      []user.ID{owner},
		Groups:      []group.ID{grpShared, grpForeign},
		Credentials: testCredentials{Name: "initial", Token: "s3cr3t"},
	})
	if err != nil {
		t.Fatalf("cannot save secret: %v", err)
	}

	return uc, repo, id
}

func mustLoad(t *testing.T, repo Repository, id ID) Secret {
	t.Helper()
	opt, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("cannot load secret: %v", err)
	}
	if opt.IsNone() {
		t.Fatalf("secret %v has been removed unexpectedly", id)
	}
	return opt.Unwrap()
}

// TestSecretHasAccess verifies the central access rule: an owner or a member of any group the secret has been
// shared into has access, everyone else has not.
func TestSecretHasAccess(t *testing.T) {
	cases := []struct {
		name    string
		secret  Secret
		subject auth.Subject
		want    bool
	}{
		{"owner without groups", Secret{Owners: []user.ID{owner}}, sub(owner), true},
		{"owner of grouped secret", Secret{Owners: []user.ID{owner}, Groups: []group.ID{grpShared}}, sub(owner), true},
		{"group member", Secret{Owners: []user.ID{owner}, Groups: []group.ID{grpShared}}, sub(member, grpShared), true},
		{"member of one of many groups", Secret{Owners: []user.ID{owner}, Groups: []group.ID{grpForeign, grpShared}}, sub(member, grpShared), true},
		{"stranger with other group", Secret{Owners: []user.ID{owner}, Groups: []group.ID{grpShared}}, sub(stranger, grpOther), false},
		{"stranger without groups", Secret{Owners: []user.ID{owner}}, sub(stranger), false},
		{"secret without groups is private", Secret{Owners: []user.ID{owner}}, sub(member, grpShared), false},
		{"invalid subject", Secret{Owners: []user.ID{owner}, Groups: []group.ID{grpShared}}, testSubject{id: owner, groups: []group.ID{grpShared}}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.secret.HasAccess(c.subject); got != c.want {
				t.Errorf("HasAccess = %v, want %v", got, c.want)
			}
		})
	}
}

// TestFindMySecretsIncludesGroupSecrets verifies that a group member sees the shared secret, whereas an
// unrelated subject does not.
func TestFindMySecretsIncludesGroupSecrets(t *testing.T) {
	uc, _, id := newTestUseCases(t)

	cases := []struct {
		name    string
		subject testSubject
		want    bool
	}{
		{"owner", sub(owner), true},
		{"group member", sub(member, grpShared), true},
		{"stranger", sub(stranger, grpOther), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var found bool
			for scr, err := range uc.FindMySecrets(c.subject) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if scr.ID == id {
					found = true
				}
			}

			if found != c.want {
				t.Errorf("found = %v, want %v", found, c.want)
			}
		})
	}
}

// TestFindMySecretByIDAccess verifies reading a single secret including the permission audit.
func TestFindMySecretByIDAccess(t *testing.T) {
	uc, _, id := newTestUseCases(t)

	t.Run("group member may read", func(t *testing.T) {
		opt, err := uc.FindMySecretByID(sub(member, grpShared), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.IsNone() {
			t.Fatal("expected the secret to be readable")
		}
	})

	t.Run("stranger is denied", func(t *testing.T) {
		_, err := uc.FindMySecretByID(sub(stranger, grpOther), id)
		if !errors.Is(err, AccessDeniedErr) {
			t.Fatalf("expected AccessDeniedErr, got %v", err)
		}
	})

	t.Run("missing permission wins over group membership", func(t *testing.T) {
		_, err := uc.FindMySecretByID(sub(member, grpShared).without(PermFindMySecrets), id)
		if !errors.Is(err, user.PermissionDeniedErr) {
			t.Fatalf("expected PermissionDeniedErr, got %v", err)
		}
	})
}

// TestUpdateMyCredentialsAccess verifies that a group member may rewrite the credentials but a stranger may not.
func TestUpdateMyCredentialsAccess(t *testing.T) {
	t.Run("group member may update", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		if err := uc.UpdateMyCredentials(sub(member, grpShared), id, testCredentials{Name: "changed"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := mustLoad(t, repo, id).Credentials.GetName(); got != "changed" {
			t.Errorf("name = %q, want %q", got, "changed")
		}
	})

	t.Run("stranger is denied", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		err := uc.UpdateMyCredentials(sub(stranger, grpOther), id, testCredentials{Name: "changed"})
		if !errors.Is(err, AccessDeniedErr) {
			t.Fatalf("expected AccessDeniedErr, got %v", err)
		}

		if got := mustLoad(t, repo, id).Credentials.GetName(); got != "initial" {
			t.Errorf("credentials have been modified: %q", got)
		}
	})

	t.Run("missing permission is denied", func(t *testing.T) {
		uc, _, id := newTestUseCases(t)
		err := uc.UpdateMyCredentials(sub(member, grpShared).without(PermUpdateMyCredentials), id, testCredentials{Name: "changed"})
		if !errors.Is(err, user.PermissionDeniedErr) {
			t.Fatalf("expected PermissionDeniedErr, got %v", err)
		}
	})
}

// TestDeleteMySecretByIDAccess verifies that deleting is bound to the access rule. Before the group based
// access has been introduced, any holder of the delete permission was able to remove any secret.
func TestDeleteMySecretByIDAccess(t *testing.T) {
	t.Run("group member may delete", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		if err := uc.DeleteMySecretByID(sub(member, grpShared), id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opt, err := repo.FindByID(id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.IsSome() {
			t.Error("secret still exists")
		}
	})

	t.Run("stranger is denied and secret survives", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		err := uc.DeleteMySecretByID(sub(stranger, grpOther), id)
		if !errors.Is(err, AccessDeniedErr) {
			t.Fatalf("expected AccessDeniedErr, got %v", err)
		}

		mustLoad(t, repo, id)
	})

	t.Run("deleting an unknown secret is idempotent", func(t *testing.T) {
		uc, _, _ := newTestUseCases(t)
		if err := uc.DeleteMySecretByID(sub(owner), "does-not-exist"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestUpdateMySecretGroups verifies the delta rule: foreign groups survive, adding requires a membership and
// a non owner must not lock itself out.
func TestUpdateMySecretGroups(t *testing.T) {
	t.Run("foreign groups survive a save of a group member", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		// the member only knows grpShared, so the UI would submit exactly that
		if err := uc.UpdateMySecretGroups(sub(member, grpShared), id, []group.ID{grpShared}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := mustLoad(t, repo, id).Groups
		if !slices.Contains(got, grpForeign) {
			t.Errorf("foreign group has been dropped: %v", got)
		}
		if !slices.Contains(got, grpShared) {
			t.Errorf("shared group has been dropped: %v", got)
		}
	})

	t.Run("owner may remove its own group", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		if err := uc.UpdateMySecretGroups(sub(owner, grpShared, grpForeign), id, []group.ID{grpForeign}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := mustLoad(t, repo, id).Groups; slices.Contains(got, grpShared) {
			t.Errorf("shared group has not been removed: %v", got)
		}
	})

	t.Run("adding a group without membership is denied", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		err := uc.UpdateMySecretGroups(sub(owner), id, []group.ID{grpShared, grpForeign, grpOther})
		if err == nil {
			t.Fatal("expected an error")
		}

		if got := mustLoad(t, repo, id).Groups; slices.Contains(got, grpOther) {
			t.Errorf("group has been added anyway: %v", got)
		}
	})

	t.Run("non owner must not lock itself out", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		err := uc.UpdateMySecretGroups(sub(member, grpShared), id, nil)
		if err == nil {
			t.Fatal("expected an error")
		}

		if got := mustLoad(t, repo, id).Groups; !slices.Contains(got, grpShared) {
			t.Errorf("member locked itself out: %v", got)
		}
	})

	t.Run("stranger is denied", func(t *testing.T) {
		uc, _, id := newTestUseCases(t)
		err := uc.UpdateMySecretGroups(sub(stranger, grpOther), id, []group.ID{grpOther})
		if !errors.Is(err, AccessDeniedErr) {
			t.Fatalf("expected AccessDeniedErr, got %v", err)
		}
	})
}

// TestUpdateMySecretOwners verifies that an owner keeps its ownership whereas a group member does not
// promote itself to an owner.
func TestUpdateMySecretOwners(t *testing.T) {
	t.Run("owner keeps its ownership", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		if err := uc.UpdateMySecretOwners(sub(owner), id, []user.ID{stranger}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := mustLoad(t, repo, id).Owners; !slices.Contains(got, owner) {
			t.Errorf("owner removed itself: %v", got)
		}
	})

	t.Run("group member does not promote itself", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		if err := uc.UpdateMySecretOwners(sub(member, grpShared), id, []user.ID{owner}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := mustLoad(t, repo, id).Owners; slices.Contains(got, member) {
			t.Errorf("member promoted itself to an owner: %v", got)
		}
	})

	t.Run("stranger is denied", func(t *testing.T) {
		uc, repo, id := newTestUseCases(t)
		err := uc.UpdateMySecretOwners(sub(stranger, grpOther), id, []user.ID{stranger})
		if !errors.Is(err, AccessDeniedErr) {
			t.Fatalf("expected AccessDeniedErr, got %v", err)
		}

		if got := mustLoad(t, repo, id).Owners; slices.Contains(got, stranger) {
			t.Errorf("stranger became an owner: %v", got)
		}
	})
}

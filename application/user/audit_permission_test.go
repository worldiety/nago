// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"strings"
	"testing"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"golang.org/x/text/language"
)

// The two ways a permission can acquire a name. Both used to be destroyed by Audit, so both
// have to be covered: Declare stores the literal, the Declare*-family from crud.go stores an
// i18n handle of the form "@1234".
type auditProbePlainUseCase func()

type auditProbeI18nUseCase func()

var (
	permAuditProbePlain = permission.Declare[auditProbePlainUseCase](
		"nago.user.audit_probe_plain",
		"Auditprobe Klartext",
		"Nur für Tests.",
	)

	permAuditProbeI18n = permission.DeclareFindByID[auditProbeI18nUseCase](
		"nago.user.audit_probe_i18n",
		"Auditprobe",
	)
)

// newAuditProbeView builds the real viewImpl rather than a fake subject. The defect under test
// lived in viewImpl.Audit itself, so a hand written test double that reimplements Audit would
// have stayed green while production stayed broken.
func newAuditProbeView(t *testing.T) *viewImpl {
	t.Helper()

	repo := Repository(json.NewSloppyJSONRepository[User, ID](mem.NewBlobStore("user")))

	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatalf("cannot create rebac db: %v", err)
	}

	// the subject must pass Valid, otherwise Audit answers with the account error before it
	// ever reaches the permission check
	usr := User{ID: "42", EMailVerified: true, Status: Enabled{}}
	if err := repo.Save(usr); err != nil {
		t.Fatalf("cannot save user: %v", err)
	}

	return newViewImpl(context.Background, rdb, repo, usr)
}

// TestAuditNamesTheRequiredPermission is the regression guard. Audit used to replace the
// permission id with the display name of the permission, and a display name never satisfies
// permission.ID.Valid. RequiredOf therefore reported false for every denial produced by Audit,
// and the user was told nothing about what was missing.
func TestAuditNamesTheRequiredPermission(t *testing.T) {
	tests := []struct {
		name     string
		perm     permission.ID
		wantName string
	}{
		{"literal name", permAuditProbePlain, "Auditprobe Klartext"},
		{"i18n handle", permAuditProbeI18n, "Auditprobe"},
	}

	bundle := auditProbeBundle(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newAuditProbeView(t)

			err := v.Audit(tt.perm)
			if err == nil {
				t.Fatal("a subject without any role must be denied")
			}

			if !permission.IsDenied(err) {
				t.Fatalf("must be classified as a denial, got %v", err)
			}

			perm, ok := permission.RequiredOf(err)
			if !ok {
				t.Fatalf("the denial must name the permission, got %q", err.Error())
			}

			if perm.ID != tt.perm {
				t.Errorf("got permission %q, want %q", perm.ID, tt.perm)
			}

			// the id alone is useless in prose, it has to survive localization
			got := perm.LocalizedName(bundle)
			if !strings.Contains(got, tt.wantName) {
				t.Errorf("localized name is %q, want it to contain %q", got, tt.wantName)
			}

			if strings.HasPrefix(got, "@") {
				t.Errorf("the i18n handle leaked into the display name: %q", got)
			}

			if got == string(tt.perm) {
				t.Errorf("the raw id leaked into the display name: %q", got)
			}
		})
	}
}

// TestAuditResourceNamesTheRequiredPermission covers the second entry point, which carried the
// same defect independently.
func TestAuditResourceNamesTheRequiredPermission(t *testing.T) {
	v := newAuditProbeView(t)

	err := v.AuditResource("nago.test", "some-instance", permAuditProbePlain)
	if err == nil {
		t.Fatal("a subject without any relation must be denied")
	}

	perm, ok := permission.RequiredOf(err)
	if !ok {
		t.Fatalf("the denial must name the permission, got %q", err.Error())
	}

	if perm.ID != permAuditProbePlain {
		t.Errorf("got permission %q, want %q", perm.ID, permAuditProbePlain)
	}
}

// TestAuditErrorCarriesTheIdNotTheName pins the representation itself, because that is the
// contract the presentation layer relies on.
func TestAuditErrorCarriesTheIdNotTheName(t *testing.T) {
	v := newAuditProbeView(t)

	err := v.Audit(permAuditProbeI18n)

	var denied PermissionDeniedError
	if !asPermissionDenied(err, &denied) {
		t.Fatalf("expected a PermissionDeniedError, got %T", err)
	}

	if string(denied) != string(permAuditProbeI18n) {
		t.Errorf("error payload is %q, want the plain id %q", string(denied), permAuditProbeI18n)
	}
}

func asPermissionDenied(err error, target *PermissionDeniedError) bool {
	d, ok := err.(PermissionDeniedError)
	if !ok {
		return false
	}

	*target = d

	return true
}

// TestDenialWithoutPermissionStaysAnonymous is the counterpart: refusals that are not about a
// permission, such as a tenant or ownership boundary, must keep reporting false so that the
// presentation layer falls back to a generic message instead of inventing a permission.
func TestDenialWithoutPermissionStaysAnonymous(t *testing.T) {
	for _, err := range []error{
		PermissionDeniedError("workspace forbids access"),
		PermissionDeniedError("workspace or share forbids access"),
		PermissionDeniedErr,
	} {
		if _, ok := permission.RequiredOf(err); ok {
			t.Errorf("%q must not be read as a permission", err.Error())
		}
	}
}

func auditProbeBundle(t *testing.T) i18n.Bundler {
	t.Helper()

	b, ok := i18n.Default.MatchBundle(language.German)
	if !ok {
		t.Fatal("no german bundle")
	}

	return b
}

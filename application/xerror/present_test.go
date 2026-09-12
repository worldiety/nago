// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerror

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xerrors"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

func TestPresentClassification(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantOk         bool
		wantKind       Kind
		wantValidation bool
	}{
		{name: "nil", err: nil, wantOk: false, wantKind: KindUnknown},
		{name: "unknown", err: errors.New("db connection reset"), wantOk: false, wantKind: KindUnknown},
		{
			name:     "permission denied sentinel",
			err:      user.PermissionDeniedErr,
			wantOk:   true,
			wantKind: KindDenied,
		},
		{
			name:     "permission denied with id",
			err:      user.PermissionDeniedError("nago.user.find_by_id"),
			wantOk:   true,
			wantKind: KindDenied,
		},
		{
			name:     "permission denied with free form reason",
			err:      user.PermissionDeniedError("workspace or share forbids access"),
			wantOk:   true,
			wantKind: KindDenied,
		},
		{
			name:     "invalid subject is not logged in",
			err:      user.InvalidSubjectErr,
			wantOk:   true,
			wantKind: KindNotLoggedIn,
		},
		{
			name:     "wrapped denial",
			err:      fmt.Errorf("cannot read file: %w", user.PermissionDeniedErr),
			wantOk:   true,
			wantKind: KindDenied,
		},
		{
			name:     "not exist",
			err:      os.ErrNotExist,
			wantOk:   true,
			wantKind: KindNotFound,
		},
		{
			name:     "already exists",
			err:      os.ErrExist,
			wantOk:   true,
			wantKind: KindAlreadyExists,
		},
		{
			name:     "localized error",
			err:      std.NewLocalizedError("Ungültiger Schlüssel", "Der Stream-Schlüssel darf nicht leer sein."),
			wantOk:   true,
			wantKind: KindLocalized,
		},
		{
			name:           "validation error",
			err:            xerrors.WithFields("invalid", "Name", "must not be empty"),
			wantOk:         true,
			wantKind:       KindValidation,
			wantValidation: true,
		},
		{
			name:           "validation joined with technical error",
			err:            errors.Join(xerrors.WithFields("invalid", "Name", "empty"), errors.New("db down")),
			wantOk:         true,
			wantKind:       KindValidation,
			wantValidation: true,
		},
		{
			// Fields are orthogonal to Kind: the denial wins the category, the field messages
			// still travel along so the client can correct the input.
			name:           "validation joined with a denial",
			err:            errors.Join(xerrors.WithFields("invalid", "Name", "empty"), user.PermissionDeniedErr),
			wantOk:         true,
			wantKind:       KindDenied,
			wantValidation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := Present(nil, tt.err)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}

			if p.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", p.Kind, tt.wantKind)
			}

			if p.Recognized() != tt.wantOk {
				t.Errorf("Recognized = %v, want %v", p.Recognized(), tt.wantOk)
			}

			if p.Validation() != tt.wantValidation {
				t.Errorf("Validation = %v, want %v", p.Validation(), tt.wantValidation)
			}

			if ok && (p.Title == "" || p.Message == "") {
				t.Errorf("recognized error must have a title and message, got %q / %q", p.Title, p.Message)
			}
		})
	}
}

// TestDeniedCoversBothAuthKinds pins the relationship that used to be maintained by hand via
// two booleans set together.
func TestDeniedCoversBothAuthKinds(t *testing.T) {
	for _, err := range []error{user.PermissionDeniedErr, user.InvalidSubjectErr} {
		p, _ := Present(nil, err)
		if !p.Denied() {
			t.Errorf("%v must count as denied, kind = %v", err, p.Kind)
		}
	}

	p, _ := Present(nil, os.ErrNotExist)
	if p.Denied() {
		t.Error("a missing element is not a denial")
	}
}

func TestKindString(t *testing.T) {
	// a Kind must never render as a bare integer in a log
	for k := KindUnknown; k <= KindLocalized; k++ {
		if s := k.String(); s == "" {
			t.Errorf("kind %d has no name", k)
		}
	}

	if got := Kind(99).String(); got != "unknown" {
		t.Errorf("out of range kind = %q", got)
	}
}

// TestNoQuestionMarkEverEscapes is the regression guard for the old
// alert/snackbar.go:160 behaviour, where every denial that was not exactly a
// user.PermissionDeniedError produced "a rights holder must grant ? first".
func TestNoQuestionMarkEverEscapes(t *testing.T) {
	denials := []error{
		user.PermissionDeniedErr,
		user.PermissionDeniedError("nago.user.find_by_id"),
		user.PermissionDeniedError("some.unregistered.permission"),
		user.PermissionDeniedError("workspace or share forbids access"),
		user.InvalidSubjectErr,
		session.NotLoggedInErr,
		fmt.Errorf("wrapped: %w", user.InvalidSubjectErr),
		tenantBoundary{},
	}

	for _, err := range denials {
		t.Run(fmt.Sprintf("%T/%v", err, err), func(t *testing.T) {
			p, ok := Present(nil, err)
			if !ok {
				t.Fatalf("denial must be recognized")
			}

			if strings.Contains(p.Message, "?") {
				t.Errorf("message must not contain a placeholder: %q", p.Message)
			}

			if !p.Denied() {
				t.Error("must be marked as denied")
			}
		})
	}
}

// tenantBoundary is a custom denial without any nameable permission, the case the other
// implementation could not express at all.
type tenantBoundary struct{}

func (tenantBoundary) Error() string          { return "tenant boundary crossed" }
func (tenantBoundary) PermissionDenied() bool { return true }

func TestCustomDenialIsRecognized(t *testing.T) {
	var err error = tenantBoundary{}

	if !permission.IsDenied(err) {
		t.Fatal("implementing PermissionDenied must be enough")
	}

	p, ok := Present(nil, err)
	if !ok || !p.Denied() {
		t.Fatal("want a recognized denial")
	}

	// must be the plain generic sentence, no grant instruction appended
	if strings.Contains(p.Message, "'") {
		t.Errorf("must not name a permission: %q", p.Message)
	}
}

func TestPresentNeverLeaksTechnicalDetail(t *testing.T) {
	// the message of a FieldBuilder is a technical dump and must never reach the user
	var errs xerrors.FieldBuilder
	errs.Add("Name", "must not be empty")
	err := fmt.Errorf("cannot decide command flow.CreateFormCmd: %w", errs.Error())

	p, ok := Present(nil, err)
	if !ok {
		t.Fatal("want recognized")
	}

	for _, leak := range []string{"flow.CreateFormCmd", "field validation failed", "map["} {
		if strings.Contains(p.Message, leak) || strings.Contains(p.Title, leak) {
			t.Errorf("message leaks %q: %q", leak, p.Message)
		}
	}

	if got := p.Fields["Name"]; got != "must not be empty" {
		t.Errorf("field message lost: %q", got)
	}
}

func TestPresentOrGeneric(t *testing.T) {
	p := PresentOrGeneric(nil, errors.New("some internal detail"))

	if p.Title == "" || p.Message == "" {
		t.Fatal("generic fallback must be populated")
	}

	if strings.Contains(p.Message, "some internal detail") {
		t.Error("generic fallback must not disclose the original message")
	}

	if p.Token() == "" || !strings.Contains(p.Message, p.Token()) {
		t.Error("generic fallback must carry the support token")
	}
}

func TestPresentOrGenericNil(t *testing.T) {
	if p := PresentOrGeneric(nil, nil); p.Title != "" || p.Cause() != nil {
		t.Error("nil error must produce the zero presentation")
	}
}

func TestTokenIsStable(t *testing.T) {
	err := errors.New("boom")
	a := PresentOrGeneric(nil, err).Token()
	b := PresentOrGeneric(nil, err).Token()

	if a != b || a == "" {
		t.Errorf("token must be stable and non empty, got %q and %q", a, b)
	}
}

// bundleFor returns the bundle of a concrete language so that the actual rendered text can be
// asserted. Every other test uses the default, which is why a nondeterministic fallback went
// unnoticed before.
func bundleFor(t *testing.T, tag language.Tag) i18n.Bundler {
	t.Helper()

	b, ok := i18n.Default.MatchBundle(tag)
	if !ok {
		t.Fatalf("no bundle for %v", tag)
	}

	return b
}

func TestPresentLocalizesContent(t *testing.T) {
	tests := []struct {
		tag         language.Tag
		wantTitle   string
		wantMessage string
	}{
		{language.German, "Zugriff verweigert", "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden."},
		{language.English, "Access denied", "You do not have permission to use this content or function."},
	}

	for _, tt := range tests {
		t.Run(tt.tag.String(), func(t *testing.T) {
			p, ok := Present(bundleFor(t, tt.tag), user.PermissionDeniedErr)
			if !ok {
				t.Fatal("want recognized")
			}

			if p.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", p.Title, tt.wantTitle)
			}

			if p.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", p.Message, tt.wantMessage)
			}
		})
	}
}

// TestDefaultBundlerIsDeterministic is the regression guard for a fallback that iterated a map
// and therefore answered in a random language per call.
func TestDefaultBundlerIsDeterministic(t *testing.T) {
	first := PresentOrGeneric(nil, user.PermissionDeniedErr).Message

	for i := 0; i < 50; i++ {
		if got := PresentOrGeneric(nil, user.PermissionDeniedErr).Message; got != first {
			t.Fatalf("iteration %d: %q != %q", i, got, first)
		}
	}

	if want := StrDeniedMsg.Get(bundleFor(t, FallbackLanguage)); first != want {
		t.Errorf("fallback = %q, want the %v text %q", first, FallbackLanguage, want)
	}
}

func TestPresentSetsRecognized(t *testing.T) {
	if p, _ := Present(nil, user.PermissionDeniedErr); !p.Recognized() {
		t.Error("a classified error must be marked recognized")
	}

	if p, _ := Present(nil, errors.New("boom")); p.Recognized() {
		t.Error("an unclassified error must not be marked recognized")
	}

	if p := PresentOrGeneric(nil, errors.New("boom")); p.Recognized() {
		t.Error("the generic fallback must not claim recognition")
	}
}

func TestPresentNamesRegisteredPermission(t *testing.T) {
	id := permission.Declare[func()]("nago.xerror.test_perm", "Testrecht", "nur für den Test")

	p, ok := Present(bundleFor(t, language.German), user.PermissionDeniedError(id))
	if !ok {
		t.Fatal("want recognized")
	}

	if !strings.Contains(p.Message, "Testrecht") {
		t.Errorf("a registered permission must be named: %q", p.Message)
	}

	if strings.Contains(p.Message, string(id)) {
		t.Errorf("the raw id must not appear in prose: %q", p.Message)
	}
}

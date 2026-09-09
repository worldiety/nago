// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink_test

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/application/permission"
	. "go.wdy.de/nago/application/speclink"
	"go.wdy.de/nago/auth"
)

// The catalogue is a process-wide global filled by package initialisation, so these declarations are the
// fixture of the whole file. Identities are prefixed so they cannot collide with a requirement any other
// package of this binary declares - spec.Declare panics on a duplicate, which would take the test binary
// down at start-up rather than fail a test.
var (
	reqPublic = spec.Declare(spec.Requirement{
		ID:     "R-TEST-PUBLIC",
		Kind:   spec.Functional,
		Status: spec.Normative,
		Title:  "Öffentliche Anforderung",
		Text:   "Diese Anforderung darf jeder sehen.",
	})

	_ = spec.Declare(spec.Requirement{
		ID:         "R-TEST-INTERNAL",
		Kind:       spec.Functional,
		Status:     spec.Normative,
		Disclosure: spec.Internal,
		Title:      "Interne Anforderung",
		Text:       "Diese Anforderung ist für die Organisation.",
	})

	_ = spec.Declare(spec.Requirement{
		ID:         "R-TEST-SECRET",
		Kind:       spec.Constraint,
		Status:     spec.Normative,
		Disclosure: spec.Secret,
		Title:      "Geheime Anforderung",
		Text:       "Diese Anforderung wird nie in einer Liste ausgeliefert.",
	})

	_ = spec.Declare(spec.Requirement{
		ID:           "R-TEST-DECISION",
		Kind:         spec.Decision,
		Status:       spec.Planned,
		Title:        "Geplante Entscheidung",
		Text:         "Etwas, das bewusst noch nicht umgesetzt ist.",
		Rationale:    "Weil es später billiger ist.",
		Consequences: "Bis dahin muss von Hand nachgearbeitet werden.",
	})
)

// A binding so there is a capability to read. It satisfies the public requirement and carries a help text.
type demoUseCase func()

var _ = spec.For[demoUseCase](
	spec.Satisfies(reqPublic),
	spec.Help("Tut etwas Vorführbares."),
)

// testSubject grants exactly the permissions it was given.
type testSubject struct {
	auth.Subject
	perms []permission.ID
}

func (s testSubject) Valid() bool { return true }

func (s testSubject) HasPermission(p permission.ID) bool { return slices.Contains(s.perms, p) }

func (s testSubject) Audit(p permission.ID) error {
	if s.HasPermission(p) {
		return nil
	}

	return errDenied
}

type deniedError struct{}

func (deniedError) Error() string { return "permission denied" }

var errDenied = deniedError{}

// reader may list and read, but not internal material.
func reader() auth.Subject {
	return testSubject{perms: []permission.ID{
		PermFindAllRequirements, PermFindRequirementByID, PermFindAllCapabilities, PermFindSourceDocument,
	}}
}

// insider may additionally read internal and confidential material.
func insider() auth.Subject {
	return testSubject{perms: []permission.ID{
		PermFindAllRequirements, PermFindRequirementByID, PermFindAllCapabilities, PermReadInternal,
	}}
}

func listIDs(t *testing.T, uc UseCases, subject auth.Subject, filter RequirementFilter) []ID {
	t.Helper()

	var got []ID
	for r, err := range uc.FindAllRequirements(subject, filter) {
		if err != nil {
			t.Fatalf("listing failed: %v", err)
		}
		got = append(got, r.ID)
	}

	slices.Sort(got)
	return got
}

// contains reports whether the listing holds the identity, ignoring whatever else the binary declared.
func contains(ids []ID, want ID) bool { return slices.Contains(ids, want) }

// TestListingHidesInternalMaterial is the control speclink deliberately does not provide. Its own
// documentation says the field "is not a control" and leaves enforcement to whoever shows the text to
// somebody - which is this package.
func TestListingHidesInternalMaterial(t *testing.T) {
	got := listIDs(t, NewUseCases(nil), reader(), RequirementFilter{})

	if !contains(got, "R-TEST-PUBLIC") {
		t.Error("a public requirement is not visible to an ordinary reader")
	}

	if contains(got, "R-TEST-INTERNAL") {
		t.Error("an internal requirement leaked into an ordinary listing")
	}
}

// TestListingShowsInternalMaterialToAnInsider is the other half; a permission that grants nothing would be
// worse than no permission.
func TestListingShowsInternalMaterialToAnInsider(t *testing.T) {
	got := listIDs(t, NewUseCases(nil), insider(), RequirementFilter{})

	if !contains(got, "R-TEST-INTERNAL") {
		t.Error("nago.speclink.requirement.read_internal grants nothing")
	}
}

// TestSecretIsNeverListed takes the classification at its word: "disclosed individually and never in bulk".
// A search that happened to match one would otherwise hand it out in exactly the way the class forbids.
func TestSecretIsNeverListed(t *testing.T) {
	for _, subject := range []auth.Subject{reader(), insider()} {
		if got := listIDs(t, NewUseCases(nil), subject, RequirementFilter{}); contains(got, "R-TEST-SECRET") {
			t.Errorf("a secret requirement appeared in a listing: %v", got)
		}
	}
}

// TestSecretIsReachableIndividually is the disclosure the class does allow, and only for somebody who may
// read internal material.
func TestSecretIsReachableIndividually(t *testing.T) {
	uc := NewUseCases(nil)

	opt, err := uc.FindRequirementByID(insider(), "R-TEST-SECRET")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	if opt.IsNone() {
		t.Error("an insider cannot read a secret requirement by identity")
	}

	opt, err = uc.FindRequirementByID(reader(), "R-TEST-SECRET")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	if opt.IsSome() {
		t.Error("an ordinary reader read a secret requirement")
	}
}

// TestHiddenRequirementReadsAsAbsent guards a subtle leak: reporting "refused" for a requirement the caller
// may not see confirms both that the identity exists and that it is classified above them, which is more
// than guessing an identity should reveal.
func TestHiddenRequirementReadsAsAbsent(t *testing.T) {
	uc := NewUseCases(nil)

	hidden, err := uc.FindRequirementByID(reader(), "R-TEST-INTERNAL")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	missing, err := uc.FindRequirementByID(reader(), "R-TEST-DOES-NOT-EXIST")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	if hidden.IsSome() || missing.IsSome() {
		t.Fatal("expected both to be absent")
	}
}

// TestPermissionsAreEnforced keeps the audit in place; without it the whole disclosure argument is moot.
func TestPermissionsAreEnforced(t *testing.T) {
	uc := NewUseCases(nil)
	nobody := testSubject{}

	for _, err := range uc.FindAllRequirements(nobody, RequirementFilter{}) {
		if err == nil {
			t.Error("listing requirements without a permission succeeded")
		}
		break
	}

	if _, err := uc.FindRequirementByID(nobody, "R-TEST-PUBLIC"); err == nil {
		t.Error("reading a requirement without a permission succeeded")
	}

	for _, err := range uc.FindAllCapabilities(nobody, CapabilityFilter{}) {
		if err == nil {
			t.Error("listing capabilities without a permission succeeded")
		}
		break
	}
}

// TestPlannedRequirementsAreVisible is the trap the speclink documentation names: the binding registry knows
// only requirements something is bound to, and a Planned one is deliberately bound to nothing. Reading the
// catalogue instead is what lets the assistant say "not implemented yet" rather than "no such rule".
func TestPlannedRequirementsAreVisible(t *testing.T) {
	got := listIDs(t, NewUseCases(nil), reader(), RequirementFilter{})

	if !contains(got, "R-TEST-DECISION") {
		t.Error("a planned requirement is missing; the assistant would report it as non-existent")
	}
}

// TestDecisionCarriesRationaleAndCost checks the two fields that make an explanation an account rather than
// advocacy.
func TestDecisionCarriesRationaleAndCost(t *testing.T) {
	opt, err := NewUseCases(nil).FindRequirementByID(reader(), "R-TEST-DECISION")
	if err != nil || opt.IsNone() {
		t.Fatalf("cannot read the decision: %v", err)
	}

	got := opt.Unwrap()

	if got.Rationale == "" {
		t.Error("the decision does not say why")
	}

	if got.Consequences == "" {
		t.Error("the decision does not say what it costs")
	}

	if got.Kind != "decision" || got.Status != "planned" {
		t.Errorf("got kind %q status %q, want decision/planned", got.Kind, got.Status)
	}
}

// TestFilterByKindAndQuery covers the narrowing a model relies on to keep an answer small.
func TestFilterByKindAndQuery(t *testing.T) {
	uc := NewUseCases(nil)

	if got := listIDs(t, uc, reader(), RequirementFilter{Kind: "decision"}); contains(got, "R-TEST-PUBLIC") {
		t.Errorf("the kind filter did not apply: %v", got)
	}

	if got := listIDs(t, uc, reader(), RequirementFilter{Query: "Öffentliche"}); !contains(got, "R-TEST-PUBLIC") {
		t.Errorf("the query filter did not find the requirement: %v", got)
	}
}

// TestUnknownFilterValueYieldsNothing rather than failing. A model that invents "mandatory" for a status
// should get an empty result and a chance to correct itself, not an error that ends the turn.
func TestUnknownFilterValueYieldsNothing(t *testing.T) {
	got := listIDs(t, NewUseCases(nil), reader(), RequirementFilter{Status: "mandatory"})

	if len(got) != 0 {
		t.Errorf("an unknown status matched %v", got)
	}
}

// TestCapabilitiesCarryHelpAndRequirement is the answer to "was kann das System", which somebody asks before
// they know which requirement to look up.
func TestCapabilitiesCarryHelpAndRequirement(t *testing.T) {
	for c, err := range NewUseCases(nil).FindAllCapabilities(reader(), CapabilityFilter{Query: "Vorführbares"}) {
		if err != nil {
			t.Fatalf("listing failed: %v", err)
		}

		if c.Help == "" {
			t.Error("the capability carries no help text")
		}

		if len(c.Requirements) == 0 || c.Requirements[0].ID != "R-TEST-PUBLIC" {
			t.Errorf("the capability does not name its requirement: %+v", c.Requirements)
		}

		if c.Requirements[0].Title == "" {
			t.Error("the requirement title was not resolved from the catalogue")
		}

		return
	}

	t.Fatal("the demo capability was not found")
}

// testSources stands in for the documents an application embeds under requirements/_sources.
var testSources = fstest.MapFS{
	"requirements/_sources/library.md": &fstest.MapFile{Data: []byte("# Was bestellt wurde\n\nDie Bibliothek soll ausleihen können.")},
}

// TestSourceDocumentIsReadable covers the answer that is often better than the requirement itself: the
// wording of whoever asked for the system, rather than the sentence somebody distilled from it.
func TestSourceDocumentIsReadable(t *testing.T) {
	uc := NewUseCases(testSources)

	// The path is taken verbatim from a requirement's sources field, anchor and all.
	got, err := uc.FindSourceDocument(reader(), "requirements/_sources/library.md#was-bestellt-wurde")
	if err != nil {
		t.Fatalf("cannot read the source: %v", err)
	}

	if !strings.Contains(got, "ausleihen") {
		t.Errorf("unexpected content: %q", got)
	}
}

// TestSourceDocumentRejectsEscapes guards against a path that leaves the embedded tree. The caller is
// frequently a model repeating a path back, so this is about a plausible mistake rather than an attacker.
func TestSourceDocumentRejectsEscapes(t *testing.T) {
	uc := NewUseCases(testSources)

	for _, bad := range []string{"../../etc/passwd", "/etc/passwd", "requirements/../../secret.md"} {
		if _, err := uc.FindSourceDocument(reader(), bad); err == nil {
			t.Errorf("path %q was accepted", bad)
		}
	}
}

// TestSourceDocumentWithoutAnyEmbedded says so plainly rather than being absent, so a caller does not have
// to test for a missing capability.
func TestSourceDocumentWithoutAnyEmbedded(t *testing.T) {
	if _, err := NewUseCases(nil).FindSourceDocument(reader(), "anything.md"); err == nil {
		t.Error("reading from an application without sources succeeded")
	}
}

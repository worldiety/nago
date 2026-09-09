// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package speclink makes the requirement catalogue of a running application readable through use cases.
//
// # What this is for
//
// An application can answer what it did. It usually cannot answer why it is allowed to do it, or why it
// refuses something a user expected to work. That answer lives in the requirements, and until now it lived
// only in the source tree - reachable by a developer with a checkout, and by nobody standing in front of the
// screen.
//
// speclink (https://github.com/worldiety/speclink) records requirements as ordinary Go values that are part
// of the normal build, and registers them in a runtime catalogue. This package turns that catalogue into
// nago use cases, so an AI assistant, a help page or a support view can reach it under the same
// authorization rules as everything else.
//
// # Why this belongs in the framework
//
// Because the alternative is what every project did before: a go/ast parser or a hand-maintained copy of the
// requirement texts, once per project, each one subtly different and each one drifting from the tree it
// claims to describe. spec.Declare exists precisely to end that, and a framework integration is what makes
// using it cheaper than reinventing it.
//
// # What it depends on
//
// Only github.com/worldiety/speclink/spec, which is a module of its own with no dependencies at all - not
// the speclink tool, which is a compiler frontend and has no business in an application's module graph.
package speclink

import (
	"github.com/worldiety/speclink/spec"
)

// ID is the stable identity of a requirement, e.g. "R-QUOTE-SUBMIT".
type ID string

// Requirement is one entry of the catalogue, flattened for reading.
//
// # Why this is not spec.Requirement
//
// Because spec.Requirement is recursive: DerivedFrom and Supersedes hold whole Requirement values, so
// serialising one embeds a large part of the derivation graph, repeatedly, at every level. That is
// unreadable for a person and ruinous for a model's context window.
//
// This view carries the identities instead and lets the reader follow them if it wants to. The `desc` tags
// are what let it be handed to a model directly; see the ai subpackage.
type Requirement struct {
	ID    ID     `json:"id" desc:"stabile Kennung, z. B. R-QUOTE-SUBMIT"`
	Title string `json:"title"`

	// Kind is functional, nonFunctional, constraint or decision.
	Kind string `json:"kind" desc:"functional, nonFunctional, constraint oder decision"`

	// Discipline is business, technical or mixed.
	Discipline string `json:"discipline" optional:"true" desc:"business, technical oder mixed"`

	// Status is normative, abstract, planned, outOfScope, informative or superseded.
	//
	// It is the field that separates "there is no such rule" from "it is not implemented yet", and an
	// assistant that ignores it will state the first when the second is true.
	Status string `json:"status" desc:"normative, abstract, planned, outOfScope, informative oder superseded. Nur normative Anforderungen müssen umgesetzt sein; planned ist bewusst noch offen."`

	// Text is the normative sentence, deliberately short.
	Text string `json:"text" desc:"die normative Aussage in einem Satz"`

	// Rationale is why a decision was taken that way. Present for decisions, usually empty otherwise.
	Rationale string `json:"rationale,omitempty" optional:"true" desc:"warum so entschieden wurde; bei Entscheidungen immer vorhanden"`

	// Consequences is what the ruling costs.
	//
	// speclink makes this mandatory for a decision and separate from the rationale, because it is the half
	// nobody writes unprompted - and without it a record reads as advocacy for the choice rather than as an
	// account of it.
	Consequences string `json:"consequences,omitempty" optional:"true" desc:"was die Entscheidung kostet; der Teil, der eine Begründung von einer Rechtfertigung unterscheidet"`

	// Disclosure says who this text may be shown to: public, internal, confidential or secret.
	Disclosure string `json:"disclosure" desc:"public, internal, confidential oder secret"`

	// DerivedFrom and Supersedes carry identities rather than values, see the type documentation.
	DerivedFrom []ID `json:"derivedFrom,omitempty" optional:"true" desc:"Kennungen der Anforderungen, aus denen diese abgeleitet ist"`
	Supersedes  []ID `json:"supersedes,omitempty" optional:"true" desc:"Kennungen der Anforderungen, die diese ablöst"`

	// Sources name where the requirement comes from: a document in the repository or an external norm.
	Sources []string `json:"sources,omitempty" optional:"true" desc:"Herkunft: Pfad#Anker eines Quelldokuments oder eine externe Norm wie „HGB §§ 383 ff.“"`

	// Topics are the themes this requirement belongs to.
	Topics []string `json:"topics,omitempty" optional:"true"`
}

// Identity makes the requirement an aggregate, which is what the data views expect.
func (r Requirement) Identity() ID { return r.ID }

// WithIdentity is required alongside Identity.
func (r Requirement) WithIdentity(id ID) Requirement {
	r.ID = id
	return r
}

// String is what a picker or a table cell displays.
func (r Requirement) String() string { return r.Title }

// Capability is one construct of the application together with what it is for.
//
// It is folded from the binding registry: the Help text written at the construct, the requirements it
// satisfies and any rationale recorded there. It answers "what can this system do, and on what grounds",
// which is the question somebody asks before they know which requirement to look up.
type Capability struct {
	// Construct is the fully qualified name of the annotated type, e.g. "example.com/erp/app/sales.SubmitQuote".
	Construct string `json:"construct" desc:"vollqualifizierter Name des Konstrukts im Code"`

	// Help is the end-user instruction recorded at the construct itself.
	Help string `json:"help,omitempty" optional:"true" desc:"Bedienhinweis, wie er am Konstrukt hinterlegt wurde"`

	// Rationale is a justification recorded at the construct.
	Rationale string `json:"rationale,omitempty" optional:"true" desc:"Begründung, die direkt am Konstrukt hinterlegt wurde"`

	// Requirements are the requirements this construct satisfies, with their titles resolved from the
	// catalogue so a reader does not have to look each identity up separately.
	Requirements []RequirementRef `json:"requirements,omitempty" optional:"true" desc:"die erfüllten Anforderungen"`
}

// RequirementRef names a requirement and repeats its title, so a capability listing is readable on its own.
type RequirementRef struct {
	ID    ID     `json:"id"`
	Title string `json:"title,omitempty" optional:"true"`
}

// RequirementFilter narrows a listing. Every field is optional: a filter narrows, it does not command.
type RequirementFilter struct {
	Kind   string `json:"kind" optional:"true" desc:"functional, nonFunctional, constraint oder decision"`
	Status string `json:"status" optional:"true" desc:"normative, abstract, planned, outOfScope, informative oder superseded"`
	Query  string `json:"query" optional:"true" desc:"Suchbegriff; wird in Kennung, Titel und Text gesucht, Groß- und Kleinschreibung egal"`
}

// CapabilityFilter narrows a capability listing.
type CapabilityFilter struct {
	Query string `json:"query" optional:"true" desc:"Suchbegriff; wird in Konstruktname, Bedienhinweis und Begründung gesucht"`

	// Requirement restricts to the capabilities satisfying one requirement, which is how "what implements
	// this rule" is answered.
	Requirement ID `json:"requirement" optional:"true" desc:"nur Fähigkeiten, die diese Anforderung erfüllen"`
}

// toRequirement flattens a catalogue entry into the readable view.
func toRequirement(r spec.Requirement) Requirement {
	out := Requirement{
		ID:           ID(r.ID),
		Title:        r.Title,
		Kind:         r.Kind.String(),
		Discipline:   r.Discipline.String(),
		Status:       r.Status.String(),
		Text:         r.Text,
		Rationale:    r.Rationale,
		Consequences: r.Consequences,
		Disclosure:   r.Disclosure.String(),
	}

	for _, d := range r.DerivedFrom {
		out.DerivedFrom = append(out.DerivedFrom, ID(d.ID))
	}

	for _, s := range r.Supersedes {
		out.Supersedes = append(out.Supersedes, ID(s.ID))
	}

	for _, s := range r.Sources {
		out.Sources = append(out.Sources, sourceLabel(s))
	}

	for _, t := range r.Topics {
		out.Topics = append(out.Topics, t.Title)
	}

	return out
}

// sourceLabel renders a source as one readable string. Exactly one of Doc and Extern is set.
func sourceLabel(s spec.Source) string {
	if s.Extern != "" {
		return s.Extern
	}

	if s.Anchor != "" {
		return s.Doc + "#" + s.Anchor
	}

	return s.Doc
}

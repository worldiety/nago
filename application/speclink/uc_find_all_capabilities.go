// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink

import (
	"iter"
	"strings"

	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/auth"
)

// FindAllCapabilities reads what this binary can do and on what grounds.
//
// It is folded from the binding registry rather than the catalogue: the Help text written at a construct,
// the requirements it satisfies, and any rationale recorded there. That is the answer to "was kann das
// System", which somebody asks before they know which requirement to look up.
type FindAllCapabilities func(subject auth.Subject, filter CapabilityFilter) iter.Seq2[Capability, error]

// NewFindAllCapabilities builds the listing over the binding registry.
func NewFindAllCapabilities() FindAllCapabilities {
	return func(subject auth.Subject, filter CapabilityFilter) iter.Seq2[Capability, error] {
		return func(yield func(Capability, error) bool) {
			if err := subject.Audit(PermFindAllCapabilities); err != nil {
				yield(Capability{}, err)
				return
			}

			// Titles are resolved from the catalogue so a capability reads on its own, and the same
			// disclosure ceiling applies: a requirement the subject may not read must not have its title
			// leak out through the thing that implements it.
			ceiling := ceilingFor(subject)
			titles := visibleTitles(ceiling)

			for _, e := range spec.Entries() {
				cap, ok := toCapability(e, titles)
				if !ok {
					continue
				}

				if !filter.matches(cap) {
					continue
				}

				if !yield(cap, nil) {
					return
				}
			}
		}
	}
}

// visibleTitles indexes the titles of every requirement the ceiling permits in a listing.
func visibleTitles(ceiling spec.Disclosure) map[ID]string {
	out := map[ID]string{}

	for _, r := range spec.Requirements() {
		if visibleInList(r, ceiling) {
			out[ID(r.ID)] = r.Title
		}
	}

	return out
}

// toCapability folds one registry entry into a capability, reporting false when there is nothing worth
// showing.
//
// # Why a bare Satisfies is not a capability
//
// Most bindings in a well annotated project say only that a construct serves a requirement, and very many
// of those sit on individual struct fields. That is traceability, and it is what the requirement tools are
// for. A capability is something the system *does*, described for a person - which is precisely what
// spec.Help is documented to carry: "the end user instruction for documentation, help system and assistant".
//
// Listing the rest would bury the dozen entries that answer "was kann das hier" under a few hundred that
// answer nothing, and a model asked for orientation would spend its context on field names.
func toCapability(e spec.Entry, titles map[ID]string) (Capability, bool) {
	cap := Capability{Construct: e.Target}

	for _, a := range e.Assertions {
		switch a.Kind {
		case "satisfies":
			for _, id := range a.Requirements {
				ref := RequirementRef{ID: ID(id)}

				title, visible := titles[ID(id)]
				if !visible {
					// The construct still satisfies it; the title is simply withheld. Dropping the reference
					// entirely would misreport the system as ungrounded.
					ref.Title = ""
				} else {
					ref.Title = title
				}

				cap.Requirements = append(cap.Requirements, ref)
			}
		case "help":
			cap.Help = a.Text
		case "rationale":
			cap.Rationale = a.Text
		}
	}

	if cap.Help == "" && cap.Rationale == "" {
		return Capability{}, false
	}

	// A binding on a declaration cannot name itself at run time - the identifier is gone by the time the
	// value arrives - so the construct is empty for those. Reporting a nameless capability helps nobody.
	if cap.Construct == "" {
		return Capability{}, false
	}

	return cap, true
}

// matches reports whether a capability falls into the filter.
func (f CapabilityFilter) matches(c Capability) bool {
	if f.Requirement != "" {
		found := false
		for _, r := range c.Requirements {
			if strings.EqualFold(string(r.ID), string(f.Requirement)) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	if q := strings.ToLower(strings.TrimSpace(f.Query)); q != "" {
		haystack := strings.ToLower(c.Construct + " " + c.Help + " " + c.Rationale)
		if !strings.Contains(haystack, q) {
			return false
		}
	}

	return true
}

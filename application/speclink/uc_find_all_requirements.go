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

// FindAllRequirements reads the requirement catalogue of this binary, narrowed by a filter and bounded by
// what the subject may be shown.
//
// # It is the catalogue of the binary, not of the repository
//
// Go initialises the package level variables of packages that are linked in, so a requirement package that
// no entry point imports is simply absent. For a program explaining itself that is arguably the right set -
// it should speak about what it is - but it is not the whole tree and must not be presented as one.
type FindAllRequirements func(subject auth.Subject, filter RequirementFilter) iter.Seq2[Requirement, error]

// NewFindAllRequirements builds the listing over the runtime catalogue.
func NewFindAllRequirements() FindAllRequirements {
	return func(subject auth.Subject, filter RequirementFilter) iter.Seq2[Requirement, error] {
		return func(yield func(Requirement, error) bool) {
			if err := subject.Audit(PermFindAllRequirements); err != nil {
				yield(Requirement{}, err)
				return
			}

			ceiling := ceilingFor(subject)

			for _, r := range spec.Requirements() {
				if !visibleInList(r, ceiling) {
					continue
				}

				if !filter.matches(r) {
					continue
				}

				if !yield(toRequirement(r), nil) {
					return
				}
			}
		}
	}
}

// matches reports whether a requirement falls into the filter.
//
// The enum fields are compared against their stable identifiers rather than parsed into the enum, so an
// unknown value simply matches nothing instead of failing the call. A caller - very often a model - that
// invents "mandatory" for a status should get an empty result and a chance to correct itself, not an error
// that ends the turn.
func (f RequirementFilter) matches(r spec.Requirement) bool {
	if k := strings.TrimSpace(f.Kind); k != "" && !strings.EqualFold(r.Kind.String(), k) {
		return false
	}

	if s := strings.TrimSpace(f.Status); s != "" && !strings.EqualFold(r.Status.String(), s) {
		return false
	}

	if q := strings.ToLower(strings.TrimSpace(f.Query)); q != "" {
		haystack := strings.ToLower(string(r.ID) + " " + r.Title + " " + r.Text + " " + r.Rationale)
		if !strings.Contains(haystack, q) {
			return false
		}
	}

	return true
}

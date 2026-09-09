// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink

import (
	"strings"

	"github.com/worldiety/option"
	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/auth"
)

// FindRequirementByID reads one requirement in full.
//
// This is the individual disclosure a Secret classification permits, and the only path by which such a
// requirement is ever handed out.
type FindRequirementByID func(subject auth.Subject, id ID) (option.Opt[Requirement], error)

// NewFindRequirementByID builds the lookup over the runtime catalogue.
func NewFindRequirementByID() FindRequirementByID {
	return func(subject auth.Subject, id ID) (option.Opt[Requirement], error) {
		if err := subject.Audit(PermFindRequirementByID); err != nil {
			return option.None[Requirement](), err
		}

		ceiling := ceilingFor(subject)

		for _, r := range spec.Requirements() {
			if !strings.EqualFold(string(r.ID), string(id)) {
				continue
			}

			// A requirement the subject may not see is reported as absent rather than as refused. The
			// alternative confirms that an identity exists and that it is classified above the reader, which
			// is more than a caller who guessed the identity should learn from asking.
			if !visibleIndividually(r, ceiling) {
				return option.None[Requirement](), nil
			}

			return option.Some(toRequirement(r)), nil
		}

		return option.None[Requirement](), nil
	}
}

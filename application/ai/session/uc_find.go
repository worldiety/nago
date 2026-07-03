// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"iter"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
)

// NewFindByID returns a [FindByID] use case. Access is resource-scoped: the subject must hold PermFindByID
// either globally or as an instance grant on this session. Otherwise the session is reported as non-existent
// (option.None) so its existence is not leaked.
func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[Session], error) {
		optSession, err := repo.FindByID(id)
		if err != nil {
			return option.None[Session](), fmt.Errorf("cannot find session by id: %w", err)
		}

		if optSession.IsNone() {
			return option.None[Session](), nil
		}

		if err := subject.AuditResource(Namespace, rebacInstance(id), PermFindByID); err != nil {
			return option.None[Session](), nil
		}

		return optSession, nil
	}
}

// NewFindAll returns a [FindAll] use case yielding the sessions the subject may see: those granted per
// instance (its own) plus all sessions if the subject holds PermFindAll globally. When [FindAllOptions.Tags]
// is set, only sessions carrying all of those tags are yielded. The (cheap) tag filter is applied before the
// (potentially more expensive) ReBAC audit.
func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject, opts FindAllOptions) iter.Seq2[Session, error] {
		return func(yield func(Session, error) bool) {
			for session, err := range repo.All() {
				if err != nil {
					if !yield(Session{}, err) {
						return
					}
					continue
				}

				if !hasAllTags(session.Tags, opts.Tags) {
					continue
				}

				if err := subject.AuditResource(Namespace, rebacInstance(session.ID), PermFindAll); err != nil {
					continue
				}

				if !yield(session, nil) {
					return
				}
			}
		}
	}
}

// hasAllTags reports whether have contains every tag in want (AND semantics). An empty want matches anything.
func hasAllTags(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	for _, w := range want {
		if !slices.Contains(have, w) {
			return false
		}
	}
	return true
}

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
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// NewFindByID returns a [FindByID] use case. Access is resource-scoped: the subject must hold PermFindByID
// either globally or as an instance grant on this session. Otherwise the session is reported as non-existent
// (option.None) so its existence is not leaked. A child session (see [Session.ParentID]) is also readable by
// whoever may read its parent: a sub-agent transcript belongs to the conversation, whoever continued it.
func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[Session], error) {
		optSession, err := repo.FindByID(id)
		if err != nil {
			return option.None[Session](), fmt.Errorf("cannot find session by id: %w", err)
		}

		if optSession.IsNone() {
			return option.None[Session](), nil
		}

		if !mayRead(subject, optSession.Unwrap(), PermFindByID) {
			return option.None[Session](), nil
		}

		return optSession, nil
	}
}

// mayRead reports whether the subject holds perm on the session or, for a child session, on its parent.
func mayRead(subject auth.Subject, s Session, perm permission.ID) bool {
	return mayReadID(subject, s.ID, s.ParentID, perm)
}

// mayReadID is mayRead for the session id with the parent.
func mayReadID(subject auth.Subject, id, parent ID, perm permission.ID) bool {
	if subject.AuditResource(Namespace, rebacInstance(id), perm) == nil {
		return true
	}

	return parent != "" && subject.AuditResource(Namespace, rebacInstance(parent), perm) == nil
}

// NewFindAll returns a [FindAll] use case yielding the sessions the subject may see: those granted per
// instance (its own) plus all sessions if the subject holds PermFindAll globally. When [FindAllOptions.Tags]
// is set, only sessions carrying all of those tags are yielded. Child sessions are skipped unless
// [FindAllOptions.IncludeChildren] is set. The (cheap) tag filter is applied before the
// (potentially more expensive) ReBAC audit. It reads every session.
func NewFindAll(repo Repository) FindAll {
	return newFindAll(repo, nil)
}

// newFindAll returns a [FindAll] use case, which reads the heads of h for the sessions of a creator without children,
// otherwise every session.
func newFindAll(repo Repository, h *heads) FindAll {
	return func(subject auth.Subject, opts FindAllOptions) iter.Seq2[Session, error] {
		return func(yield func(Session, error) bool) {
			if h != nil && opts.CreatedBy != "" && !opts.IncludeChildren {
				h.findAll(subject, opts, yield)
				return
			}

			scanAll(repo, subject, opts, yield)
		}
	}
}

// scanAll yields the sessions of [FindAll] by reading every session.
func scanAll(repo Repository, subject auth.Subject, opts FindAllOptions, yield func(Session, error) bool) {
	for session, err := range repo.All() {
		if err != nil {
			if !yield(Session{}, err) {
				return
			}
			continue
		}

		if findable(subject, session, opts) && !yield(session, nil) {
			return
		}
	}
}

// findable reports whether [FindAll] yields the session.
func findable(subject auth.Subject, s Session, opts FindAllOptions) bool {
	return summaryFindable(subject, Summary{ID: s.ID, ParentID: s.ParentID, CreatedBy: s.CreatedBy, Tags: s.Tags}, opts)
}

// summaryFindable reports whether [FindAll] yields the session of the head.
func summaryFindable(subject auth.Subject, s Summary, opts FindAllOptions) bool {
	if s.ParentID != "" && !opts.IncludeChildren {
		return false
	}

	if opts.CreatedBy != "" && s.CreatedBy != opts.CreatedBy {
		return false
	}

	if !hasAllTags(s.Tags, opts.Tags) {
		return false
	}

	return mayReadID(subject, s.ID, s.ParentID, PermFindAll)
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

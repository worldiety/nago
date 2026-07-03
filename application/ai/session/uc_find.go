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
// instance (its own) plus all sessions if the subject holds PermFindAll globally.
func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[Session, error] {
		return func(yield func(Session, error) bool) {
			for session, err := range repo.All() {
				if err != nil {
					if !yield(Session{}, err) {
						return
					}
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

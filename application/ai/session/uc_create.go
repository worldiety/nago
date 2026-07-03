// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

// NewCreate returns a [Create] use case that persists a new session and grants its creator owner access.
//
// Creation is a global capability: the audit uses an empty instance, which succeeds when the subject holds
// PermCreate globally (e.g. via an IAM group). After persisting, the creator is granted the per-instance
// permissions (see grantOwner) so they - and only they, unless granted globally - can find, continue, rename
// and delete this specific session.
//
// No lock is needed: the session id is freshly generated and collision-free, so there is no concurrent
// read-modify-write on the same instance.
func NewCreate(repo Repository, rdb *rebac.DB) Create {
	return func(subject auth.Subject, opts CreateOptions) (Session, error) {
		if err := subject.AuditResource(Namespace, "", PermCreate); err != nil {
			return Session{}, err
		}

		now := xtime.Now()

		var messages []completion.Message
		if len(opts.Input) > 0 {
			messages = []completion.Message{
				{Role: completion.User, Content: opts.Input},
			}
		}

		session := Session{
			ID:           data.RandIdent[ID](),
			Title:        opts.Title,
			Model:        opts.Model,
			System:       opts.System,
			ProviderHint: opts.ProviderHint,
			Tags:         opts.Tags,
			Messages:     messages,
			CreatedAt:    now,
			CreatedBy:    subject.ID(),
			UpdatedAt:    now,
		}

		if err := repo.Save(session); err != nil {
			return Session{}, fmt.Errorf("cannot persist new session: %w", err)
		}

		// Grant the creator ownership of this instance. Roll back the just-saved session if the grant fails,
		// so we never leave a session nobody (but a global admin) can access.
		if err := grantOwner(rdb, subject.ID(), session.ID); err != nil {
			if delErr := repo.DeleteByID(session.ID); delErr != nil {
				return Session{}, fmt.Errorf("cannot grant session ownership: %w; cannot roll back session: %v", err, delErr)
			}
			return Session{}, fmt.Errorf("cannot grant session ownership: %w", err)
		}

		return session, nil
	}
}

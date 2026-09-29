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
	"log/slog"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

// NewDelete returns a [Delete] use case. Deleting a non-existent session is a no-op (idempotent). A session
// the subject may not access (no PermDelete globally nor as an instance grant) is treated as non-existent, so
// it is neither deleted nor revealed. On success the session's ReBAC grants are revoked as well. Serializes
// with other mutations of the same session via its keyed lock.
//
// Child sessions (see [Session.ParentID]) are deleted with their parent, and the background tasks of the session
// in tasks (which may be nil) are cancelled.
func NewDelete(locks *locker, repo Repository, rdb *rebac.DB, tasks *completion.TaskRegistry) Delete {
	var deleteFn Delete
	deleteFn = func(subject auth.Subject, id ID) error {
		deleted, err := deleteOne(locks, repo, rdb, subject, id)
		if err != nil || !deleted {
			return err
		}

		// Only after the audit passed: otherwise anybody could stop the tasks of a foreign session.
		if tasks != nil {
			tasks.Cancel(string(id))
		}

		// Children are hidden from every listing, so nobody could ever remove them otherwise. Collect first:
		// deleting while iterating the repository is not safe.
		var children []ID
		for s, err := range repo.All() {
			if err != nil {
				return fmt.Errorf("cannot list child sessions: %w", err)
			}
			if s.ParentID == id {
				children = append(children, s.ID)
			}
		}

		for _, child := range children {
			if err := deleteFn(subject, child); err != nil {
				return err
			}
		}

		return nil
	}

	return deleteFn
}

// deleteOne removes a single session under its lock and reports whether it did, see [NewDelete].
func deleteOne(locks *locker, repo Repository, rdb *rebac.DB, subject auth.Subject, id ID) (bool, error) {
	defer locks.lock(id)()

	optSession, err := repo.FindByID(id)
	if err != nil {
		return false, fmt.Errorf("cannot load session: %w", err)
	}

	if optSession.IsNone() {
		return false, nil
	}

	if err := subject.AuditResource(Namespace, rebacInstance(id), PermDelete); err != nil {
		// Silently ignore inaccessible sessions instead of deleting or leaking their existence.
		return false, nil
	}

	if err := repo.DeleteByID(id); err != nil {
		return false, fmt.Errorf("cannot delete session: %w", err)
	}

	// Best-effort cleanup of the instance grants. A failure here does not fail the delete (the session is
	// already gone); the dangling triples are harmless as they reference a non-existent instance.
	if err := revokeInstance(rdb, id); err != nil {
		slog.Error("cannot revoke rebac grants for deleted session", "session", id, "err", err)
	}

	return true, nil
}

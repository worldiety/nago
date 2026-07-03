// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

// NewDelete returns a [Delete] use case. Deleting a non-existent session is a no-op (idempotent). A session
// the subject may not access (no PermDelete globally nor as an instance grant) is treated as non-existent, so
// it is neither deleted nor revealed. On success the session's ReBAC grants are revoked as well. Serializes
// with other mutations of the same session via its keyed lock.
func NewDelete(locks *locker, repo Repository, rdb *rebac.DB) Delete {
	return func(subject auth.Subject, id ID) error {
		defer locks.lock(id)()

		optSession, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot load session: %w", err)
		}

		if optSession.IsNone() {
			return nil
		}

		if err := subject.AuditResource(Namespace, rebacInstance(id), PermDelete); err != nil {
			// Silently ignore inaccessible sessions instead of deleting or leaking their existence.
			return nil
		}

		if err := repo.DeleteByID(id); err != nil {
			return fmt.Errorf("cannot delete session: %w", err)
		}

		// Best-effort cleanup of the instance grants. A failure here does not fail the delete (the session is
		// already gone); the dangling triples are harmless as they reference a non-existent instance.
		if err := revokeInstance(rdb, id); err != nil {
			slog.Error("cannot revoke rebac grants for deleted session", "session", id, "err", err)
		}

		return nil
	}
}

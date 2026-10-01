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

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

// NewDelete returns a [Delete] use case. Deleting a non-existent session is a no-op (idempotent). A session
// the subject may not access (no PermDelete globally nor as an instance grant) is treated as non-existent, so
// it is neither deleted nor revealed. On success the session's ReBAC grants are revoked as well. Serializes
// with other mutations of the same session via its keyed lock.
//
// Child sessions (see [Session.ParentID]) are deleted with their parent, whoever created them, and the
// background tasks of the session in tasks (which may be nil) are cancelled first, so that a running run of
// the session ends soon and no task creates another child meanwhile.
func NewDelete(locks *locker, repo Repository, rdb *rebac.DB, tasks *completion.TaskRegistry, ledger *usageLedger) Delete {
	return func(subject auth.Subject, id ID) error {
		// The audit happens before anything else, otherwise anybody could stop the tasks of a foreign session.
		// It does not need the lock: a run of the session holds that until its tasks are done, and those are
		// cancelled right below.
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

		if tasks != nil {
			tasks.Cancel(string(id))
		}

		if err := deleteUnchecked(locks, repo, rdb, ledger, id); err != nil {
			return err
		}

		// Children are hidden from every listing, so nobody could ever remove them otherwise. They are deleted
		// without an audit of their own: the parent has been audited, and a child may belong to whoever
		// continued the parent. A task which passed the checks of its creation right before the parent was
		// gone may still save a child, thus the search is repeated once.
		for range 2 {
			children, err := childrenOf(repo, id)
			if err != nil {
				return err
			}

			if len(children) == 0 {
				return nil
			}

			for _, child := range children {
				if tasks != nil {
					tasks.Cancel(string(child))
				}
				if err := deleteUnchecked(locks, repo, rdb, ledger, child); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

// childrenOf returns the ids of the child sessions of id. Collecting first is required, because deleting while
// iterating the repository is not safe.
func childrenOf(repo Repository, id ID) ([]ID, error) {
	var children []ID
	for s, err := range repo.All() {
		if err != nil {
			return nil, fmt.Errorf("cannot list child sessions: %w", err)
		}
		if s.ParentID == id {
			children = append(children, s.ID)
		}
	}

	return children, nil
}

// deleteUnchecked removes a single session under its lock, see [NewDelete]. The caller has audited the access.
func deleteUnchecked(locks *locker, repo Repository, rdb *rebac.DB, ledger *usageLedger, id ID) error {
	release, err := locks.lock(nil, id)
	if err != nil {
		return err
	}
	defer release()

	optSession, err := repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("cannot load session: %w", err)
	}

	if optSession.IsNone() {
		return nil
	}

	if err := repo.DeleteByID(id); err != nil {
		return fmt.Errorf("cannot delete session: %w", err)
	}

	// the usage of sub-agents nobody will book any more
	ledger.drain(id)

	// Best-effort cleanup of the instance grants. A failure here does not fail the delete (the session is
	// already gone); the dangling triples are harmless as they reference a non-existent instance.
	if err := revokeInstance(rdb, id); err != nil {
		slog.Error("cannot revoke rebac grants for deleted session", "session", id, "err", err)
	}

	return nil
}

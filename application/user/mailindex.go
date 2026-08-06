// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/data"
)

// NormalizeEmail trims and lowercases the given address, so that mail addresses can be compared and indexed
// case-insensitive, which is consistent with [Email.Equals].
func NormalizeEmail(e Email) Email {
	return Email(strings.ToLower(strings.TrimSpace(string(e))))
}

// MailIndex is an inverse index from the normalized mail address to the user id, so that mail based lookups
// are O(1) instead of a full repository scan.
//
// The index is populated lazily on first use and afterwards kept in sync using the saved and deleted observers
// of the given repository. Therefore, it is only consistent as long as all writers use the same
// [data.NotifyRepository] instance, which is guaranteed by [NewUseCases]. Bypassing the repository, e.g. by
// writing into the underlying blob store directly, will make this index go stale.
type MailIndex struct {
	repo    data.NotifyRepository[User, ID]
	once    sync.Once
	initErr error

	mutex  sync.RWMutex
	byMail map[Email]ID
	byID   map[ID]Email
}

// NewMailIndex creates the index and immediately attaches itself to the given repository. Thus, it must be
// created before any writing use case is created, otherwise updates may be lost.
func NewMailIndex(repo data.NotifyRepository[User, ID]) *MailIndex {
	idx := &MailIndex{
		repo:   repo,
		byMail: make(map[Email]ID),
		byID:   make(map[ID]Email),
	}

	// note: we never detach, because the index lives as long as the use cases do.
	repo.AddSavedObserver(func(_ data.Repository[User, ID], saved data.Saved[User, ID]) error {
		idx.put(saved.ID, saved.Value.Email)
		return nil
	})

	repo.AddDeletedObserver(func(_ data.Repository[User, ID], deleted data.Deleted[ID]) error {
		idx.remove(deleted.ID)
		return nil
	})

	return idx
}

// Lookup returns the user id which is associated with the given mail address.
func (idx *MailIndex) Lookup(mail Email) (ID, bool, error) {
	if err := idx.ensure(); err != nil {
		return "", false, err
	}

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	id, ok := idx.byMail[NormalizeEmail(mail)]

	return id, ok, nil
}

// Used returns true, if any user has the given mail address assigned.
func (idx *MailIndex) Used(mail Email) (bool, error) {
	_, ok, err := idx.Lookup(mail)
	return ok, err
}

// UsedByOther returns true, if a user other than the given one has the mail address assigned. This is the
// uniqueness check for the login identity.
func (idx *MailIndex) UsedByOther(mail Email, except ID) (bool, error) {
	id, ok, err := idx.Lookup(mail)
	if err != nil || !ok {
		return false, err
	}

	return id != except, nil
}

func (idx *MailIndex) ensure() error {
	idx.once.Do(func() {
		// hold the write lock, so that concurrent observers are applied strictly before or after the
		// initial scan. Applying an observed value twice is idempotent.
		idx.mutex.Lock()
		defer idx.mutex.Unlock()

		for usr, err := range idx.repo.All() {
			if err != nil {
				idx.initErr = fmt.Errorf("cannot build user mail index: %w", err)
				return
			}

			idx.putLocked(usr.ID, usr.Email)
		}
	})

	return idx.initErr
}

func (idx *MailIndex) put(id ID, mail Email) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.putLocked(id, mail)
}

func (idx *MailIndex) putLocked(id ID, mail Email) {
	// evict a potentially outdated key, otherwise a renamed address would stay resolvable
	if oldMail, ok := idx.byID[id]; ok {
		delete(idx.byMail, oldMail)
	}

	newMail := NormalizeEmail(mail)
	if other, ok := idx.byMail[newMail]; ok && other != id {
		// this must never happen, because the mail is the unique login identity. We do not fail here,
		// because that would break otherwise working lookups, but an admin must be able to see it.
		slog.Error("unique mail violation in user repository", "mail", newMail, "user", id, "other", other)
	}

	idx.byMail[newMail] = id
	idx.byID[id] = newMail
}

func (idx *MailIndex) remove(id ID) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	mail, ok := idx.byID[id]
	if !ok {
		return
	}

	delete(idx.byID, id)

	// only drop the key, if it still points to us, otherwise we would remove a foreign entry
	if other, ok := idx.byMail[mail]; ok && other == id {
		delete(idx.byMail, mail)
	}
}

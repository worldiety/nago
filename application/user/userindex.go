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

// UserIndex is an inverse index from the identifying properties of a user, namely the normalized mail address
// and the subject id of the identity provider, to the user id. Thus, these lookups are O(1) instead of a full
// repository scan.
//
// The index is populated lazily on first use and afterwards kept in sync using the saved and deleted observers
// of the given repository. Therefore, it is only consistent as long as all writers use the same
// [data.NotifyRepository] instance, which is guaranteed by [NewUseCases]. Bypassing the repository, e.g. by
// writing into the underlying blob store directly, will make this index go stale.
type UserIndex struct {
	repo    data.NotifyRepository[User, ID]
	once    sync.Once
	initErr error

	mutex       sync.RWMutex
	byMail      map[Email]ID
	byNLSUserID map[NLSUserID]ID
	byID        map[ID]userIndexKeys
}

// userIndexKeys keeps the last indexed keys of a user, which is required to evict outdated entries when a
// user is updated or deleted.
type userIndexKeys struct {
	mail      Email
	nlsUserID NLSUserID
}

// NewUserIndex creates the index and immediately attaches itself to the given repository. Thus, it must be
// created before any writing use case is created, otherwise updates may be lost.
func NewUserIndex(repo data.NotifyRepository[User, ID]) *UserIndex {
	idx := &UserIndex{
		repo:        repo,
		byMail:      make(map[Email]ID),
		byNLSUserID: make(map[NLSUserID]ID),
		byID:        make(map[ID]userIndexKeys),
	}

	// note: we never detach, because the index lives as long as the use cases do.
	repo.AddSavedObserver(func(_ data.Repository[User, ID], saved data.Saved[User, ID]) error {
		idx.put(saved.Value)
		return nil
	})

	repo.AddDeletedObserver(func(_ data.Repository[User, ID], deleted data.Deleted[ID]) error {
		idx.remove(deleted.ID)
		return nil
	})

	return idx
}

// LookupMail returns the user id which is associated with the given mail address.
func (idx *UserIndex) LookupMail(mail Email) (ID, bool, error) {
	if err := idx.ensure(); err != nil {
		return "", false, err
	}

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	id, ok := idx.byMail[NormalizeEmail(mail)]

	return id, ok, nil
}

// LookupNLSUserID returns the user id which is associated with the given subject id of the identity provider.
// An empty nlsUserID never matches, because that just means that a user has not been merged yet.
func (idx *UserIndex) LookupNLSUserID(nlsUserID NLSUserID) (ID, bool, error) {
	if nlsUserID == "" {
		return "", false, nil
	}

	if err := idx.ensure(); err != nil {
		return "", false, err
	}

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	id, ok := idx.byNLSUserID[nlsUserID]

	return id, ok, nil
}

// Used returns true, if any user has the given mail address assigned.
func (idx *UserIndex) Used(mail Email) (bool, error) {
	_, ok, err := idx.LookupMail(mail)
	return ok, err
}

// UsedByOther returns true, if a user other than the given one has the mail address assigned. This is the
// uniqueness check for the login identity.
func (idx *UserIndex) UsedByOther(mail Email, except ID) (bool, error) {
	id, ok, err := idx.LookupMail(mail)
	if err != nil || !ok {
		return false, err
	}

	return id != except, nil
}

func (idx *UserIndex) ensure() error {
	idx.once.Do(func() {
		// hold the write lock, so that concurrent observers are applied strictly before or after the
		// initial scan. Applying an observed value twice is idempotent.
		idx.mutex.Lock()
		defer idx.mutex.Unlock()

		for usr, err := range idx.repo.All() {
			if err != nil {
				idx.initErr = fmt.Errorf("cannot build user index: %w", err)
				return
			}

			idx.putLocked(usr)
		}
	})

	return idx.initErr
}

func (idx *UserIndex) put(usr User) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.putLocked(usr)
}

func (idx *UserIndex) putLocked(usr User) {
	// evict potentially outdated keys, otherwise a renamed address would stay resolvable
	idx.removeKeysLocked(usr.ID)

	keys := userIndexKeys{
		mail:      NormalizeEmail(usr.Email),
		nlsUserID: usr.NLSUserID,
	}

	if other, ok := idx.byMail[keys.mail]; ok && other != usr.ID {
		// this must never happen, because the mail is the unique login identity. We do not fail here,
		// because that would break otherwise working lookups, but an admin must be able to see it.
		slog.Error("unique mail violation in user repository", "mail", keys.mail, "user", usr.ID, "other", other)
	}

	idx.byMail[keys.mail] = usr.ID

	if keys.nlsUserID != "" {
		if other, ok := idx.byNLSUserID[keys.nlsUserID]; ok && other != usr.ID {
			slog.Error("unique nls user id violation in user repository", "nlsUserId", keys.nlsUserID, "user", usr.ID, "other", other)
		}

		idx.byNLSUserID[keys.nlsUserID] = usr.ID
	}

	idx.byID[usr.ID] = keys
}

func (idx *UserIndex) remove(id ID) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.removeKeysLocked(id)
	delete(idx.byID, id)
}

// removeKeysLocked drops all keys which currently point to the given user. Keys which have been taken over by
// another user are kept, otherwise we would remove a foreign entry.
func (idx *UserIndex) removeKeysLocked(id ID) {
	keys, ok := idx.byID[id]
	if !ok {
		return
	}

	if other, ok := idx.byMail[keys.mail]; ok && other == id {
		delete(idx.byMail, keys.mail)
	}

	if keys.nlsUserID != "" {
		if other, ok := idx.byNLSUserID[keys.nlsUserID]; ok && other == id {
			delete(idx.byNLSUserID, keys.nlsUserID)
		}
	}
}

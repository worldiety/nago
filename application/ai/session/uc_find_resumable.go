// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// FindResumableOptions names the chat context of [FindResumable].
type FindResumableOptions struct {
	// Tags of the chat. A session continues it only with exactly the same set of tags.
	Tags []string

	// ProviderHint of the chat, see [Session.ProviderHint].
	ProviderHint string
}

// FindResumable returns the session a freshly opened chat of the subject continues: the newest session which waits
// on a user decision, has no parent, has been created by the subject and belongs to exactly the given chat context,
// see [Session.ResumableBy]. The subject must be allowed to read it like through [FindAll]. It returns none for a
// subject which is not signed in.
//
// The result equals reading every session through [FindAll], but FindResumable reads an index of the waiting
// sessions instead, which [NewUseCases] keeps for its repository, and loads only what it finds there.
type FindResumable func(subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error)

// ResumableBy reports whether a chat of the given account in the given context continues the session: the session
// waits on a user decision, has no parent, has been created by the account, has the same provider hint and exactly
// the same tags. Child sessions hold the transcripts of sub-agents, which nobody continues, and sessions of other
// contexts are never picked up; a history still offers them.
func (s Session) ResumableBy(account user.ID, opts FindResumableOptions) bool {
	return resumable(s.Pending != nil, s.ParentID, s.CreatedBy, s.ProviderHint, s.Tags, account, opts)
}

// resumable is the rule of [Session.ResumableBy], which the index applies to its entries as well.
func resumable(pending bool, parent ID, createdBy user.ID, providerHint string, tags []string, account user.ID, opts FindResumableOptions) bool {
	return pending && parent == "" && createdBy == account && providerHint == opts.ProviderHint && sameTagSet(tags, opts.Tags)
}

// sameTagSet reports whether have has as many tags as want and contains each of them.
func sameTagSet(have, want []string) bool {
	if len(have) != len(want) {
		return false
	}

	for _, t := range want {
		if !slices.Contains(have, t) {
			return false
		}
	}

	return true
}

// newFindResumable returns a [FindResumable] which narrows its search with the index. The persisted session
// decides, the index only names the candidates.
func newFindResumable(repo Repository, index *pendingIndex) FindResumable {
	return func(subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error) {
		if subject == nil || !subject.Valid() {
			return option.None[Session](), nil
		}

		candidates, err := index.candidates(repo, subject.ID(), opts)
		if err != nil {
			return option.None[Session](), err
		}

		for _, id := range candidates {
			optSession, err := repo.FindByID(id)
			if err != nil {
				return option.None[Session](), fmt.Errorf("cannot load session: %w", err)
			}

			if optSession.IsNone() {
				// deleted after the index named it
				continue
			}

			s := optSession.Unwrap()
			if !s.ResumableBy(subject.ID(), opts) || !mayRead(subject, s, PermFindAll) {
				continue
			}

			return optSession, nil
		}

		return option.None[Session](), nil
	}
}

// pendingIndex knows the sessions of one repository which wait on a user decision and have no parent, the only
// ones [FindResumable] may return, with what [Session.ResumableBy] needs and the time of their last change.
//
// It is built at the first lookup by reading every session once. From then on, indexedRepository keeps it in step
// with every write of the use cases: a write and the update of the index happen under mu, thus the index follows
// the order in which the sessions have been persisted, whoever writes, and a write never meets a half-built index.
// A failed write, or a deletion of which it cannot tell what it deleted, discards the index, so that the next
// lookup reads everything again instead of trusting a guess. Writes which bypass the use cases, e.g. a restored
// backup or another process on the same store, are not seen.
type pendingIndex struct {
	mu      sync.RWMutex
	loaded  bool
	entries map[ID]pendingEntry
}

// pendingEntry is what the index keeps of a waiting session.
type pendingEntry struct {
	createdBy    user.ID
	providerHint string
	tags         []string
	updatedAt    xtime.UnixMilliseconds
}

// candidates returns the ids of the indexed sessions which the account continues in the context of opts, the
// newest first and among equally new ones the smallest id first, which is the one a scan in the order of the
// repository keeps. The first call reads every session of repo.
func (x *pendingIndex) candidates(repo Repository, account user.ID, opts FindResumableOptions) ([]ID, error) {
	if err := x.load(repo); err != nil {
		return nil, err
	}

	x.mu.RLock()
	defer x.mu.RUnlock()

	type candidate struct {
		id        ID
		updatedAt xtime.UnixMilliseconds
	}

	var found []candidate
	for id, e := range x.entries {
		if resumable(true, "", e.createdBy, e.providerHint, e.tags, account, opts) {
			found = append(found, candidate{id: id, updatedAt: e.updatedAt})
		}
	}

	slices.SortFunc(found, func(a, b candidate) int {
		if c := cmp.Compare(b.updatedAt, a.updatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.id, b.id)
	})

	ids := make([]ID, 0, len(found))
	for _, c := range found {
		ids = append(ids, c.id)
	}

	return ids, nil
}

// load builds the index by reading every session of repo, unless it is built already. A failed read leaves it
// unbuilt, so that the next lookup tries again.
func (x *pendingIndex) load(repo Repository) error {
	x.mu.RLock()
	loaded := x.loaded
	x.mu.RUnlock()
	if loaded {
		return nil
	}

	x.mu.Lock()
	defer x.mu.Unlock()
	if x.loaded {
		return nil
	}

	entries := map[ID]pendingEntry{}
	for s, err := range repo.All() {
		if err != nil {
			return fmt.Errorf("cannot read the sessions for the index of waiting sessions: %w", err)
		}

		if e, ok := entryOf(s); ok {
			entries[s.ID] = e
		}
	}

	x.entries = entries
	x.loaded = true
	return nil
}

// entryOf returns the entry of a waiting session without parent.
func entryOf(s Session) (pendingEntry, bool) {
	if s.Pending == nil || s.ParentID != "" {
		return pendingEntry{}, false
	}

	return pendingEntry{
		createdBy:    s.CreatedBy,
		providerHint: s.ProviderHint,
		tags:         slices.Clone(s.Tags),
		updatedAt:    s.UpdatedAt,
	}, true
}

// saved takes over a persisted session. The caller holds mu.
func (x *pendingIndex) saved(s Session) {
	if !x.loaded {
		return
	}

	if e, ok := entryOf(s); ok {
		x.entries[s.ID] = e
		return
	}

	delete(x.entries, s.ID)
}

// deleted forgets a deleted session. The caller holds mu.
func (x *pendingIndex) deleted(id ID) {
	if !x.loaded {
		return
	}

	delete(x.entries, id)
}

// discard drops the index, so that the next lookup reads every session again. The caller holds mu.
func (x *pendingIndex) discard() {
	x.loaded = false
	x.entries = nil
}

// indexedRepository is the repository of the use cases, which keeps their pendingIndex in step with every write,
// see there. Reads go to the wrapped repository unchanged.
type indexedRepository struct {
	Repository
	index *pendingIndex
}

func (r indexedRepository) Save(s Session) error {
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	if err := r.Repository.Save(s); err != nil {
		// whether it has been written is unknown
		r.index.discard()
		return err
	}

	r.index.saved(s)
	return nil
}

func (r indexedRepository) SaveAll(it iter.Seq[Session]) error {
	// collected first, so that no foreign code runs under the lock
	all := slices.Collect(it)

	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	if err := r.Repository.SaveAll(slices.Values(all)); err != nil {
		r.index.discard()
		return err
	}

	for _, s := range all {
		r.index.saved(s)
	}

	return nil
}

func (r indexedRepository) DeleteByID(id ID) error {
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	if err := r.Repository.DeleteByID(id); err != nil {
		r.index.discard()
		return err
	}

	r.index.deleted(id)
	return nil
}

func (r indexedRepository) DeleteByEntity(s Session) error {
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	if err := r.Repository.DeleteByEntity(s); err != nil {
		r.index.discard()
		return err
	}

	r.index.deleted(s.Identity())
	return nil
}

func (r indexedRepository) DeleteAllByID(ids iter.Seq[ID]) error {
	all := slices.Collect(ids)

	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	if err := r.Repository.DeleteAllByID(slices.Values(all)); err != nil {
		r.index.discard()
		return err
	}

	for _, id := range all {
		r.index.deleted(id)
	}

	return nil
}

func (r indexedRepository) DeleteAll() error {
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	// what is left after a failure is unknown, and after a success nothing is left: reading again is right in both
	// cases and costs nothing for an empty repository
	defer r.index.discard()

	return r.Repository.DeleteAll()
}

func (r indexedRepository) Delete(predicate func(Session) (bool, error)) error {
	// The predicate is foreign code and must not run under the lock. The index is discarded afterwards, because it
	// cannot tell what has been deleted; a build which ran meanwhile has finished when the lock is acquired, thus
	// it is discarded as well.
	err := r.Repository.Delete(predicate)

	r.index.mu.Lock()
	defer r.index.mu.Unlock()
	r.index.discard()

	return err
}

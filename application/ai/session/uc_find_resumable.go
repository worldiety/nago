// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
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
// see [Session.ResumableBy], and among equally new ones the first in the order of the ids. The subject must be allowed
// to read it like through [FindAll]. It returns none for a subject which is not signed in.
//
// The result equals reading every session through [FindAll], leaving out those which cannot be decoded. With [Heads],
// FindResumable reads the heads of the waiting sessions of the subject and loads only the session it returns.
type FindResumable func(subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error)

// ResumableBy reports whether a chat of the given account in the given context continues the session: the session
// waits on a user decision, has no parent, has been created by the account, has the same provider hint and exactly
// the same tags. Child sessions hold the transcripts of sub-agents, which nobody continues, and sessions of other
// contexts are never picked up; a history still offers them.
func (s Session) ResumableBy(account user.ID, opts FindResumableOptions) bool {
	return resumable(s.Pending != nil, s.ParentID, s.CreatedBy, s.ProviderHint, s.Tags, account, opts)
}

// resumable is the rule of [Session.ResumableBy], which applies to the heads as well.
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

// newFindResumable returns a [FindResumable], which reads the heads of h, or every session without them.
func newFindResumable(repo Repository, h *heads) FindResumable {
	return func(subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error) {
		if subject == nil || !subject.Valid() {
			return option.None[Session](), nil
		}

		if h != nil {
			return h.findResumable(subject, opts)
		}

		return scanResumable(repo, subject, opts)
	}
}

// scanResumable finds the session of [FindResumable] by reading every session. A session, which cannot be decoded,
// is left out.
func scanResumable(repo Repository, subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error) {
	var best *Session
	for s, err := range repo.All() {
		if err != nil {
			if undecodable(err) {
				continue
			}
			return option.None[Session](), fmt.Errorf("cannot read the sessions: %w", err)
		}

		if !s.ResumableBy(subject.ID(), opts) || !mayRead(subject, s, PermFindAll) {
			continue
		}

		// the first among equally new ones in the order of the repository
		if best == nil || s.UpdatedAt > best.UpdatedAt {
			best = &s
		}
	}

	if best == nil {
		return option.None[Session](), nil
	}

	return option.Some(*best), nil
}

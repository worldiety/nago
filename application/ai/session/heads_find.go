// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"cmp"
	"iter"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// The searches with heads. Each reads the marks first, then the heads, which the coverage covers and which have no
// mark, then the sessions with a mark and the sessions the coverage does not cover. Without any coverage, it reads
// every session like the use cases without heads.

// findResumable is [FindResumable] with heads: it reads the heads of the waiting sessions of the subject and the
// newest session among them, which the subject continues.
func (h *heads) findResumable(subject auth.Subject, opts FindResumableOptions) (option.Opt[Session], error) {
	c, err := h.coverage()
	if err != nil {
		return option.None[Session](), err
	}

	if c.none() {
		return scanResumable(h.repo, subject, opts)
	}

	marked, err := h.marked()
	if err != nil {
		return option.None[Session](), err
	}

	type candidate struct {
		id        ID
		updatedAt xtime.UnixMilliseconds
		loaded    *Session
	}

	var found []candidate
	ids, err := h.candidates(c, marked, keyHeads+string(subject.ID())+"/w/")
	if err != nil {
		return option.None[Session](), err
	}

	for _, id := range ids {
		var sum Summary
		ok, err := h.get(keyHeads+string(subject.ID())+"/w/"+string(id), &sum)
		if err != nil {
			return option.None[Session](), err
		}

		if ok && resumable(sum.Pending, sum.ParentID, sum.CreatedBy, sum.ProviderHint, sum.Tags, subject.ID(), opts) {
			found = append(found, candidate{id: id, updatedAt: sum.UpdatedAt})
		}
	}

	consider := func(s Session) {
		if s.ResumableBy(subject.ID(), opts) {
			found = append(found, candidate{id: s.ID, updatedAt: s.UpdatedAt, loaded: &s})
		}
	}

	for id := range marked {
		if !c.covers(id) {
			continue
		}

		s, ok, err := h.load(id)
		if err != nil {
			if undecodable(err) {
				continue
			}
			return option.None[Session](), err
		}

		if ok {
			consider(s)
		}
	}

	for s, err := range h.uncovered(c) {
		if err != nil {
			if undecodable(err) {
				continue
			}
			return option.None[Session](), err
		}

		consider(s)
	}

	// the newest first, among equally new ones the first in the order of the repository
	slices.SortFunc(found, func(a, b candidate) int {
		if n := cmp.Compare(b.updatedAt, a.updatedAt); n != 0 {
			return n
		}
		return cmp.Compare(a.id, b.id)
	})

	for _, cand := range found {
		s := cand.loaded
		if s == nil {
			loaded, ok, err := h.load(cand.id)
			if err != nil {
				if undecodable(err) {
					continue
				}
				return option.None[Session](), err
			}

			if !ok {
				// a head without its session does not count
				continue
			}

			s = &loaded
		}

		if s.ResumableBy(subject.ID(), opts) && mayRead(subject, *s, PermFindAll) {
			return option.Some(*s), nil
		}
	}

	return option.None[Session](), nil
}

// findAll is [FindAll] with heads for the sessions of a creator without children. It reads the heads of the creator
// and the sessions they name. A session, which cannot be decoded, is yielded as an error only, if the heads name it
// for the creator.
func (h *heads) findAll(subject auth.Subject, opts FindAllOptions, yield func(Session, error) bool) {
	c, err := h.coverage()
	if err != nil {
		yield(Session{}, err)
		return
	}

	if c.none() {
		scanAll(h.repo, subject, opts, yield)
		return
	}

	marked, err := h.marked()
	if err != nil {
		yield(Session{}, err)
		return
	}

	ids, err := h.candidates(c, marked, keyHeads+string(opts.CreatedBy)+"/")
	if err != nil {
		yield(Session{}, err)
		return
	}

	for id := range marked {
		if c.covers(id) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)

	for _, id := range ids {
		s, ok, err := h.load(id)
		if err != nil {
			// only a session the heads name for the creator is known to be one of the result
			if (marked[id] && undecodable(err)) || yield(Session{}, err) {
				continue
			}
			return
		}

		if ok && findable(subject, s, opts) && !yield(s, nil) {
			return
		}
	}

	for s, err := range h.uncovered(c) {
		if err != nil {
			if undecodable(err) || yield(Session{}, err) {
				continue
			}
			return
		}

		if findable(subject, s, opts) && !yield(s, nil) {
			return
		}
	}
}

// findSummaries is [FindSummaries] with heads. It reads the heads of the creator, or all of them, and the children
// only with [FindAllOptions.IncludeChildren].
func (h *heads) findSummaries(subject auth.Subject, opts FindAllOptions) iter.Seq2[Summary, error] {
	return func(yield func(Summary, error) bool) {
		c, err := h.coverage()
		if err != nil {
			yield(Summary{}, err)
			return
		}

		var res []Summary
		add := func(s Session, err error) bool {
			if err != nil {
				// a session, which cannot be decoded, is left out, otherwise it would break every list
				return undecodable(err) || yield(Summary{}, err)
			}

			if findable(subject, s, opts) {
				res = append(res, s.Summary())
			}

			return true
		}

		if c.none() {
			for s, err := range h.uncovered(c) {
				if !add(s, err) {
					return
				}
			}

			sortSummaries(res, opts)
			yieldSummaries(res, yield)
			return
		}

		marked, err := h.marked()
		if err != nil {
			yield(Summary{}, err)
			return
		}

		prefixes := []string{keyHeads + string(opts.CreatedBy) + "/"}
		if opts.CreatedBy == "" {
			prefixes = []string{keyHeads}
		}
		if opts.IncludeChildren {
			prefixes = append(prefixes, keyChildren)
		}

		// a session, whose state changes, has both heads for a moment: the newer one counts
		seen := map[ID]int{}
		for _, prefix := range prefixes {
			keys, err := h.keys(prefix)
			if err != nil {
				yield(Summary{}, err)
				return
			}

			for _, key := range keys {
				id := idOfKey(key)
				if !c.covers(id) || marked[id] {
					continue
				}

				var sum Summary
				ok, err := h.get(key, &sum)
				if err != nil {
					if !yield(Summary{}, err) {
						return
					}
					continue
				}

				if !ok || !summaryFindable(subject, sum, opts) {
					continue
				}

				exists, err := h.contentExists(id)
				if err != nil {
					if !yield(Summary{}, err) {
						return
					}
					continue
				}

				if !exists {
					// a head without its session does not count
					continue
				}

				if i, dup := seen[id]; dup {
					if sum.UpdatedAt > res[i].UpdatedAt {
						res[i] = sum
					}
					continue
				}

				seen[id] = len(res)
				res = append(res, sum)
			}
		}

		for id := range marked {
			if !c.covers(id) {
				continue
			}

			s, ok, err := h.load(id)
			if err != nil || ok {
				if !add(s, err) {
					return
				}
			}
		}

		for s, err := range h.uncovered(c) {
			if !add(s, err) {
				return
			}
		}

		sortSummaries(res, opts)
		yieldSummaries(res, yield)
	}
}

func yieldSummaries(res []Summary, yield func(Summary, error) bool) {
	for _, s := range res {
		if !yield(s, nil) {
			return
		}
	}
}

// children returns the children of the session. It reads their heads. A child, which cannot be decoded, is left out.
func (h *heads) children(parent ID) ([]ID, error) {
	c, err := h.coverage()
	if err != nil {
		return nil, err
	}

	if c.none() {
		return scanChildren(h.repo, parent)
	}

	marked, err := h.marked()
	if err != nil {
		return nil, err
	}

	ids, err := h.candidates(c, marked, keyChildren+string(parent)+"/")
	if err != nil {
		return nil, err
	}

	consider := func(s Session, err error) error {
		switch {
		case err != nil && undecodable(err):
			return nil
		case err != nil:
			return err
		case s.ParentID == parent:
			ids = append(ids, s.ID)
		}

		return nil
	}

	for id := range marked {
		if !c.covers(id) {
			continue
		}

		s, ok, err := h.load(id)
		if err != nil || ok {
			if err := consider(s, err); err != nil {
				return nil, err
			}
		}
	}

	for s, err := range h.uncovered(c) {
		if err := consider(s, err); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

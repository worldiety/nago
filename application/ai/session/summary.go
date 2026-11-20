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

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// Summary is the head of a session: what a list shows and a search needs, without the conversation. Its JSON names
// are the ones of [Session].
type Summary struct {
	ID           ID                     `json:"id,omitempty"`
	ParentID     ID                     `json:"parentId,omitempty"`
	CreatedBy    user.ID                `json:"createdBy,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Preview      string                 `json:"preview,omitempty"` // see [Session.Preview]
	ProviderHint string                 `json:"providerHint,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Pending      bool                   `json:"pending,omitempty"`     // see [Session.Pending]
	TurnLimited  bool                   `json:"turnLimited,omitempty"` // see [Session.TurnLimited]
	Messages     int                    `json:"messages,omitempty"`    // the number of [Session.Messages]
	CreatedAt    xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	UpdatedAt    xtime.UnixMilliseconds `json:"updatedAt,omitempty"`
}

// Summary returns the head of the session.
func (s Session) Summary() Summary {
	return Summary{
		ID:           s.ID,
		ParentID:     s.ParentID,
		CreatedBy:    s.CreatedBy,
		Title:        s.Title,
		Preview:      s.Preview(),
		ProviderHint: s.ProviderHint,
		Tags:         slices.Clone(s.Tags),
		Pending:      s.Pending != nil,
		TurnLimited:  s.TurnLimited,
		Messages:     len(s.Messages),
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// FindSummaries yields the heads of the sessions [FindAll] yields with the same options, without reading a
// conversation: with [FindAllOptions.CreatedBy] in the order of the ids, otherwise by creator and id. A session, which
// cannot be decoded, is left out. Only use cases without [Heads] read every session for it.
type FindSummaries func(subject auth.Subject, opts FindAllOptions) iter.Seq2[Summary, error]

// newFindSummaries returns a [FindSummaries] use case, which reads the heads of h, or every session without them.
func newFindSummaries(repo Repository, h *heads) FindSummaries {
	findAll := newFindAll(repo, h)
	return func(subject auth.Subject, opts FindAllOptions) iter.Seq2[Summary, error] {
		if h != nil {
			return h.findSummaries(subject, opts)
		}

		return func(yield func(Summary, error) bool) {
			var res []Summary
			for s, err := range findAll(subject, opts) {
				if err != nil {
					if undecodable(err) {
						continue
					}
					if !yield(Summary{}, err) {
						return
					}
					continue
				}

				res = append(res, s.Summary())
			}

			sortSummaries(res, opts)
			for _, s := range res {
				if !yield(s, nil) {
					return
				}
			}
		}
	}
}

// sortSummaries sorts as [FindSummaries] yields: with CreatedBy by id, otherwise by creator and id.
func sortSummaries(res []Summary, opts FindAllOptions) {
	slices.SortFunc(res, func(a, b Summary) int {
		if opts.CreatedBy == "" {
			if c := cmp.Compare(a.CreatedBy, b.CreatedBy); c != 0 {
				return c
			}
		}

		return cmp.Compare(a.ID, b.ID)
	})
}

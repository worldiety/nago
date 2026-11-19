// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"iter"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
)

// countingRepo counts the reads of every session of the store.
type countingRepo struct {
	session.Repository
	full atomic.Int32
}

func (r *countingRepo) All() iter.Seq2[session.Session, error] {
	r.full.Add(1)
	return r.Repository.All()
}

func (r *countingRepo) Identifiers() iter.Seq2[session.ID, error] {
	r.full.Add(1)
	return r.Repository.Identifiers()
}

func (r *countingRepo) FindAllByPrefix(prefix session.ID) iter.Seq2[session.Session, error] {
	r.full.Add(1)
	return r.Repository.FindAllByPrefix(prefix)
}

// TestFindResumable_ChatReadsEverySessionOnce proves the wokoda decision R-DEC-ASSISTANT-RESUME-INDEX for the lookup
// of a freshly built chat (findResumable): with the use cases of session.NewUseCases it reads the stored sessions once,
// at its first call, and then never again, also after a write; it still finds the own waiting session of the context,
// never the one of another person. Counter row: use cases assembled without FindResumable read every session on every
// call (the fallback over FindAll, see TestFindResumable_OnlyOwnPendingSessionOfThisContext), so the count sees it.
func TestFindResumable_ChatReadsEverySessionOnce(t *testing.T) {
	su := user.SU()
	pending := &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q"}}}}
	stored := []session.Session{
		{ID: "mine", CreatedBy: su.ID(), ProviderHint: "p", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: xtime.UnixMilliseconds(1)},
		{ID: "foreign", CreatedBy: "someone-else", ProviderHint: "p", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: xtime.UnixMilliseconds(2)},
	}

	for _, c := range []struct {
		name  string
		index bool
		reads int32
	}{
		{"through the index", true, 1},
		{"counter row: without FindResumable", false, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := &countingRepo{Repository: session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](mem.NewBlobStore(string(session.Namespace))))}
			for _, s := range stored {
				if err := repo.Repository.Save(s); err != nil {
					t.Fatal(err)
				}
			}
			rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
			if err != nil {
				t.Fatal(err)
			}
			uc := session.NewUseCases(repo, rdb)
			if !c.index {
				uc.FindResumable = nil
			}

			for i := range 3 {
				if got := findResumable(su, uc, []string{"ctx"}, "p"); got == nil || got.ID != "mine" {
					t.Fatalf("lookup %d: %+v, expected the own waiting session", i+1, got)
				}
				if i == 1 {
					if err := uc.Rename(su, "mine", "renamed"); err != nil {
						t.Fatal(err)
					}
				}
			}
			if got := findResumable(su, uc, nil, "p"); got != nil {
				t.Fatalf("an untagged chat continues %s", got.ID)
			}
			if n := repo.full.Load(); n != c.reads {
				t.Errorf("%d reads of every session for four lookups, expected %d", n, c.reads)
			}
		})
	}
}

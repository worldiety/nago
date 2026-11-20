// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"context"
	"iter"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
)

// readLog records the conversations the use cases read from the store.
type readLog struct {
	session.Repository
	mu    sync.Mutex
	reads []session.ID
}

func (r *readLog) note(id session.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads = append(r.reads, id)
}

func (r *readLog) take() []session.ID {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := r.reads
	r.reads = nil
	return res
}

func (r *readLog) FindByID(id session.ID) (option.Opt[session.Session], error) {
	o, err := r.Repository.FindByID(id)
	if err == nil && o.IsSome() {
		r.note(id)
	}
	return o, err
}

func (r *readLog) noted(seq iter.Seq2[session.Session, error]) iter.Seq2[session.Session, error] {
	return func(yield func(session.Session, error) bool) {
		for s, err := range seq {
			if err == nil {
				r.note(s.ID)
			}
			if !yield(s, err) {
				return
			}
		}
	}
}

func (r *readLog) All() iter.Seq2[session.Session, error] { return r.noted(r.Repository.All()) }

func (r *readLog) FindAllByPrefix(p session.ID) iter.Seq2[session.Session, error] {
	return r.noted(r.Repository.FindAllByPrefix(p))
}

func (r *readLog) FindAllByID(ids iter.Seq[session.ID]) iter.Seq2[session.Session, error] {
	return r.noted(r.Repository.FindAllByID(ids))
}

// TestFindResumable_ChatReadsNoForeignSession proves the wokoda decision R-DEC-ASSISTANT-SESSION-HEADS for the lookup
// of a freshly built chat (findResumable), replacing the test of the former index of the waiting sessions, which read
// every session once: with the use cases of session.NewUseCases with session.Heads it reads no conversation of
// another person, on any call, also after a write; it still finds the own waiting session of the context, never the
// one of another person. Counter row: the same use cases without Heads read the conversation of the other person, so
// that the probe sees such a read.
func TestFindResumable_ChatReadsNoForeignSession(t *testing.T) {
	su := user.SU()
	pending := &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q"}}}}
	stored := []session.Session{
		{ID: "mine", CreatedBy: su.ID(), ProviderHint: "p", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: xtime.UnixMilliseconds(1)},
		{ID: "foreign", CreatedBy: "someone-else", ProviderHint: "p", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: xtime.UnixMilliseconds(2)},
	}

	for _, c := range []struct {
		name    string
		heads   bool
		foreign bool
	}{
		{"with Heads", true, false},
		{"counter row: without Heads", false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := &readLog{Repository: session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](mem.NewBlobStore("nago.ai.session.chatreads")))}
			for _, s := range stored {
				if err := repo.Repository.Save(s); err != nil {
					t.Fatal(err)
				}
			}
			rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
			if err != nil {
				t.Fatal(err)
			}
			var opts []session.Option
			if c.heads {
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				opts = append(opts, session.Heads(ctx, mem.NewBlobStore("nago.ai.session.head.chatreads")))
			}
			uc := session.NewUseCases(repo, rdb, opts...)
			if c.heads {
				// the store comes from a former version: wait until its heads are built, i.e. a lookup reads nothing
				// of another person any more
				deadline := time.Now().Add(10 * time.Second)
				for {
					repo.take()
					findResumable(su, uc, []string{"ctx"}, "p")
					if !slices.Contains(repo.take(), "foreign") {
						break
					}
					if time.Now().After(deadline) {
						t.Fatalf("the heads of the stored sessions are never built")
					}
					time.Sleep(time.Millisecond)
				}
			}

			foreign := 0
			for i := range 3 {
				repo.take()
				if got := findResumable(su, uc, []string{"ctx"}, "p"); got == nil || got.ID != "mine" {
					t.Fatalf("lookup %d: %+v, expected the own waiting session", i+1, got)
				}
				if slices.Contains(repo.take(), "foreign") {
					foreign++
				}
				if i == 1 {
					if err := uc.Rename(su, "mine", "renamed"); err != nil {
						t.Fatal(err)
					}
				}
			}
			if (foreign > 0) != c.foreign {
				t.Errorf("%d of 3 lookups read the conversation of another person, expected some: %v", foreign, c.foreign)
			}
			if got := findResumable(su, uc, nil, "p"); got != nil {
				t.Fatalf("an untagged chat continues %s", got.ID)
			}
		})
	}
}

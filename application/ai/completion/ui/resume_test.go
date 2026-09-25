// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"iter"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

func sessionsOf(all ...session.Session) session.UseCases {
	return session.UseCases{FindAll: func(auth.Subject, session.FindAllOptions) iter.Seq2[session.Session, error] {
		return func(yield func(session.Session, error) bool) {
			for _, s := range all {
				if !yield(s, nil) {
					return
				}
			}
		}
	}}
}

func TestFindResumable_OnlyOwnPendingSessionOfThisContext(t *testing.T) {
	su := user.SU()
	pending := &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q"}}}}
	mine := func(id session.ID, updated int64, tags ...string) session.Session {
		return session.Session{ID: id, CreatedBy: su.ID(), ProviderHint: "p", Tags: tags, Pending: pending, UpdatedAt: xtime.UnixMilliseconds(updated)}
	}

	uc := sessionsOf(
		mine("old", 1, "ctx"),
		mine("new", 2, "ctx"),
		mine("other-context", 3, "ctx", "more"),
		session.Session{ID: "foreign", CreatedBy: "someone-else", ProviderHint: "p", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: 4},
		session.Session{ID: "answered", CreatedBy: su.ID(), ProviderHint: "p", Tags: []string{"ctx"}, UpdatedAt: 5},
		session.Session{ID: "other-provider", CreatedBy: su.ID(), ProviderHint: "x", Tags: []string{"ctx"}, Pending: pending, UpdatedAt: 6},
	)

	got := findResumable(su, uc, []string{"ctx"}, "p")
	if got == nil || got.ID != "new" {
		t.Fatalf("expected the newest own pending session of this context, got %+v", got)
	}

	if got := findResumable(su, sessionsOf(mine("tagged", 1, "ctx")), nil, "p"); got != nil {
		t.Fatalf("an untagged chat must not resume a tagged session, got %s", got.ID)
	}
}

func TestOptimisticAnswers_OnlyAnsweredQuestions(t *testing.T) {
	cont := &completion.Continuation{Pending: []completion.PendingCall{
		{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q"}},
		{Kind: completion.PendingApproval, Call: completion.ToolCall{ID: "w"}},
	}}
	msg := optimisticAnswers(cont, []completion.Resolution{{CallID: "q", Answer: "red"}, {CallID: "w", Approved: true}})
	if len(msg.Content) != 1 {
		t.Fatalf("expected only the answer, got %+v", msg.Content)
	}
	if a, ok := askAnswer(msg.Content[0].(completion.ToolResult)); !ok || a != "red" {
		t.Fatalf("answer must render like a real ask_user result, got %q", a)
	}
}

func TestSuspensionKey_DiffersPerSuspension(t *testing.T) {
	a := &completion.Continuation{Pending: []completion.PendingCall{{Call: completion.ToolCall{ID: "1"}}}}
	b := &completion.Continuation{Pending: []completion.PendingCall{{Call: completion.ToolCall{ID: "2"}}}}
	if suspensionKey(a, 0) == suspensionKey(b, 0) {
		t.Fatalf("different suspensions of a transient chat must not share decision state")
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	stdjson "encoding/json"
	"errors"
	"iter"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// scriptedCompletions answers with the queued results in order and counts the calls.
type scriptedCompletions struct {
	results []completion.Result
	calls   int
	fail    error
}

func (s *scriptedCompletions) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (s *scriptedCompletions) Complete(_ auth.Subject, _ completion.Options) (completion.Result, error) {
	if s.fail != nil {
		return completion.Result{}, s.fail
	}
	r := s.results[min(s.calls, len(s.results)-1)]
	s.calls++
	return r, nil
}

func (s *scriptedCompletions) Stream(auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func askResult() completion.Result {
	return completion.Result{
		Message: completion.Message{Role: completion.Assistant, Content: []completion.Content{
			completion.ToolCall{ID: "q1", Name: completion.AskUserToolName, Arguments: stdjson.RawMessage(`{"question":"which?"}`)},
		}},
		StopReason: completion.StopToolUse,
	}
}

func answerResult(text string) completion.Result {
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: text}}},
		StopReason: completion.StopEndTurn,
	}
}

func newPendingSession(t *testing.T) (UseCases, Repository, Session, *scriptedCompletions) {
	t.Helper()
	uc, repo, _ := newTestUseCases(t)
	subject := user.SU()

	s, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &scriptedCompletions{results: []completion.Result{askResult(), answerResult("done")}}
	s, err = uc.Append(subject, s.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Tools:       []completion.Tool{completion.NewAskUserTool()},
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	return uc, repo, s, fake
}

func TestAppend_QuestionPersistsAndReleases(t *testing.T) {
	uc, repo, s, fake := newPendingSession(t)

	if s.Pending == nil || s.PendingRevision != 1 || s.Pending.Pending[0].Question != "which?" {
		t.Fatalf("expected a persisted pending question, got %+v", s.Pending)
	}
	if fake.calls != 1 {
		t.Fatalf("model must not be asked again while waiting, calls=%d", fake.calls)
	}

	stored, _ := repo.FindByID(s.ID)
	if stored.Unwrap().Pending == nil {
		t.Fatalf("pending question not stored")
	}

	// A new question while waiting must not break the pending exchange.
	if _, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: []completion.Content{completion.Text{Text: "x"}}}); !errors.Is(err, ErrPendingDecision) {
		t.Fatalf("expected ErrPendingDecision, got %v", err)
	}
}

// TestResolve_AfterRestart resumes with fresh use cases on the same store, as after a server restart; nothing
// in memory survives but the persisted session.
func TestResolve_AfterRestart(t *testing.T) {
	_, repo, s, _ := newPendingSession(t)
	restarted := NewUseCases(repo, newTestRDB(t))
	fake := &scriptedCompletions{results: []completion.Result{answerResult("thanks")}}

	resolved, err := restarted.Resolve(user.SU(), s.ID, ResolveOptions{
		Run:         AppendOptions{Completions: fake, Tools: []completion.Tool{completion.NewAskUserTool()}},
		Revision:    s.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "the red one"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Pending != nil {
		t.Fatalf("pending must be cleared")
	}
	// user, ask tool_use, answer tool_result, final
	if len(resolved.Messages) != 4 || firstText(t, resolved.Messages[3]) != "thanks" {
		t.Fatalf("unexpected history %+v", resolved.Messages)
	}
}

func TestResolve_RejectsStaleRevision(t *testing.T) {
	uc, _, s, fake := newPendingSession(t)
	opts := ResolveOptions{
		Run:         AppendOptions{Completions: fake, Tools: []completion.Tool{completion.NewAskUserTool()}},
		Revision:    s.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "a"}},
	}

	if _, err := uc.Resolve(user.SU(), s.ID, opts); err != nil {
		t.Fatal(err)
	}
	// the same answer again, e.g. from a second tab
	if _, err := uc.Resolve(user.SU(), s.ID, opts); !errors.Is(err, ErrNoPendingDecision) {
		t.Fatalf("expected ErrNoPendingDecision, got %v", err)
	}
}

func TestResolve_RejectsModelSwitch(t *testing.T) {
	uc, _, s, fake := newPendingSession(t)
	_, err := uc.Resolve(user.SU(), s.ID, ResolveOptions{
		Run:         AppendOptions{Completions: fake, Model: "other-model"},
		Revision:    s.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "a"}},
	})
	if err == nil {
		t.Fatalf("expected an error when switching models while pending")
	}
}

func TestResolve_FailureKeepsAnswer(t *testing.T) {
	uc, repo, s, _ := newPendingSession(t)
	failing := &scriptedCompletions{fail: errors.New("provider down")}

	_, err := uc.Resolve(user.SU(), s.ID, ResolveOptions{
		Run:         AppendOptions{Completions: failing, Tools: []completion.Tool{completion.NewAskUserTool()}},
		Revision:    s.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "a"}},
	})
	if err == nil {
		t.Fatal("expected the provider error")
	}

	stored := func() Session { o, _ := repo.FindByID(s.ID); return o.Unwrap() }()
	if stored.Pending != nil || len(stored.Messages) != 3 {
		t.Fatalf("the accepted answer must be persisted even though the provider failed, got %+v", stored)
	}
}

func TestDismiss_ClosesQuestionWithoutModel(t *testing.T) {
	uc, _, s, fake := newPendingSession(t)

	dismissed, err := uc.Dismiss(user.SU(), s.ID, s.PendingRevision)
	if err != nil {
		t.Fatal(err)
	}
	if dismissed.Pending != nil || len(dismissed.Messages) != 3 || fake.calls != 1 {
		t.Fatalf("unexpected dismiss result %+v (calls=%d)", dismissed, fake.calls)
	}

	// a new topic may follow
	fake.results = []completion.Result{answerResult("new topic")}
	fake.calls = 0
	if _, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: []completion.Content{completion.Text{Text: "other"}}}); err != nil {
		t.Fatal(err)
	}
}

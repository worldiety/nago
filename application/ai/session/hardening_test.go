// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"iter"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// rebacSubject is a user whose resource permissions come from the ReBAC database, like a real one.
type rebacSubject struct {
	auth.Subject
	id      user.ID
	rdb     *rebac.DB
	globals map[permission.ID]bool
}

func newRebacSubject(id user.ID, rdb *rebac.DB, globals ...permission.ID) rebacSubject {
	s := rebacSubject{id: id, rdb: rdb, globals: map[permission.ID]bool{}}
	for _, g := range globals {
		s.globals[g] = true
	}
	return s
}

func (s rebacSubject) ID() user.ID                        { return s.id }
func (s rebacSubject) Valid() bool                        { return true }
func (s rebacSubject) HasPermission(p permission.ID) bool { return s.globals[p] }
func (s rebacSubject) HasGroup(group.ID) bool             { return false }
func (s rebacSubject) Context() context.Context           { return context.Background() }

func (s rebacSubject) Audit(p permission.ID) error {
	if s.globals[p] {
		return nil
	}
	return errors.New("denied")
}

func (s rebacSubject) HasResourcePermission(ns rebac.Namespace, inst rebac.Instance, p permission.ID) bool {
	if s.globals[p] {
		return true
	}

	ok, err := s.rdb.Resolve(rebac.Triple{
		Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.id)},
		Relation: rebac.Relation(p),
		Target:   rebac.Entity{Namespace: ns, Instance: inst},
	})
	return err == nil && ok
}

func (s rebacSubject) AuditResource(ns rebac.Namespace, inst rebac.Instance, p permission.ID) error {
	if !s.HasResourcePermission(ns, inst, p) {
		return errors.New("denied")
	}
	return nil
}

// funcCompletions answers every request with fn.
type funcCompletions struct {
	fn func(ctx context.Context, opts completion.Options) (completion.Result, error)
}

func (f *funcCompletions) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *funcCompletions) Complete(ctx context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	return f.fn(ctx, opts)
}

func (f *funcCompletions) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

// grant gives uid the given permissions on the session.
func grant(t *testing.T, rdb *rebac.DB, uid user.ID, sid ID, perms ...permission.ID) {
	t.Helper()
	for _, p := range perms {
		err := rdb.Put(rebac.Triple{
			Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)},
			Relation: rebac.Relation(p),
			Target:   rebac.Entity{Namespace: Namespace, Instance: rebacInstance(sid)},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// A child session belongs to the conversation: whoever may read the parent reads the child, and the owner of
// the parent deletes the children, even those which another user created by continuing the conversation.
func TestChildSessionFollowsParentAccess(t *testing.T) {
	uc, repo, rdb := newTestUseCases(t)
	alice := newRebacSubject("alice", rdb, PermCreate)
	bob := newRebacSubject("bob", rdb, PermCreate)
	carol := newRebacSubject("carol", rdb, PermCreate)

	parent, err := uc.Create(alice, CreateOptions{Model: "fake-model", Title: "shared"})
	if err != nil {
		t.Fatal(err)
	}

	// a child needs a parent the creator may continue
	if _, err := uc.Create(bob, CreateOptions{Model: "fake-model", ParentID: parent.ID}); err == nil {
		t.Fatal("bob must not create a child of a conversation he cannot continue")
	}
	if _, err := uc.Create(bob, CreateOptions{Model: "fake-model", ParentID: "missing"}); err == nil {
		t.Fatal("a child of a missing parent must be refused")
	}

	grant(t, rdb, "bob", parent.ID, PermAppend, PermFindByID)

	// bob continues the conversation, which delegates
	runner := NewSubRunner(uc, parent.ID)
	res, err := runner(context.Background(), bob, completion.SubRunRequest{
		Title: "task",
		Completions: &agentFake{sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			return answerResult("sub answer"), nil
		}},
		Options: completion.Options{Model: "fake-model", System: "You are a sub-agent. test"},
		Input:   []completion.Content{completion.Text{Text: "A"}},
	})
	if err != nil || res.SessionID == "" {
		t.Fatalf("expected a child session: %+v %v", res, err)
	}
	child := ID(res.SessionID)

	for _, c := range []struct {
		name    string
		subject auth.Subject
		sees    bool
	}{{"owner of the parent", alice, true}, {"creator of the child", bob, true}, {"stranger", carol, false}} {
		found, err := uc.FindByID(c.subject, child)
		if err != nil {
			t.Fatal(err)
		}
		if found.IsSome() != c.sees {
			t.Fatalf("%s: expected sees=%v", c.name, c.sees)
		}

		listed := false
		for s, err := range uc.FindAll(c.subject, FindAllOptions{IncludeChildren: true}) {
			if err != nil {
				t.Fatal(err)
			}
			if s.ID == child {
				listed = true
			}
		}
		if listed != c.sees {
			t.Fatalf("%s: expected listed=%v", c.name, c.sees)
		}
	}

	// alice deletes her conversation together with bob's child
	if err := uc.Delete(alice, parent.ID); err != nil {
		t.Fatal(err)
	}
	if opt, _ := repo.FindByID(child); opt.IsSome() {
		t.Fatal("the child of a deleted conversation is left behind")
	}
}

// A run which failed without any progress still books the usage it paid for.
func TestFailedRunWithoutProgressBooksUsage(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	s, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	fake := &funcCompletions{fn: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
		calls++
		if calls == 1 {
			// a truncated answer, which the loop repeats with a larger budget
			return completion.Result{StopReason: completion.StopMaxTokens, Usage: completion.Usage{InputTokens: 300, OutputTokens: 320}}, nil
		}
		cancel()
		return completion.Result{}, ctx.Err()
	}}

	_, err = uc.Append(subject, s.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Agentic:     true,
		Context:     ctx,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancelled run, got %v", err)
	}

	reloaded, _ := uc.FindByID(subject, s.ID)
	if got := reloaded.Unwrap().Usage; got.InputTokens != 300 || got.OutputTokens != 320 {
		t.Fatalf("usage not booked: %+v", got)
	}
	if n := len(reloaded.Unwrap().Messages); n != 0 {
		t.Fatalf("a run without progress must not change the history, got %d messages", n)
	}
}

func TestDismissBooksSubUsageAndDeleteDrainsLedger(t *testing.T) {
	uc, _, s, _ := newPendingSession(t)
	subject := user.SU()

	uc.subUsage.add(s.ID, completion.Usage{InputTokens: 7, OutputTokens: 3})
	dismissed, err := uc.Dismiss(subject, s.ID, s.PendingRevision)
	if err != nil {
		t.Fatal(err)
	}
	if dismissed.SubUsage.InputTokens != 7 || dismissed.SubUsage.OutputTokens != 3 {
		t.Fatalf("sub usage not booked on dismiss: %+v", dismissed.SubUsage)
	}

	uc.subUsage.add(s.ID, completion.Usage{InputTokens: 1})
	if err := uc.Delete(subject, s.ID); err != nil {
		t.Fatal(err)
	}
	uc.subUsage.mu.Lock()
	_, left := uc.subUsage.pending[s.ID]
	uc.subUsage.mu.Unlock()
	if left {
		t.Fatal("the ledger keeps the usage of a deleted session")
	}
}

// Stopping a run which waits for the lock of its session, e.g. because another window runs the same
// conversation, returns right away.
func TestAppendGivesUpWaitingForTheLock(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	s, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	block := &blockingCompletions{started: make(chan struct{}), release: make(chan struct{})}
	first := make(chan error, 1)
	go func() {
		_, err := uc.Append(subject, s.ID, AppendOptions{Completions: block, Input: []completion.Content{completion.Text{Text: "a"}}})
		first <- err
	}()
	<-block.started

	ctx, cancel := context.WithCancel(context.Background())
	second := make(chan error, 1)
	go func() {
		_, err := uc.Append(subject, s.ID, AppendOptions{Completions: &fakeCompletions{}, Input: []completion.Content{completion.Text{Text: "b"}}, Context: ctx})
		second <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-second:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected a cancelled run, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the stopped run still waits for the lock")
	}

	close(block.release)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
}

// A sub-agent without tools still runs the agentic loop, so a truncated answer is continued.
func TestSubRunnerWithoutToolsIsAgentic(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	parent, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	runner := NewSubRunner(uc, parent.ID)
	res, err := runner(context.Background(), subject, completion.SubRunRequest{
		Title: "long",
		Completions: &agentFake{sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			calls++
			if calls == 1 {
				r := answerResult("first part")
				r.StopReason = completion.StopMaxTokens
				return r, nil
			}
			return answerResult("second part"), nil
		}},
		Options: completion.Options{Model: "fake-model", System: "You are a sub-agent. test"},
		Input:   []completion.Content{completion.Text{Text: "A"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if calls != 2 || !strings.Contains(res.Answer, "second part") {
		t.Fatalf("expected the continued answer after %d calls: %q", calls, res.Answer)
	}
}

// A sub-agent which failed before its first turn leaves its task in the child session.
func TestSubRunnerKeepsTheTaskOfAFailedSubAgent(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	parent, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	runner := NewSubRunner(uc, parent.ID)
	res, err := runner(context.Background(), subject, completion.SubRunRequest{
		Title: "broken",
		Completions: &agentFake{sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			return completion.Result{}, errors.New("boom")
		}},
		Options: completion.Options{Model: "fake-model", System: "You are a sub-agent. test"},
		Input:   []completion.Content{completion.Text{Text: "the task"}},
	})
	if err == nil || res.SessionID == "" {
		t.Fatalf("expected a failed run with a child: %v", err)
	}

	child, _ := uc.FindByID(subject, ID(res.SessionID))
	if msgs := child.Unwrap().Messages; len(msgs) != 1 || firstText(t, msgs[0]) != "the task" {
		t.Fatalf("expected the task in the child session, got %+v", msgs)
	}
}

// The usage of the sub-agents of a sub-agent is booked to the conversation.
func TestNestedSubAgentUsageIsBooked(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	parent, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &agentFake{
		parent: func(opts completion.Options) completion.Result {
			if _, ok := lastResult(opts); ok {
				return answerResult("done")
			}
			return toolUseResult(completion.ToolCall{ID: "d1", Name: completion.DelegateToolName, Arguments: stdjson.RawMessage(`{"tasks":[{"title":"child","task":"A"}]}`)})
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if strings.HasPrefix(firstText(t, opts.Messages[0]), "B") {
				r := answerResult("grandchild answer")
				r.Usage = completion.Usage{InputTokens: 100, OutputTokens: 50}
				return r, nil
			}

			if _, ok := lastResult(opts); ok {
				r := answerResult("child answer")
				r.Usage = completion.Usage{InputTokens: 10, OutputTokens: 5}
				return r, nil
			}
			r := toolUseResult(completion.ToolCall{ID: "d2", Name: completion.DelegateToolName, Arguments: stdjson.RawMessage(`{"tasks":[{"title":"grandchild","task":"B"}]}`)})
			r.Usage = completion.Usage{InputTokens: 10, OutputTokens: 5}
			return r, nil
		},
	}

	updated, err := uc.Append(subject, parent.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Tools:       []completion.Tool{completion.NewDelegateTool(completion.DelegateConfig{Runner: NewSubRunner(uc, parent.ID), MaxDepth: 2})},
	})
	if err != nil {
		t.Fatal(err)
	}

	// the child's own two completions plus the grandchild
	if got := updated.SubUsage; got.InputTokens != 120 || got.OutputTokens != 60 {
		t.Fatalf("expected the usage of child and grandchild, got %+v", got)
	}
}

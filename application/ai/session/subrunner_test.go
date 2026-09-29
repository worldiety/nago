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
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// agentFake is a goroutine-safe provider: sub-agents (recognized by their preamble) and the parent are scripted
// separately, because they run concurrently.
type agentFake struct {
	parent func(opts completion.Options) completion.Result
	sub    func(ctx context.Context, opts completion.Options) (completion.Result, error)

	mu sync.Mutex
}

func (f *agentFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *agentFake) Complete(ctx context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	if strings.HasPrefix(opts.System, "You are a sub-agent.") {
		return f.sub(ctx, opts)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	return f.parent(opts), nil
}

func (f *agentFake) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func toolUseResult(calls ...completion.ToolCall) completion.Result {
	var content []completion.Content
	for _, c := range calls {
		content = append(content, c)
	}
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: content},
		StopReason: completion.StopToolUse,
		Usage:      completion.Usage{InputTokens: 1, OutputTokens: 1},
	}
}

// lastResult returns the tool result of the last request message, if any.
func lastResult(opts completion.Options) (completion.ToolResult, bool) {
	for _, c := range opts.Messages[len(opts.Messages)-1].Content {
		if r, ok := c.(completion.ToolResult); ok {
			return r, true
		}
	}
	return completion.ToolResult{}, false
}

func TestSubRunnerPersistsLinkedChildSessions(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	subject := user.SU()

	parent, err := uc.Create(subject, CreateOptions{Model: "fake-model", Title: "parent"})
	if err != nil {
		t.Fatal(err)
	}

	var delegated completion.ToolResult
	fake := &agentFake{
		parent: func(opts completion.Options) completion.Result {
			if r, ok := lastResult(opts); ok {
				delegated = r
				final := answerResult("all done")
				final.Usage = completion.Usage{InputTokens: 1, OutputTokens: 1}
				return final
			}
			return toolUseResult(completion.ToolCall{ID: "d1", Name: completion.DelegateToolName,
				Arguments: stdjson.RawMessage(`{"tasks":[{"title":"one","task":"A"},{"title":"two","task":"B"}]}`)})
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			res := answerResult("sub answer")
			res.Usage = completion.Usage{InputTokens: 10, OutputTokens: 5}
			return res, nil
		},
	}

	updated, err := uc.Append(subject, parent.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Tools:       []completion.Tool{completion.NewDelegateTool(completion.DelegateConfig{Runner: NewSubRunner(uc, parent.ID)})},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Usage of every parent turn, not only the last one, plus the sub-agents separately.
	if updated.Usage != (completion.Usage{InputTokens: 2, OutputTokens: 2}) {
		t.Fatalf("parent usage = %+v", updated.Usage)
	}
	if updated.SubUsage != (completion.Usage{InputTokens: 20, OutputTokens: 10}) {
		t.Fatalf("sub usage = %+v", updated.SubUsage)
	}

	tasks, ok := completion.ParseTaskResults(delegated)
	if !ok || len(tasks) != 2 {
		t.Fatalf("unexpected delegate result %+v", delegated)
	}

	for _, task := range tasks {
		if task.Status != completion.TaskCompleted || task.Answer != "sub answer" || task.SessionID == "" {
			t.Fatalf("unexpected task %+v", task)
		}

		opt, err := uc.FindByID(subject, ID(task.SessionID))
		if err != nil || opt.IsNone() {
			t.Fatalf("child session %q not found: %v", task.SessionID, err)
		}
		child := opt.Unwrap()
		if child.ParentID != parent.ID || child.ParentCallID != "d1" || child.Title != task.Title {
			t.Fatalf("child not linked: %+v", child)
		}
		if len(child.Messages) != 2 || child.Usage != (completion.Usage{InputTokens: 10, OutputTokens: 5}) {
			t.Fatalf("child transcript/usage wrong: %+v", child)
		}

		owner := rebac.Triple{
			Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(subject.ID())},
			Relation: rebac.Owner,
			Target:   rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(child.ID)},
		}
		if ok, err := rdb.Contains(owner); err != nil || !ok {
			t.Fatalf("child has no owner grant: %v %v", ok, err)
		}
	}

	count := func(opts FindAllOptions) int {
		n := 0
		for _, err := range uc.FindAll(subject, opts) {
			if err != nil {
				t.Fatal(err)
			}
			n++
		}
		return n
	}
	if got := count(FindAllOptions{}); got != 1 {
		t.Fatalf("children must be hidden, got %d sessions", got)
	}
	if got := count(FindAllOptions{IncludeChildren: true}); got != 3 {
		t.Fatalf("expected parent and 2 children, got %d", got)
	}

	// Deleting the parent takes the children with it.
	if err := uc.Delete(subject, parent.ID); err != nil {
		t.Fatal(err)
	}
	if got := count(FindAllOptions{IncludeChildren: true}); got != 0 {
		t.Fatalf("children survived their parent: %d", got)
	}
}

func TestSubRunnerFailedSubIsBookedAndPersisted(t *testing.T) {
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
		Input:   []completion.Content{completion.Text{Text: "A"}},
		Tools:   []completion.Tool{completion.NewTool("noop", "does nothing", func(struct{}) (struct{}, error) { return struct{}{}, nil })},
	})
	if err == nil || !strings.Contains(err.Error(), "boom") || res.SessionID == "" {
		t.Fatalf("expected a failed sub run with a child session: %+v %v", res, err)
	}
}

func TestAppendCancelledPersistsPartialHistory(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	s, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopper := completion.NewTool("stop", "stops", func(struct{}) (struct{}, error) { cancel(); return struct{}{}, nil })
	fake := &scriptedCompletions{results: []completion.Result{
		toolUseResult(completion.ToolCall{ID: "1", Name: "stop", Arguments: stdjson.RawMessage(`{}`)}),
		answerResult("never"),
	}}

	_, err = uc.Append(subject, s.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Tools:       []completion.Tool{stopper},
		Context:     ctx,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancellation, got %v", err)
	}

	opt, _ := uc.FindByID(subject, s.ID)
	saved := opt.Unwrap()
	if len(saved.Messages) != 3 || saved.Usage != (completion.Usage{InputTokens: 1, OutputTokens: 1}) {
		t.Fatalf("partial run not persisted: %+v", saved)
	}
}

func TestDismissCancelsBackgroundTasks(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	subject := user.SU()

	s, err := uc.Create(subject, CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &agentFake{
		parent: func(opts completion.Options) completion.Result {
			if r, ok := lastResult(opts); ok && r.ToolCallID == "s" {
				return toolUseResult(completion.ToolCall{ID: "q", Name: completion.AskUserToolName, Arguments: stdjson.RawMessage(`{"question":"which?"}`)})
			}
			return toolUseResult(completion.ToolCall{ID: "s", Name: completion.StartTasksToolName,
				Arguments: stdjson.RawMessage(`{"tasks":[{"title":"slow","task":"A"}]}`)})
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			<-ctx.Done()
			return completion.Result{}, ctx.Err()
		},
	}

	group := uc.Tasks.Group(string(s.ID))
	tools := append(completion.NewTaskTools(completion.DelegateConfig{}, group), completion.NewAskUserTool())

	s, err = uc.Append(subject, s.ID, AppendOptions{
		Completions: fake,
		Input:       []completion.Content{completion.Text{Text: "go"}},
		Tools:       tools,
	})
	if err != nil || s.Pending == nil {
		t.Fatalf("expected a pending question: %v", err)
	}
	if group.Running() != 1 {
		t.Fatalf("background task not running: %+v", group.Snapshot())
	}

	if _, err := uc.Dismiss(subject, s.ID, s.PendingRevision); err != nil {
		t.Fatal(err)
	}

	finished, running := group.Await(context.Background(), nil, 5*time.Second)
	if len(running) != 0 || len(finished) != 1 || finished[0].Status != completion.TaskCancelled {
		t.Fatalf("dismiss did not cancel the task: %+v %+v", finished, running)
	}
}

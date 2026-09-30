// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/auth"
)

// A run which is cancelled while its history is compacted must keep the full history. Otherwise, the excerpt
// a compactor degrades to would replace the conversation for good.
func TestRun_CancelDuringCompactionKeepsHistory(t *testing.T) {
	history := longHistory(20)

	compactors := map[string]func(cancel context.CancelFunc) Compactor{
		"summary": func(cancel context.CancelFunc) Compactor {
			return NewSummaryCompactor(SummaryCompactorConfig{})
		},
		"degrading": func(cancel context.CancelFunc) Compactor {
			return func(ctx context.Context, subject auth.Subject, c Completions, opts Options, history []Message) ([]Message, error) {
				cancel()
				return history[len(history)-2:], nil
			}
		},
	}

	for name, compactor := range compactors {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
				if opts.System == defaultSummaryPrompt {
					// the user stops the run while the summary is requested
					cancel()
					return Result{}, ctx.Err()
				}
				return Result{}, ContextWindowError{Limit: 100, Tokens: 200}
			}}

			out, err := Start(nil, fake, RunOptions{
				Options:   Options{Messages: history},
				Compactor: compactor(cancel),
				Context:   ctx,
			})

			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected a cancelled run, got %v", err)
			}

			if len(out.History) != len(history) || extractText(out.History[0]) != extractText(history[0]) {
				t.Fatalf("history has been replaced: %d messages, first %q", len(out.History), extractText(out.History[0]))
			}
		})
	}
}

func TestTool_MayMutate(t *testing.T) {
	tools := subToolFixture()

	cases := []struct {
		name string
		tool Tool
		want bool
	}{
		{"reading", tools[0], false},
		{"mutating", tools[1], true},
		{"read-only delegation", NewDelegateTool(DelegateConfig{Tools: tools}), false},
		{"mutating delegation", NewDelegateTool(DelegateConfig{Tools: tools, AllowMutating: true}), true},
		{"confirming delegation", NewDelegateTool(DelegateConfig{Tools: tools, AllowMutating: true, ConfirmMutating: true}), false},
	}

	for _, c := range cases {
		if got := c.tool.MayMutate(); got != c.want {
			t.Errorf("%s: expected %v, got %v", c.name, c.want, got)
		}
	}
}

// A mutating delegation tool must keep its sub-agents read-only, when the run confirms mutations. The chat relies
// on this, see [Tool.MayMutate].
func TestDelegate_ConfirmingRunKeepsSubAgentsReadOnly(t *testing.T) {
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if isSub(opts) {
			return echoSub(ctx, opts)
		}
		return parentScript(delegateCall("c1", delegateTaskIn{Title: "a", Task: "A"}))(opts), nil
	}}

	delegate := NewDelegateTool(DelegateConfig{Tools: subToolFixture(), AllowMutating: true})
	_, _, _, err := runDelegation(t, fake, RunOptions{
		Options:         Options{Messages: userMsg("go")},
		Tools:           []Tool{delegate},
		ConfirmMutating: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	subs := fake.subRequests()
	if len(subs) != 1 {
		t.Fatalf("expected a single sub-agent, got %d", len(subs))
	}

	var names []string
	for _, def := range subs[0].Tools {
		names = append(names, def.Name)
	}
	if slices.Contains(names, "write") {
		t.Fatalf("sub-agent got a mutating tool: %v", names)
	}
}

func TestTaskGroup_LimiterHoldsForTheConversation(t *testing.T) {
	g := NewTaskRegistry().Group("k")

	first := g.Limiter(1, 2, true)
	if !first.reserve(2) {
		t.Fatal("cannot reserve the budget")
	}

	// continuing a suspended run keeps the budget
	if cont := g.Limiter(1, 2, false); cont != first || cont.Remaining() != 0 {
		t.Fatalf("continued run got a new budget: %d", cont.Remaining())
	}

	// a new run gets a new budget, but the slots of running tasks stay occupied
	if err := first.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}

	renewed := g.Limiter(1, 2, true)
	if renewed.Remaining() != 2 {
		t.Fatalf("expected a fresh budget, got %d", renewed.Remaining())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := renewed.acquire(ctx); err == nil {
		t.Fatal("a new run must not exceed the parallelism of the conversation")
	}
}

func TestLimiter_CancelledContextNeverAcquires(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for range 100 {
		l := NewLimiter(5, 5)
		if l.acquire(ctx) == nil {
			t.Fatal("acquired a slot with a cancelled context")
		}
	}
}

// Tasks of a cancelled group finish right away and never run their sub-agent.
func TestTasks_StartInCancelledGroup(t *testing.T) {
	subs := newGatedSubs(nil)
	group := NewTaskRegistry().Group("k")
	group.Cancel()

	var runs atomic.Int32
	runner := func(ctx context.Context, subject auth.Subject, req SubRunRequest) (SubRunResult, error) {
		runs.Add(1)
		return DefaultSubRunner(ctx, subject, req)
	}
	tools := toolsByName(NewTaskTools(DelegateConfig{MaxTasksPerRun: 100, Runner: runner}, group))

	for range 20 {
		started := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], subs, delegateIn{Tasks: []delegateTaskIn{{Title: "a", Task: "A"}}}))
		if len(started.Tasks) != 1 || started.Tasks[0].Status != TaskRunning {
			t.Fatalf("unexpected start result %+v", started)
		}

		done := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], subs, awaitIn{IDs: []string{started.Tasks[0].ID}, TimeoutSeconds: 5}))
		if len(done.Finished) != 1 || done.Finished[0].Status != TaskCancelled {
			t.Fatalf("unexpected result %+v", done)
		}
	}

	if n := runs.Load(); n != 0 {
		t.Fatalf("%d sub-agents ran within a cancelled group", n)
	}
}

func TestDelegate_PanickingOnEventDoesNotCrash(t *testing.T) {
	tool := NewDelegateTool(DelegateConfig{OnEvent: func(SubEvent) { panic("boom") }})
	fake := &scriptFake{respond: echoSub}

	out := decode[delegateOut](t, invoke(t, tool, fake, delegateIn{Tasks: []delegateTaskIn{{Title: "a", Task: "A"}}}))
	if len(out.Results) != 1 || out.Results[0].Status != TaskCompleted {
		t.Fatalf("unexpected result %+v", out)
	}
}

func TestTasks_CancelledTasksAreNotHandedOver(t *testing.T) {
	subs := newGatedSubs(nil)
	group := NewTaskRegistry().Group("k")
	tools := toolsByName(NewTaskTools(DelegateConfig{}, group))

	started := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], subs, delegateIn{Tasks: []delegateTaskIn{{Title: "a", Task: "A"}}}))
	invoke(t, tools[CancelTasksToolName], subs, cancelIn{IDs: []string{started.Tasks[0].ID}})

	if extra := group.BeforeFinish(time.Second)(context.Background()); extra != "" {
		t.Fatalf("a cancelled task has been handed over: %s", extra)
	}

	status := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], subs, awaitIn{}))
	if len(status.Finished) != 0 || len(status.Running) != 0 {
		t.Fatalf("a cancelled task is still awaited: %+v", status)
	}
}

func TestTaskRegistry_HeldGroupIsNotSwept(t *testing.T) {
	clock := time.Now()
	r := NewTaskRegistry()
	r.now = func() time.Time { return clock }

	g := r.Group("a")
	release := g.Hold()

	clock = clock.Add(2 * DefaultTaskTTL)
	r.Group("b") // sweeps
	if r.Group("a") != g || g.ctx.Err() != nil {
		t.Fatal("a held group has been swept")
	}

	release()
	clock = clock.Add(2 * DefaultTaskTTL)
	r.Group("b")
	if r.Group("a") == g || g.ctx.Err() == nil {
		t.Fatal("an idle group has not been swept")
	}
}

// BeforeFinish is also consulted, when the model signals tool use without any call.
func TestRun_BeforeFinishOnToolUseWithoutCalls(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopToolUse, Text{Text: "first"}),
		assistantText(StopEndTurn, Text{Text: "second"}),
	}}

	hooks := 0
	out, err := Start(nil, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		BeforeFinish: func(ctx context.Context) string {
			hooks++
			if hooks == 1 {
				return "late results"
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if hooks != 2 || fake.calls != 2 || FinalAnswer(out.History) != "second" {
		t.Fatalf("hooks=%d calls=%d history=%+v", hooks, fake.calls, out.History)
	}
}

// A held group keeps the results of finished tasks, even when the run works longer than the TTL.
func TestTaskRegistry_HeldGroupKeepsResults(t *testing.T) {
	clock := time.Now()
	r := NewTaskRegistry()
	r.now = func() time.Time { return clock }

	g := r.Group("a")
	release := g.Hold()
	tools := toolsByName(NewTaskTools(DelegateConfig{}, g))
	fake := &scriptFake{respond: echoSub}

	started := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], fake, delegateIn{Tasks: []delegateTaskIn{{Title: "a", Task: "A"}}}))
	<-g.tasks[started.Tasks[0].ID].done

	clock = clock.Add(2 * DefaultTaskTTL)
	r.Group("other") // sweeps

	if extra := g.BeforeFinish(time.Second)(context.Background()); !strings.Contains(extra, "answer to A") {
		t.Fatalf("the result of a held group has been dropped: %q", extra)
	}
	release()
}

func TestTaskGroup_CancelRunningKeepsResults(t *testing.T) {
	subs := newGatedSubs(nil)
	group := NewTaskRegistry().Group("k")
	tools := toolsByName(NewTaskTools(DelegateConfig{}, group))
	fake := &scriptFake{respond: echoSub}

	done := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], fake, delegateIn{Tasks: []delegateTaskIn{{Title: "done", Task: "D"}}}))
	<-group.tasks[done.Tasks[0].ID].done
	running := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], subs, delegateIn{Tasks: []delegateTaskIn{{Title: "running", Task: "R"}}}))

	if n := group.CancelRunning(); n != 1 {
		t.Fatalf("expected a single cancelled task, got %d", n)
	}

	res := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], fake, awaitIn{IDs: []string{done.Tasks[0].ID, running.Tasks[0].ID}, TimeoutSeconds: 5}))
	byID := map[string]TaskResult{}
	for _, r := range res.Finished {
		byID[r.ID] = r
	}

	if byID[done.Tasks[0].ID].Status != TaskCompleted || byID[running.Tasks[0].ID].Status != TaskCancelled {
		t.Fatalf("unexpected results %+v", res)
	}
}

// BeforeFinish is also consulted, when the answer stays truncated at the token limit.
func TestRun_BeforeFinishOnMaxTokens(t *testing.T) {
	fake := &fakeCompletions{results: []Result{assistantText(StopMaxTokens, Text{Text: "partial"})}}

	hooks := 0
	out, err := Start(nil, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		BeforeFinish: func(ctx context.Context) string {
			hooks++
			if hooks == 1 {
				return "late results"
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if hooks != 2 {
		t.Fatalf("expected the hook on both ends, got %d", hooks)
	}

	if !slices.ContainsFunc(out.History, func(m Message) bool { return IsLoopPrompt(m) && strings.Contains(extractText(m), "late results") }) {
		t.Fatalf("the hook has not been injected: %+v", out.History)
	}
}

// The summary compactor reports a cancellation instead of degrading to an excerpt.
func TestSummaryCompactor_ReturnsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		cancel()
		return Result{}, ctx.Err()
	}}

	history := longHistory(20)
	out, err := NewSummaryCompactor(SummaryCompactorConfig{})(ctx, nil, fake, Options{}, history)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancellation, got %v with %d messages", err, len(out))
	}
}

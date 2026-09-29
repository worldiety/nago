// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

// scriptFake is a goroutine-safe [Completions] whose answer is chosen per request by respond, so parent and
// sub-agents - which run concurrently - can be scripted independently. It tracks how many requests are in
// flight at the same time.
type scriptFake struct {
	respond func(ctx context.Context, opts Options) (Result, error)

	mu        sync.Mutex
	reqs      []Options
	active    int
	maxActive int
}

func (f *scriptFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *scriptFake) Complete(ctx context.Context, _ auth.Subject, opts Options) (Result, error) {
	f.mu.Lock()
	f.reqs = append(f.reqs, opts)
	f.active++
	f.maxActive = max(f.maxActive, f.active)
	f.mu.Unlock()

	defer func() {
		f.mu.Lock()
		f.active--
		f.mu.Unlock()
	}()

	return f.respond(ctx, opts)
}

func (f *scriptFake) Stream(context.Context, auth.Subject, Options) iter.Seq2[Delta, error] {
	return nil
}

// subRequests returns the requests issued by sub-agents.
func (f *scriptFake) subRequests() []Options {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []Options
	for _, r := range f.reqs {
		if isSub(r) {
			out = append(out, r)
		}
	}
	return out
}

func isSub(opts Options) bool {
	return strings.HasPrefix(opts.System, subAgentPreamble)
}

// firstUserText is the task of a sub-agent.
func firstUserText(opts Options) string {
	if len(opts.Messages) == 0 {
		return ""
	}
	return extractText(opts.Messages[0])
}

// lastToolResult returns the tool result of the last message of a request.
func lastToolResult(opts Options) (ToolResult, bool) {
	if len(opts.Messages) == 0 {
		return ToolResult{}, false
	}
	for _, c := range opts.Messages[len(opts.Messages)-1].Content {
		if r, ok := c.(ToolResult); ok {
			return r, true
		}
	}
	return ToolResult{}, false
}

func textResult(text string, usage Usage) Result {
	return Result{Message: Message{Role: Assistant, Content: []Content{Text{Text: text}}}, StopReason: StopEndTurn, Usage: usage}
}

func delegateCall(id string, tasks ...delegateTaskIn) ToolCall {
	buf, _ := json.Marshal(delegateIn{Tasks: tasks})
	return ToolCall{ID: id, Name: DelegateToolName, Arguments: buf}
}

// parentScript lets the parent call first, then answers with a text once the call has a result.
func parentScript(first ToolCall) func(opts Options) Result {
	return func(opts Options) Result {
		if _, ok := lastToolResult(opts); ok {
			return textResult("parent done", Usage{InputTokens: 1, OutputTokens: 1})
		}
		return Result{Message: Message{Role: Assistant, Content: []Content{first}}, StopReason: StopToolUse, Usage: Usage{InputTokens: 1, OutputTokens: 1}}
	}
}

// runDelegation drives a parent which calls delegate once and returns the decoded tool result.
func runDelegation(t *testing.T, fake *scriptFake, opts RunOptions) (delegateOut, ToolResult, Outcome, error) {
	t.Helper()
	out, err := Start(nil, fake, opts)

	for _, m := range out.History {
		for _, c := range m.Content {
			if r, ok := c.(ToolResult); ok {
				var decoded delegateOut
				_ = json.Unmarshal([]byte(extractText(Message{Content: r.Content})), &decoded)
				return decoded, r, out, err
			}
		}
	}

	t.Fatalf("no tool result in history (err %v): %+v", err, out.History)
	return delegateOut{}, ToolResult{}, out, err
}

func echoSub(ctx context.Context, opts Options) (Result, error) {
	return textResult("answer to "+firstUserText(opts), Usage{InputTokens: 10, OutputTokens: 5}), nil
}

func TestDelegate_RunsTasksInParallel(t *testing.T) {
	const n = 3
	var arrivals sync.WaitGroup
	arrivals.Add(n)
	release := make(chan struct{})
	go func() { arrivals.Wait(); close(release) }()

	call := delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}, delegateTaskIn{Title: "b", Task: "B"}, delegateTaskIn{Title: "c", Task: "C"})
	parent := parentScript(call)

	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}

		// Barrier: every sub-agent must be in flight at the same time, which is impossible sequentially.
		arrivals.Done()
		select {
		case <-release:
		case <-time.After(5 * time.Second):
			return Result{}, errors.New("sub-agents did not run in parallel")
		}
		return echoSub(ctx, opts)
	}}

	res, raw, out, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go"), Model: "parent-model", System: "be precise"},
		Tools:   []Tool{NewDelegateTool(DelegateConfig{})},
	})
	if err != nil {
		t.Fatal(err)
	}

	if raw.IsError || len(res.Results) != n {
		t.Fatalf("unexpected result %+v", res)
	}
	for i, r := range res.Results {
		want := "answer to " + []string{"A", "B", "C"}[i]
		if r.Status != TaskCompleted || r.Answer != want {
			t.Fatalf("task %d: %+v", i, r)
		}
	}

	if FinalAnswer(out.History) != "parent done" {
		t.Fatalf("parent did not finish: %+v", out.History)
	}

	// The sub-agents inherit model and system prompt of the parent, behind the fixed preamble.
	for _, r := range fake.subRequests() {
		if r.Model != "parent-model" || !strings.HasSuffix(r.System, "be precise") {
			t.Fatalf("sub request did not inherit model/system: %q %q", r.Model, r.System)
		}
		if len(r.Messages) != 1 {
			t.Fatalf("sub-agent must not see the parent history: %+v", r.Messages)
		}
	}
}

func TestDelegate_RespectsMaxParallel(t *testing.T) {
	var tasks []delegateTaskIn
	for i := range 6 {
		tasks = append(tasks, delegateTaskIn{Title: fmt.Sprint(i), Task: fmt.Sprint("T", i)})
	}
	parent := parentScript(delegateCall("d1", tasks...))

	var active, maxActive atomic.Int32
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		cur := active.Add(1)
		for {
			old := maxActive.Load()
			if cur <= old || maxActive.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return echoSub(ctx, opts)
	}}

	res, _, _, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		Tools:   []Tool{NewDelegateTool(DelegateConfig{MaxParallel: 2})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 6 {
		t.Fatalf("expected 6 results, got %+v", res)
	}
	if got := maxActive.Load(); got != 2 {
		t.Fatalf("expected exactly 2 concurrent sub-agents, got %d", got)
	}
}

// subToolFixture is a parent tool set with every kind of tool a sub-agent must or must not see.
func subToolFixture() []Tool {
	read := NewTool("read", "reads", func(struct{}) (struct{}, error) { return struct{}{}, nil })
	write := NewTool("write", "writes", func(struct{}) (struct{}, error) { return struct{}{}, nil }).AsMutating("writes")
	screen := NewTool("screen", "looks at the screen", func(struct{}) (struct{}, error) { return struct{}{}, nil })
	screen.NoDelegate = true
	return []Tool{read, write, screen, NewAskUserTool()}
}

func toolNames(defs []ToolDef) []string {
	var names []string
	for _, d := range defs {
		names = append(names, d.Name)
	}
	slices.Sort(names)
	return names
}

func TestDelegate_SubToolSet(t *testing.T) {
	cases := []struct {
		name    string
		cfg     DelegateConfig
		confirm bool
		want    []string
	}{
		{name: "read-only by default", want: []string{"read"}},
		{name: "mutating allowed", cfg: DelegateConfig{AllowMutating: true}, want: []string{"read", "write"}},
		{name: "never with confirmation in config", cfg: DelegateConfig{AllowMutating: true, ConfirmMutating: true}, want: []string{"read"}},
		{name: "never with confirming parent", cfg: DelegateConfig{AllowMutating: true}, confirm: true, want: []string{"read"}},
		{name: "allowlist", cfg: DelegateConfig{AllowMutating: true, AllowedTools: []string{"write", "screen"}}, want: []string{"write"}},
		{name: "nested delegate below max depth", cfg: DelegateConfig{MaxDepth: 2}, want: []string{"delegate", "read"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}))
			fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
				if !isSub(opts) {
					return parent(opts), nil
				}
				return echoSub(ctx, opts)
			}}

			// cfg.Tools stays nil: the sub tools are derived from the tools of the calling run.
			tools := append(subToolFixture(), NewDelegateTool(tc.cfg))
			res, _, _, err := runDelegation(t, fake, RunOptions{
				Options:         Options{Messages: userMsg("go")},
				Tools:           tools,
				ConfirmMutating: tc.confirm,
			})
			if err != nil {
				t.Fatal(err)
			}
			if res.Results[0].Status != TaskCompleted {
				t.Fatalf("task failed: %+v", res.Results[0])
			}

			subs := fake.subRequests()
			if len(subs) != 1 {
				t.Fatalf("expected one sub request, got %d", len(subs))
			}
			if got := toolNames(subs[0].Tools); !slices.Equal(got, tc.want) {
				t.Fatalf("sub tools = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDelegate_NestedSubCannotDelegateFurther(t *testing.T) {
	// depth 1 delegates once more; the sub-agent at depth 2 must not get a delegate tool any more.
	fake := &scriptFake{}
	fake.respond = func(ctx context.Context, opts Options) (Result, error) {
		switch {
		case !isSub(opts):
			return parentScript(delegateCall("d1", delegateTaskIn{Title: "outer", Task: "OUTER"}))(opts), nil
		case firstUserText(opts) == "OUTER":
			return parentScript(delegateCall("d2", delegateTaskIn{Title: "inner", Task: "INNER"}))(opts), nil
		default:
			return echoSub(ctx, opts)
		}
	}

	res, _, _, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go"), System: "root"},
		Tools:   []Tool{NewDelegateTool(DelegateConfig{MaxDepth: 2})},
	})
	if err != nil || res.Results[0].Status != TaskCompleted {
		t.Fatalf("err %v, %+v", err, res)
	}

	var inner *Options
	for _, r := range fake.subRequests() {
		if firstUserText(r) == "INNER" {
			inner = &r
		}
	}
	if inner == nil {
		t.Fatal("the nested sub-agent never ran")
	}
	if len(inner.Tools) != 0 {
		t.Fatalf("depth limit violated, nested sub-agent got tools %v", toolNames(inner.Tools))
	}
	if strings.Count(inner.System, subAgentPreamble) != 1 || !strings.HasSuffix(inner.System, "root") {
		t.Fatalf("nested system prompt is wrong: %q", inner.System)
	}
}

func TestDelegate_ForbiddenToolFailsOnlyThatTask(t *testing.T) {
	parent := parentScript(delegateCall("d1",
		delegateTaskIn{Title: "bad", Task: "A", Tools: []string{"write"}},
		delegateTaskIn{Title: "good", Task: "B", Tools: []string{"read"}},
	))
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		return echoSub(ctx, opts)
	}}

	res, raw, _, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		Tools:   append(subToolFixture(), NewDelegateTool(DelegateConfig{})),
	})
	if err != nil {
		t.Fatal(err)
	}
	if raw.IsError {
		t.Fatal("one successful task must not flag the call as error")
	}
	if res.Results[0].Status != TaskFailed || !strings.Contains(res.Results[0].Error, `"write"`) {
		t.Fatalf("forbidden tool not refused: %+v", res.Results[0])
	}
	if res.Results[1].Status != TaskCompleted {
		t.Fatalf("other task affected: %+v", res.Results[1])
	}
	if subs := fake.subRequests(); len(subs) != 1 || !slices.Equal(toolNames(subs[0].Tools), []string{"read"}) {
		t.Fatalf("unexpected sub requests %+v", subs)
	}
}

func TestDelegate_FailuresAndIsError(t *testing.T) {
	respond := func(failAll bool) func(ctx context.Context, opts Options) (Result, error) {
		parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}, delegateTaskIn{Title: "b", Task: "B"}))
		return func(ctx context.Context, opts Options) (Result, error) {
			if !isSub(opts) {
				return parent(opts), nil
			}
			if failAll || firstUserText(opts) == "A" {
				return Result{}, errors.New("boom")
			}
			return echoSub(ctx, opts)
		}
	}

	res, raw, _, err := runDelegation(t, &scriptFake{respond: respond(false)}, RunOptions{
		Options: Options{Messages: userMsg("go")}, Tools: []Tool{NewDelegateTool(DelegateConfig{})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if raw.IsError || res.Results[0].Status != TaskFailed || !strings.Contains(res.Results[0].Error, "boom") || res.Results[1].Status != TaskCompleted {
		t.Fatalf("one failure must keep the others: %+v", res)
	}

	res, raw, _, err = runDelegation(t, &scriptFake{respond: respond(true)}, RunOptions{
		Options: Options{Messages: userMsg("go")}, Tools: []Tool{NewDelegateTool(DelegateConfig{})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !raw.IsError || len(res.Results) != 2 {
		t.Fatalf("all failed must be an error result: %+v", raw)
	}
}

func TestDelegate_InvalidInput(t *testing.T) {
	for name, call := range map[string]ToolCall{
		"empty":      delegateCall("d1"),
		"too many":   delegateCall("d1", delegateTaskIn{Task: "a"}, delegateTaskIn{Task: "b"}, delegateTaskIn{Task: "c"}),
		"over quota": delegateCall("d1", delegateTaskIn{Task: "a"}, delegateTaskIn{Task: "b"}),
	} {
		t.Run(name, func(t *testing.T) {
			parent := parentScript(call)
			fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
				if !isSub(opts) {
					return parent(opts), nil
				}
				return echoSub(ctx, opts)
			}}

			cfg := DelegateConfig{MaxTasksPerCall: 2}
			if name == "over quota" {
				cfg.MaxTasksPerRun = 1
			}
			_, raw, _, err := runDelegation(t, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: []Tool{NewDelegateTool(cfg)}})
			if err != nil {
				t.Fatal(err)
			}
			if !raw.IsError || len(fake.subRequests()) != 0 {
				t.Fatalf("invalid input must fail without starting anything: %+v", raw)
			}
		})
	}
}

func TestDelegate_TimeoutAndCancellation(t *testing.T) {
	blockingSub := func(ctx context.Context, opts Options) (Result, error) {
		<-ctx.Done()
		return Result{}, ctx.Err()
	}

	t.Run("timeout", func(t *testing.T) {
		parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "slow", Task: "A"}))
		fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
			if !isSub(opts) {
				return parent(opts), nil
			}
			return blockingSub(ctx, opts)
		}}

		res, raw, _, err := runDelegation(t, fake, RunOptions{
			Options: Options{Messages: userMsg("go")},
			Tools:   []Tool{NewDelegateTool(DelegateConfig{SubTimeout: 50 * time.Millisecond})},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !raw.IsError || res.Results[0].Status != TaskTimeout {
			t.Fatalf("expected a timeout, got %+v", res.Results)
		}
	})

	t.Run("parent cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}, delegateTaskIn{Title: "b", Task: "B"}))
		var started atomic.Int32
		fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
			if !isSub(opts) {
				return parent(opts), nil
			}
			if started.Add(1) == 2 {
				cancel()
			}
			return blockingSub(ctx, opts)
		}}

		res, _, out, err := runDelegation(t, fake, RunOptions{
			Options: Options{Messages: userMsg("go")},
			Tools:   []Tool{NewDelegateTool(DelegateConfig{})},
			Context: ctx,
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected the parent run to be cancelled, got %v", err)
		}
		for _, r := range res.Results {
			if r.Status != TaskCancelled {
				t.Fatalf("expected cancelled tasks, got %+v", res.Results)
			}
		}
		if hasDanglingToolUse(out.History) || !out.Progressed {
			t.Fatalf("cancelled run left an invalid history: %+v", out.History)
		}
	})
}

func TestDelegate_UsageAndModels(t *testing.T) {
	parent := parentScript(delegateCall("d1",
		delegateTaskIn{Title: "default", Task: "A"},
		delegateTaskIn{Title: "picked", Task: "B", Model: "small"},
		delegateTaskIn{Title: "refused", Task: "C", Model: "huge"},
	))
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		return echoSub(ctx, opts)
	}}

	var mu sync.Mutex
	var subUsage Usage
	res, _, out, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go"), Model: "big"},
		Tools: []Tool{NewDelegateTool(DelegateConfig{
			Models: []model.ID{"small"},
			OnUsage: func(u Usage) {
				mu.Lock()
				subUsage = subUsage.Add(u)
				mu.Unlock()
			},
		})},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.Results[0].Status != TaskCompleted || res.Results[1].Status != TaskCompleted || res.Results[2].Status != TaskFailed {
		t.Fatalf("unexpected statuses %+v", res.Results)
	}

	models := map[string]model.ID{}
	for _, r := range fake.subRequests() {
		models[firstUserText(r)] = r.Model
	}
	if models["A"] != "big" || models["B"] != "small" || len(models) != 2 {
		t.Fatalf("unexpected sub models %v", models)
	}

	if res.Results[0].Usage != (Usage{InputTokens: 10, OutputTokens: 5}) {
		t.Fatalf("task usage missing: %+v", res.Results[0])
	}
	if subUsage != (Usage{InputTokens: 20, OutputTokens: 10}) {
		t.Fatalf("sub usage = %+v", subUsage)
	}
	// The parent reports only its own two turns; sub-agents are accounted separately.
	if out.Usage != (Usage{InputTokens: 2, OutputTokens: 2}) {
		t.Fatalf("parent usage = %+v", out.Usage)
	}
}

func TestDelegate_RetriesRateLimit(t *testing.T) {
	prev := rateLimitBackoff
	rateLimitBackoff = []time.Duration{time.Millisecond, time.Millisecond}
	defer func() { rateLimitBackoff = prev }()

	parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}))
	var attempts atomic.Int32
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		if attempts.Add(1) < 3 {
			return Result{}, fmt.Errorf("provider says: %w", TooManyRequests)
		}
		return echoSub(ctx, opts)
	}}

	res, _, _, err := runDelegation(t, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: []Tool{NewDelegateTool(DelegateConfig{})}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Results[0].Status != TaskCompleted || attempts.Load() != 3 {
		t.Fatalf("rate limit not retried: %+v after %d attempts", res.Results[0], attempts.Load())
	}
}

func TestDelegate_DeferredNextToQuestionWhenSubsMayMutate(t *testing.T) {
	for _, allow := range []bool{false, true} {
		fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
			if isSub(opts) {
				return echoSub(ctx, opts)
			}
			return toolUse(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}), askCall("q")), nil
		}}

		tools := append(subToolFixture(), NewDelegateTool(DelegateConfig{AllowMutating: allow}))
		out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: tools})
		if err != nil || out.Suspended == nil || len(out.Suspended.Completed) != 1 {
			t.Fatalf("expected a suspension: %v", err)
		}

		deferred := extractText(Message{Content: out.Suspended.Completed[0].Content}) == deferredText
		if ran := len(fake.subRequests()) > 0; ran == deferred {
			t.Fatalf("allow=%v: deferred=%v but sub-agents ran=%v", allow, deferred, ran)
		}
		if deferred != allow {
			t.Fatalf("allow=%v: deferred=%v, completed=%+v", allow, deferred, out.Suspended.Completed)
		}
	}
}

func TestDelegate_EventsAndSuspendedSubIsError(t *testing.T) {
	parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}))
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		return echoSub(ctx, opts)
	}}

	var mu sync.Mutex
	var kinds []SubEventKind
	suspending := func(ctx context.Context, subject auth.Subject, req SubRunRequest) (SubRunResult, error) {
		return SubRunResult{}, ErrSubRunSuspended
	}

	res, _, _, err := runDelegation(t, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		Tools: []Tool{NewDelegateTool(DelegateConfig{Runner: suspending, OnEvent: func(e SubEvent) {
			mu.Lock()
			kinds = append(kinds, e.Kind)
			mu.Unlock()
		}})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Results[0].Status != TaskFailed {
		t.Fatalf("a suspended sub-agent must fail its task: %+v", res.Results[0])
	}
	if !slices.Equal(kinds, []SubEventKind{SubStarted, SubFinished}) {
		t.Fatalf("unexpected events %v", kinds)
	}
}

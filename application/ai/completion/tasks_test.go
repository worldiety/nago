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
	"strings"
	"sync"
	"testing"
	"time"
)

// toolsByName indexes tools for direct invocation in tests.
func toolsByName(tools []Tool) map[string]Tool {
	m := map[string]Tool{}
	for _, t := range tools {
		m[t.Def.Name] = t
	}
	return m
}

// invoke runs a built-in tool as the loop would, for a parent run with the given provider.
func invoke(t *testing.T, tool Tool, c Completions, args any) ToolResult {
	t.Helper()
	buf, _ := json.Marshal(args)
	env := toolEnv{ctx: context.Background(), completions: c, opts: &RunOptions{}}
	return tool.run(env, ToolCall{ID: "call", Name: tool.Def.Name, Arguments: buf})
}

func decode[T any](t *testing.T, res ToolResult) T {
	t.Helper()
	var out T
	if err := json.Unmarshal([]byte(extractText(Message{Content: res.Content})), &out); err != nil {
		t.Fatalf("cannot decode %+v: %v", res, err)
	}
	return out
}

// gatedSubs is a provider whose sub-agents answer only once released.
type gatedSubs struct {
	*scriptFake
	release chan struct{}
}

func newGatedSubs(parent func(opts Options) Result) gatedSubs {
	g := gatedSubs{release: make(chan struct{})}
	g.scriptFake = &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		select {
		case <-g.release:
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
		return echoSub(ctx, opts)
	}}
	return g
}

func TestTasks_StartAwaitCancel(t *testing.T) {
	subs := newGatedSubs(nil)
	group := NewTaskRegistry().Group("s1")
	tools := toolsByName(NewTaskTools(DelegateConfig{}, group))

	started := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], subs, delegateIn{Tasks: []delegateTaskIn{
		{Title: "a", Task: "A"}, {Title: "b", Task: "B"}, {Title: "c", Task: "C"},
	}}))
	if len(started.Tasks) != 3 {
		t.Fatalf("unexpected start result %+v", started)
	}
	for _, task := range started.Tasks {
		if task.ID == "" || task.Status != TaskRunning {
			t.Fatalf("task not running: %+v", task)
		}
	}
	a, b, c := started.Tasks[0].ID, started.Tasks[1].ID, started.Tasks[2].ID

	// timeout 0 only reports the state
	status := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], subs, awaitIn{}))
	if len(status.Finished) != 0 || len(status.Running) != 3 {
		t.Fatalf("unexpected status %+v", status)
	}

	cancelled := decode[cancelOut](t, invoke(t, tools[CancelTasksToolName], subs, cancelIn{IDs: []string{c}}))
	if len(cancelled.Tasks) != 1 || cancelled.Tasks[0].Status != TaskCancelled {
		t.Fatalf("cancel failed: %+v", cancelled)
	}

	close(subs.release)
	done := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], subs, awaitIn{IDs: []string{a, b, c}, TimeoutSeconds: 5}))
	if len(done.Finished) != 3 || len(done.Running) != 0 {
		t.Fatalf("expected all finished: %+v", done)
	}

	byID := map[string]TaskResult{}
	for _, r := range done.Finished {
		byID[r.ID] = r
	}
	if byID[a].Status != TaskCompleted || byID[a].Answer != "answer to A" || byID[b].Status != TaskCompleted {
		t.Fatalf("unexpected results %+v", done.Finished)
	}
	if byID[c].Status != TaskCancelled {
		t.Fatalf("cancelled task finished as %+v", byID[c])
	}

	// every result is delivered once; afterwards the ids are unknown
	again := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], subs, awaitIn{IDs: []string{a, "task_unknown"}}))
	if len(again.Finished) != 2 || again.Finished[0].Status != TaskLost || again.Finished[1].Status != TaskLost {
		t.Fatalf("expected lost ids: %+v", again)
	}
}

func TestTasks_UnknownIDsAfterRestartAreLost(t *testing.T) {
	// A fresh registry is what a restarted process has.
	group := NewTaskRegistry().Group("s1")
	tools := toolsByName(NewTaskTools(DelegateConfig{}, group))

	res := decode[awaitOut](t, invoke(t, tools[AwaitTasksToolName], nil, awaitIn{IDs: []string{"task_0123"}, TimeoutSeconds: 1}))
	if len(res.Finished) != 1 || res.Finished[0].Status != TaskLost || res.Finished[0].ID != "task_0123" {
		t.Fatalf("expected a lost task: %+v", res)
	}
}

func TestTasks_BeforeFinishJoinsForgottenTasks(t *testing.T) {
	start := ToolCall{ID: "s", Name: StartTasksToolName, Arguments: json.RawMessage(`{"tasks":[{"title":"a","task":"A"}]}`)}

	fake := &scriptFake{}
	fake.respond = func(ctx context.Context, opts Options) (Result, error) {
		if isSub(opts) {
			time.Sleep(30 * time.Millisecond)
			return echoSub(ctx, opts)
		}

		last := opts.Messages[len(opts.Messages)-1]
		switch {
		case IsLoopPrompt(last):
			return textResult("final with "+extractText(last), Usage{}), nil
		case len(opts.Messages) == 1:
			return toolUse(start), nil
		default:
			// The model answers without ever awaiting its task.
			return textResult("premature answer", Usage{}), nil
		}
	}

	group := NewTaskRegistry().Group("s1")
	out, err := Start(nil, fake, RunOptions{
		Options:      Options{Messages: userMsg("go")},
		Tools:        NewTaskTools(DelegateConfig{}, group),
		BeforeFinish: group.BeforeFinish(time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}

	final := FinalAnswer(out.History)
	if !strings.HasPrefix(final, "final with") || !strings.Contains(final, "answer to A") {
		t.Fatalf("background result not joined: %q", final)
	}
	if group.Running() != 0 || len(group.Snapshot()) != 1 || group.Snapshot()[0].Status != TaskCompleted {
		t.Fatalf("unexpected group state %+v", group.Snapshot())
	}
}

func TestTasks_SurviveSuspensionAndContinue(t *testing.T) {
	registry := NewTaskRegistry()
	start := ToolCall{ID: "s", Name: StartTasksToolName, Arguments: json.RawMessage(`{"tasks":[{"title":"a","task":"A"},{"title":"b","task":"B"}]}`)}

	var mu sync.Mutex
	var ids []string
	subs := newGatedSubs(func(opts Options) Result {
		if r, ok := lastToolResult(opts); ok {
			switch {
			case r.ToolCallID == "s":
				var out startTasksOut
				_ = json.Unmarshal([]byte(extractText(Message{Content: r.Content})), &out)
				mu.Lock()
				for _, task := range out.Tasks {
					ids = append(ids, task.ID)
				}
				mu.Unlock()
				return toolUse(askCall("q"))
			case r.ToolCallID == "q":
				mu.Lock()
				args, _ := json.Marshal(awaitIn{IDs: ids, TimeoutSeconds: 5})
				mu.Unlock()
				return toolUse(ToolCall{ID: "w", Name: AwaitTasksToolName, Arguments: args})
			case r.ToolCallID == "w":
				return textResult("done: "+extractText(Message{Content: r.Content}), Usage{})
			}
		}
		return toolUse(start)
	})

	tools := func() []Tool {
		// Tools are rebuilt per run, like a UI does; the group is found again by its key.
		return append(NewTaskTools(DelegateConfig{}, registry.Group("s1")), NewAskUserTool())
	}

	out, err := Start(nil, subs, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: tools()})
	if err != nil || out.Suspended == nil {
		t.Fatalf("expected a suspension, err %v", err)
	}

	// The tasks keep running while nobody waits; they finish before the user answers.
	close(subs.release)
	deadline := time.Now().Add(5 * time.Second)
	for registry.Group("s1").Running() > 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	final, err := Continue(nil, subs, RunOptions{Options: Options{Messages: out.History}, Tools: tools()}, *out.Suspended,
		[]Resolution{{CallID: "q", Answer: "a"}})
	if err != nil {
		t.Fatal(err)
	}

	answer := FinalAnswer(final.History)
	if !strings.Contains(answer, "answer to A") || !strings.Contains(answer, "answer to B") {
		t.Fatalf("results lost across the suspension: %q", answer)
	}
}

func TestTasks_RegistryCancelStopsTasks(t *testing.T) {
	subs := newGatedSubs(nil)
	registry := NewTaskRegistry()
	group := registry.Group("s1")
	tools := toolsByName(NewTaskTools(DelegateConfig{}, group))

	started := decode[startTasksOut](t, invoke(t, tools[StartTasksToolName], subs, delegateIn{Tasks: []delegateTaskIn{{Title: "a", Task: "A"}}}))

	registry.Cancel("s1")

	finished, running := group.Await(context.Background(), []string{started.Tasks[0].ID}, 5*time.Second)
	if len(running) != 0 || len(finished) != 1 || finished[0].Status != TaskCancelled {
		t.Fatalf("expected a cancelled task: %+v %+v", finished, running)
	}

	if registry.Group("s1") == group {
		t.Fatal("a cancelled group must be forgotten")
	}
}

func TestTasks_RegistryExpiresResults(t *testing.T) {
	now := time.Now()
	registry := NewTaskRegistry()
	registry.now = func() time.Time { return now }

	group := registry.Group("s1")
	group.spawn("a", func(ctx context.Context, id string) TaskResult {
		return TaskResult{Status: TaskCompleted, Answer: "x"}
	})
	for group.Running() > 0 {
		time.Sleep(time.Millisecond)
	}

	now = now.Add(DefaultTaskTTL + time.Minute)
	registry.mu.Lock()
	registry.sweepLocked()
	_, kept := registry.groups["s1"]
	registry.mu.Unlock()

	if len(group.Snapshot()) != 0 {
		t.Fatalf("expired result kept: %+v", group.Snapshot())
	}
	if kept {
		t.Fatal("idle expired group kept")
	}
}

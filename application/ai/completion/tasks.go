// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"
)

// DefaultTaskTTL is how long a [TaskRegistry] keeps the result of a finished background task nobody picked up,
// and an idle group without tasks.
const DefaultTaskTTL = 30 * time.Minute

// TaskRegistry keeps the background tasks started via [NewTaskTools], grouped by a key - typically the id of the
// parent session, or a random id for a transient chat. It lives in memory only: a group outlives the suspension
// of its parent run (a question to the user, an approval), so [Continue] can pick up the results under the same
// key, but not a restart of the process. Ids which are not known (any more) are reported as [TaskLost].
//
// It is safe for concurrent use.
type TaskRegistry struct {
	mu     sync.Mutex
	groups map[string]*TaskGroup
	ttl    time.Duration
	now    func() time.Time
}

// NewTaskRegistry creates an empty registry with [DefaultTaskTTL].
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{groups: map[string]*TaskGroup{}, ttl: DefaultTaskTTL, now: time.Now}
}

// Group returns the group of key, creating it if necessary. Expired results and idle groups are dropped on the
// way.
func (r *TaskRegistry) Group(key string) *TaskGroup {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sweepLocked()

	g, ok := r.groups[key]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		g = &TaskGroup{key: key, ctx: ctx, cancel: cancel, tasks: map[string]*bgTask{}, now: r.now}
		r.groups[key] = g
	}

	g.touch()
	return g
}

// Cancel cancels every running task of the group of key and forgets the group. A no-op for an unknown key.
func (r *TaskRegistry) Cancel(key string) {
	r.mu.Lock()
	g, ok := r.groups[key]
	delete(r.groups, key)
	r.mu.Unlock()

	if ok {
		g.Cancel()
	}
}

// sweepLocked drops expired task results and idle groups. The caller holds r.mu.
func (r *TaskRegistry) sweepLocked() {
	now := r.now()
	for key, g := range r.groups {
		if g.sweep(now, r.ttl) {
			g.cancel()
			delete(r.groups, key)
		}
	}
}

// TaskGroup holds the background tasks of one conversation. See [TaskRegistry].
type TaskGroup struct {
	key    string
	ctx    context.Context
	cancel context.CancelFunc
	now    func() time.Time

	mu       sync.Mutex
	tasks    map[string]*bgTask
	order    []string
	lastUsed time.Time
	limiter  *Limiter
	// holds counts the runs which use the group, see Hold.
	holds int
}

type bgTask struct {
	result   TaskResult
	cancel   context.CancelFunc
	done     chan struct{}
	consumed bool
	// abandoned is set by CancelTasks: nobody waits for the result any more.
	abandoned  bool
	finishedAt time.Time
}

// Key returns the key of the group.
func (g *TaskGroup) Key() string {
	return g.key
}

func (g *TaskGroup) touch() {
	g.mu.Lock()
	g.lastUsed = g.now()
	g.mu.Unlock()
}

// sweep drops delivered and expired results and reports whether the whole group is idle and expired.
func (g *TaskGroup) sweep(now time.Time, ttl time.Duration) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.order = slices.DeleteFunc(g.order, func(id string) bool {
		t := g.tasks[id]
		if t.result.Status == TaskRunning {
			return false
		}

		// A result expires after the TTL, but not while a run holds the group: it may still pick it up. The
		// TTL restarts when the run releases the group, e.g. on a suspension.
		expired := g.holds == 0 && now.Sub(later(t.finishedAt, g.lastUsed)) > ttl
		if t.consumed || expired {
			delete(g.tasks, id)
			return true
		}
		return false
	})

	return len(g.tasks) == 0 && g.holds == 0 && now.Sub(g.lastUsed) > ttl
}

func later(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// Hold marks the group as used by a run until release is called, so that it is not dropped as idle while the
// run works on something else for longer than the TTL of the registry. Otherwise, the run would start its
// tasks within a cancelled group which the registry does not know any more.
func (g *TaskGroup) Hold() (release func()) {
	g.mu.Lock()
	g.holds++
	g.lastUsed = g.now()
	g.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			g.mu.Lock()
			g.holds--
			g.lastUsed = g.now()
			g.mu.Unlock()
		})
	}
}

// Limiter returns the limiter of the group for [DelegateConfig.Limiter]. Background tasks outlive the run which
// started them, e.g. while it is suspended on a question to the user, thus the limit of concurrent sub-agents
// must hold for the whole conversation instead of each run. With renew, the task budget starts again, e.g. for a
// new question of the user, while the slots of still running tasks remain occupied. Without, e.g. to continue a
// suspended run, the budget carries on. The limits only apply when the limiter is created, see [NewLimiter].
func (g *TaskGroup) Limiter(maxParallel, maxTasks int, renew bool) *Limiter {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch {
	case g.limiter == nil:
		g.limiter = NewLimiter(maxParallel, maxTasks)
	case renew:
		g.limiter = g.limiter.renewed(maxTasks)
	}

	return g.limiter
}

// Cancel cancels every running task of the group. Their results report [TaskCancelled].
func (g *TaskGroup) Cancel() {
	g.cancel()
}

// CancelRunning cancels every task which is still running and returns their number. Nobody waits for them any
// more, see [TaskGroup.CancelTasks]. In contrast to [TaskGroup.Cancel], the group stays usable and keeps the
// results of the finished tasks, so a later run can still pick them up. Use it when a run ended, so that
// nothing keeps working unobserved.
func (g *TaskGroup) CancelRunning() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := 0
	for _, t := range g.tasks {
		if t.result.Status == TaskRunning && !t.abandoned {
			t.abandoned = true
			t.cancel()
			n++
		}
	}

	return n
}

// Running returns the number of tasks which have not finished yet.
func (g *TaskGroup) Running() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := 0
	for _, t := range g.tasks {
		if t.result.Status == TaskRunning {
			n++
		}
	}
	return n
}

// Snapshot returns the current state of every known task, in start order, without marking anything delivered.
func (g *TaskGroup) Snapshot() []TaskResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	out := make([]TaskResult, 0, len(g.order))
	for _, id := range g.order {
		out = append(out, g.tasks[id].result)
	}
	return out
}

// spawn starts run on its own goroutine as a new task of the group.
func (g *TaskGroup) spawn(title string, run func(ctx context.Context, id string) TaskResult) TaskResult {
	id := newTaskID()
	ctx, cancel := context.WithCancel(g.ctx)

	initial := TaskResult{ID: id, Title: title, Status: TaskRunning}
	t := &bgTask{
		result: initial,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	g.mu.Lock()
	g.tasks[id] = t
	g.order = append(g.order, id)
	g.lastUsed = g.now()
	g.mu.Unlock()

	go func() {
		defer cancel()

		res := TaskResult{Title: title, Status: TaskFailed}
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("background task panicked", "task", id, "panic", r, "stack", string(debug.Stack()))
					res = TaskResult{Title: title, Status: TaskFailed, Error: fmt.Sprintf("the sub-agent crashed: %v", r)}
				}
			}()
			res = run(ctx, id)
		}()

		res.ID = id
		if res.Status == "" || res.Status == TaskRunning {
			res.Status = TaskFailed
		}

		g.mu.Lock()
		t.result = res
		t.finishedAt = g.now()
		g.mu.Unlock()
		close(t.done)
	}()

	// t.result belongs to the goroutine now, which may already have finished
	return initial
}

// pick resolves ids to tasks. An empty ids selects every task not delivered or cancelled yet. Unknown ids are
// returned separately.
func (g *TaskGroup) pick(ids []string) (known []*bgTask, lost []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(ids) == 0 {
		for _, id := range g.order {
			if t := g.tasks[id]; !t.consumed && !t.abandoned {
				known = append(known, t)
			}
		}
		return known, nil
	}

	for _, id := range ids {
		// A delivered result is not delivered twice; for the model it is gone like after a restart.
		if t, ok := g.tasks[id]; ok && !t.consumed {
			if !slices.Contains(known, t) {
				known = append(known, t)
			}
			continue
		}
		lost = append(lost, id)
	}
	return known, lost
}

// wait blocks until every task is done, the timeout passed or ctx is cancelled.
func wait(ctx context.Context, tasks []*bgTask, timeout time.Duration) {
	if timeout <= 0 || len(tasks) == 0 {
		return
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for _, t := range tasks {
		select {
		case <-t.done:
		case <-timer.C:
			return
		case <-ctx.Done():
			return
		}
	}
}

// collect splits tasks into finished (marked delivered) and running ones.
func (g *TaskGroup) collect(tasks []*bgTask) (finished, running []TaskResult) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.lastUsed = g.now()
	for _, t := range tasks {
		if t.result.Status == TaskRunning {
			running = append(running, t.result)
			continue
		}
		t.consumed = true
		finished = append(finished, t.result)
	}
	return finished, running
}

// Await waits up to timeout for the given tasks (every undelivered task if ids is empty) and returns the
// finished ones - which are marked delivered - and the ones still running. Unknown ids are reported as
// [TaskLost]. A zero timeout only reports the current state.
func (g *TaskGroup) Await(ctx context.Context, ids []string, timeout time.Duration) (finished, running []TaskResult) {
	known, lost := g.pick(ids)
	wait(ctx, known, timeout)
	finished, running = g.collect(known)

	for _, id := range lost {
		finished = append(finished, TaskResult{ID: id, Status: TaskLost, Error: "unknown task: it never existed, its result was already delivered or it was lost by a restart"})
	}
	return finished, running
}

// CancelTasks cancels the given tasks and returns their current state. Unknown ids are reported as [TaskLost].
// The caller does not wait for them any more, thus they are only returned by [TaskGroup.Await] when asked for
// explicitly, and [TaskGroup.BeforeFinish] does not hand them over.
func (g *TaskGroup) CancelTasks(ids []string) []TaskResult {
	known, lost := g.pick(ids)
	for _, t := range known {
		t.cancel()
	}

	out := make([]TaskResult, 0, len(ids))
	g.mu.Lock()
	g.lastUsed = g.now()
	for _, t := range known {
		t.abandoned = true
		r := t.result
		if r.Status == TaskRunning {
			r.Status = TaskCancelled
		}
		out = append(out, r)
	}
	g.mu.Unlock()

	for _, id := range lost {
		out = append(out, TaskResult{ID: id, Status: TaskLost, Error: "unknown task"})
	}
	return out
}

// BeforeFinish returns a hook for [RunOptions.BeforeFinish]: when the model gives its final answer while
// background tasks were never awaited, it waits up to maxWait for them and hands their results to the model,
// so they are not silently lost. Zero maxWait means [DefaultDelegateSubTimeout].
func (g *TaskGroup) BeforeFinish(maxWait time.Duration) func(ctx context.Context) string {
	if maxWait <= 0 {
		maxWait = DefaultDelegateSubTimeout
	}

	return func(ctx context.Context) string {
		known, _ := g.pick(nil)
		if len(known) == 0 {
			return ""
		}

		wait(ctx, known, maxWait)
		finished, running := g.collect(known)
		if len(finished) == 0 && len(running) == 0 {
			return ""
		}

		buf, err := json.Marshal(awaitOut{Finished: finished, Running: running})
		if err != nil {
			return ""
		}

		return "Background tasks you started finished (or are still running), but you answered without awaiting them. " +
			"Their results follow. Take them into account: complete or correct your answer, and do not start them again.\n" +
			string(buf)
	}
}

// HandedOverTasks extracts the task results which [TaskGroup.BeforeFinish] handed over to the model within the
// given message, see [IsLoopPrompt]. It returns false for any other message. UIs use it to show the final state
// of tasks the model never awaited.
func HandedOverTasks(msg Message) ([]TaskResult, bool) {
	if !IsLoopPrompt(msg) {
		return nil, false
	}

	text := extractText(msg)
	if !strings.HasPrefix(text, finishPromptMarker) {
		return nil, false
	}

	start := strings.Index(text, "{")
	if start < 0 {
		return nil, false
	}

	var out awaitOut
	if err := json.Unmarshal([]byte(text[start:]), &out); err != nil {
		return nil, false
	}

	all := slices.Concat(out.Finished, out.Running)
	return all, len(all) > 0
}

func newTaskID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "task_" + hex.EncodeToString(b[:])
}

type startTasksOut struct {
	Tasks []TaskResult `json:"tasks"`
}

type awaitIn struct {
	IDs            []string `json:"ids,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

type awaitOut struct {
	Finished []TaskResult `json:"finished"`
	Running  []TaskResult `json:"running,omitempty"`
}

type cancelIn struct {
	IDs []string `json:"ids"`
}

type cancelOut struct {
	Tasks []TaskResult `json:"tasks"`
}

// NewTaskTools creates the background task tools start_tasks, await_tasks and cancel_tasks on top of group.
//
// In contrast to [NewDelegateTool], start_tasks returns immediately with the task ids, and the model continues
// working while the sub-agents run; await_tasks collects their results later. The tasks are bound to the group,
// not to the calling run: they keep running while the run is suspended on a user decision, and they are only
// cancelled by cancel_tasks, [TaskGroup.Cancel] or [TaskRegistry.Cancel]. Use [TaskGroup.BeforeFinish] as
// [RunOptions.BeforeFinish], so results the model forgot to await are handed over before the run ends.
//
// The sub-agents are restricted exactly like the ones of the delegate tool, see [DelegateConfig];
// [DelegateConfig.Context] does not apply.
func NewTaskTools(cfg DelegateConfig, group *TaskGroup) []Tool {
	d := &delegator{cfg: cfg.withDefaults()}
	maxWait := d.cfg.SubTimeout

	start := Tool{
		Def: newToolDefFromSchema(StartTasksToolName,
			"Starts independent tasks as sub-agents in the background and returns their ids immediately, so you can continue "+
				"working in the meantime. Collect the results later with await_tasks. Use it when you have other work to do "+
				"while the sub-agents run; if you would only wait anyway, use delegate instead. A sub-agent does not see this "+
				"conversation and cannot ask anything: put everything it needs into task and context.",
			d.taskSchema()),
		NoDelegate:       true,
		indirectMutating: d.mayMutate(),
		run: func(env toolEnv, call ToolCall) ToolResult {
			plans, err := d.planAll(env, call.Arguments, call.ID)
			if err != nil {
				return errorResult(err)
			}

			out := startTasksOut{Tasks: make([]TaskResult, 0, len(plans))}
			started := 0
			for _, p := range plans {
				if p.err != nil {
					out.Tasks = append(out.Tasks, TaskResult{Title: p.title, Status: TaskFailed, Error: p.err.Error()})
					continue
				}

				p := p
				subject := env.subject
				out.Tasks = append(out.Tasks, group.spawn(p.title, func(ctx context.Context, id string) TaskResult {
					return d.execute(ctx, subject, p, id, true)
				}))
				started++
			}

			buf, err := json.Marshal(out)
			if err != nil {
				return errorResult(err)
			}
			return ToolResult{IsError: started == 0, Content: []Content{Text{Text: string(buf)}}}
		},
	}

	awaitSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"},
				"description": "the ids returned by start_tasks; omit to await every task whose result you have not received yet"},
			"timeout_seconds": map[string]any{"type": "integer",
				"description": fmt.Sprintf("how long to wait at most, in seconds (at most %d). 0 only reports the current state without waiting", int(maxWait/time.Second))},
		},
		"required":             []string{"timeout_seconds"},
		"additionalProperties": false,
	}

	await := Tool{
		Def: newToolDefFromSchema(AwaitTasksToolName,
			"Waits for background tasks started with start_tasks and returns the results of the finished ones together with "+
				"the ones still running. Each result is delivered only once. Call it again for tasks which are still running.",
			awaitSchema),
		NoDelegate: true,
		run: func(env toolEnv, call ToolCall) ToolResult {
			var in awaitIn
			if err := json.Unmarshal(call.Arguments, &in); err != nil {
				return errorResult(fmt.Errorf("invalid arguments: %w", err))
			}

			timeout := min(time.Duration(max(in.TimeoutSeconds, 0))*time.Second, maxWait)
			ctx := env.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			finished, running := group.Await(ctx, in.IDs, timeout)
			if finished == nil {
				finished = []TaskResult{}
			}

			buf, err := json.Marshal(awaitOut{Finished: finished, Running: running})
			if err != nil {
				return errorResult(err)
			}
			return ToolResult{Content: []Content{Text{Text: string(buf)}}}
		},
	}

	cancelSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "the ids of the tasks to cancel"},
		},
		"required":             []string{"ids"},
		"additionalProperties": false,
	}

	cancel := Tool{
		Def: newToolDefFromSchema(CancelTasksToolName,
			"Cancels background tasks started with start_tasks whose results are no longer needed.",
			cancelSchema),
		NoDelegate: true,
		run: func(env toolEnv, call ToolCall) ToolResult {
			var in cancelIn
			if err := json.Unmarshal(call.Arguments, &in); err != nil {
				return errorResult(fmt.Errorf("invalid arguments: %w", err))
			}
			if len(in.IDs) == 0 {
				return errorResult(fmt.Errorf("ids must not be empty"))
			}

			buf, err := json.Marshal(cancelOut{Tasks: group.CancelTasks(in.IDs)})
			if err != nil {
				return errorResult(err)
			}
			return ToolResult{Content: []Content{Text{Text: string(buf)}}}
		},
	}

	return []Tool{start, await, cancel}
}

// IsDelegationTool reports whether name is one of the delegation tools of this package. UIs use it to render
// their calls and results as sub tasks instead of plain tool hints.
func IsDelegationTool(name string) bool {
	switch name {
	case DelegateToolName, StartTasksToolName, AwaitTasksToolName, CancelTasksToolName:
		return true
	default:
		return false
	}
}

// ParseTaskResults extracts the task results from the result of a delegation tool call, see [IsDelegationTool].
// It returns false if res does not carry any.
func ParseTaskResults(res ToolResult) ([]TaskResult, bool) {
	var sb strings.Builder
	for _, c := range res.Content {
		if t, ok := c.(Text); ok {
			sb.WriteString(t.Text)
		}
	}

	var out struct {
		Results  []TaskResult `json:"results"`
		Tasks    []TaskResult `json:"tasks"`
		Finished []TaskResult `json:"finished"`
		Running  []TaskResult `json:"running"`
	}
	if err := json.Unmarshal([]byte(sb.String()), &out); err != nil {
		return nil, false
	}

	all := slices.Concat(out.Results, out.Tasks, out.Finished, out.Running)
	return all, len(all) > 0
}

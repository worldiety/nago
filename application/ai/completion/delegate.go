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
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

// Names of the delegation tools. They are reserved: a sub-agent never receives a tool of these names from its
// parent (see [Tool.NoDelegate]).
const (
	DelegateToolName    = "delegate"
	StartTasksToolName  = "start_tasks"
	AwaitTasksToolName  = "await_tasks"
	CancelTasksToolName = "cancel_tasks"
)

// Defaults of [DelegateConfig].
const (
	// DefaultDelegateMaxDepth lets sub-agents work, but not delegate any further.
	DefaultDelegateMaxDepth = 1
	// DefaultDelegateMaxParallel bounds how many sub-agents of one run work at the same time. More tasks are
	// accepted but wait for a free slot.
	DefaultDelegateMaxParallel = 4
	// DefaultDelegateMaxTasksPerCall bounds the number of tasks of a single delegate or start_tasks call.
	DefaultDelegateMaxTasksPerCall = 8
	// DefaultDelegateMaxTasksPerRun bounds the number of sub-agents one run may start in total.
	DefaultDelegateMaxTasksPerRun = 16
	// DefaultDelegateSubMaxTurns bounds the agentic loop of a single sub-agent.
	DefaultDelegateSubMaxTurns = 20
	// DefaultDelegateSubTimeout bounds how long a single sub-agent may work.
	DefaultDelegateSubTimeout = 5 * time.Minute
)

// ErrSubRunSuspended is reported for a sub-agent whose run suspended on a user decision. A sub-agent has no
// user; it never receives [Tool.AwaitsUser] tools and always runs without [RunOptions.ConfirmMutating] and
// [RunOptions.ConfirmMarked], so
// this only happens with a custom [SubRunner] or a misconfigured tool.
var ErrSubRunSuspended = errors.New("sub task tried to wait for a user decision")

// TaskStatus is the state of a delegated task, see [TaskResult].
type TaskStatus string

const (
	// TaskRunning is a background task which has not finished yet.
	TaskRunning TaskStatus = "running"
	// TaskCompleted is a task whose sub-agent produced an answer.
	TaskCompleted TaskStatus = "completed"
	// TaskFailed is a task whose sub-agent failed or which could not even be started.
	TaskFailed TaskStatus = "failed"
	// TaskTimeout is a task which exceeded [DelegateConfig.SubTimeout].
	TaskTimeout TaskStatus = "timeout"
	// TaskCancelled is a task which was cancelled, e.g. because the parent run was stopped.
	TaskCancelled TaskStatus = "cancelled"
	// TaskLost is a background task id nobody knows (any more), e.g. after a server restart or once its result
	// expired, see [DefaultTaskTTL].
	TaskLost TaskStatus = "lost"
)

// TaskResult is what the model learns about one delegated task.
type TaskResult struct {
	// ID identifies a background task (see [NewTaskTools]). Empty for a task of the synchronous delegate tool.
	ID string `json:"id,omitempty"`
	// Title is the short label the model gave the task.
	Title string `json:"title"`
	// Status is the state of the task.
	Status TaskStatus `json:"status"`
	// Answer is the final answer of the sub-agent, set for [TaskCompleted].
	Answer string `json:"answer,omitempty"`
	// Error explains why the task did not complete.
	Error string `json:"error,omitempty"`
	// Usage is the token usage of the sub-agent.
	Usage Usage `json:"usage,omitzero"`
	// SessionID is the persisted child session holding the transcript of the sub-agent, if the [SubRunner]
	// persists one.
	SessionID string `json:"sessionId,omitempty"`
}

// SubRunRequest is everything a [SubRunner] needs to run one sub-agent.
type SubRunRequest struct {
	// Title is the short label of the task.
	Title string
	// ParentCallID is the id of the tool call which started the task.
	ParentCallID string
	// Completions runs the sub-agent. It retries rate limited requests on its own.
	Completions Completions
	// Options carries model, system prompt and token budget. Messages is empty; the first user turn is Input.
	Options Options
	// Input is the first and only user turn of the sub-agent: the task and its context.
	Input []Content
	// Tools are the tools of the sub-agent, already restricted as described at [DelegateConfig].
	Tools []Tool
	// MaxTurns bounds the loop of the sub-agent.
	MaxTurns int
	// FileUploader and OnBeforeToolCall are inherited from the parent run.
	FileUploader     FileUploader
	OnBeforeToolCall BeforeToolCallFunc
	// OnProgress reports the progress of the sub-agent. It is never the parent's callback.
	OnProgress ProgressFunc
}

// RunOptions assembles the [RunOptions] of the sub-agent. It never confirms mutating calls: a sub-agent has no
// user to ask, which is why it only gets mutating tools when that was explicitly allowed.
func (r SubRunRequest) RunOptions(ctx context.Context) RunOptions {
	opts := r.Options
	opts.Messages = []Message{{Role: User, Content: r.Input}}

	return RunOptions{
		Options:          opts,
		Tools:            r.Tools,
		MaxTurns:         r.MaxTurns,
		OnProgress:       r.OnProgress,
		FileUploader:     r.FileUploader,
		OnBeforeToolCall: r.OnBeforeToolCall,
		Context:          ctx,
	}
}

// SubRunResult is the outcome of one sub-agent.
type SubRunResult struct {
	// Answer is the final answer of the sub-agent.
	Answer string
	// History is the transcript of the sub-agent.
	History []Message
	// Usage is the token usage of the sub-agent. It should be set even when the run failed.
	Usage Usage
	// SessionID is the persisted child session, if any.
	SessionID string
}

// SubRunner runs one sub-agent to its end. ctx carries the timeout of the task and the cancellation of the
// parent. The default is [DefaultSubRunner], which keeps nothing; session.NewSubRunner persists every
// sub-agent as a child session of the parent session instead.
//
// A runner accounts the usage of the sub-agent through [RunOptions.OnUsage] of the run it starts: besides
// every completion of the sub-agent itself, the sub-agents it delegates to report their usage there, so
// [SubRunResult.Usage] covers the whole cost of the task.
type SubRunner func(ctx context.Context, subject auth.Subject, req SubRunRequest) (SubRunResult, error)

// DefaultSubRunner runs the sub-agent transiently via [Start]. A suspension is reported as
// [ErrSubRunSuspended].
func DefaultSubRunner(ctx context.Context, subject auth.Subject, req SubRunRequest) (SubRunResult, error) {
	var mu sync.Mutex
	var usage Usage
	opts := req.RunOptions(ctx)
	opts.OnUsage = func(u Usage) {
		mu.Lock()
		usage = usage.Add(u)
		mu.Unlock()
	}

	out, err := Start(subject, req.Completions, opts)

	mu.Lock()
	res := SubRunResult{History: out.History, Usage: usage}
	mu.Unlock()
	if err != nil {
		return res, err
	}

	if out.Suspended != nil {
		return res, ErrSubRunSuspended
	}

	res.Answer = FinalAnswer(out.History)
	return res, nil
}

// FinalAnswer returns the visible text of the last assistant turn of history which has any.
func FinalAnswer(history []Message) string {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role != Assistant {
			continue
		}
		if text := strings.TrimSpace(extractText(history[i])); text != "" {
			return text
		}
	}
	return ""
}

// SubEventKind classifies a [SubEvent].
type SubEventKind string

const (
	// SubStarted is reported when a task was accepted. It may still wait for a free slot.
	SubStarted SubEventKind = "started"
	// SubProgressed forwards a [Progress] event of the sub-agent.
	SubProgressed SubEventKind = "progress"
	// SubFinished is reported once the task reached its final status.
	SubFinished SubEventKind = "finished"
)

// SubEvent reports what a delegated task is doing, see [DelegateConfig.OnEvent].
type SubEvent struct {
	Kind SubEventKind
	// TaskID identifies the task: the background task id, or "<call id>/<index>" for the delegate tool.
	TaskID string
	// Title is the label of the task.
	Title string
	// Background is set for tasks of [NewTaskTools].
	Background bool
	// Progress is set for [SubProgressed].
	Progress *Progress
	// Result is set for [SubFinished].
	Result *TaskResult
}

// Limiter bounds the sub-agents of one run: how many work at the same time, and how many may be started at
// all. The delegate tool and the background task tools of the same run should share one, so the limits hold
// for both together; see [DelegateConfig.Limiter].
type Limiter struct {
	sem    chan struct{}
	budget *atomic.Int64
}

// NewLimiter creates a limiter admitting maxParallel concurrent sub-agents and maxTasks sub-agents in total.
// Values <= 0 fall back to [DefaultDelegateMaxParallel] and [DefaultDelegateMaxTasksPerRun].
func NewLimiter(maxParallel, maxTasks int) *Limiter {
	if maxParallel <= 0 {
		maxParallel = DefaultDelegateMaxParallel
	}
	if maxTasks <= 0 {
		maxTasks = DefaultDelegateMaxTasksPerRun
	}

	budget := &atomic.Int64{}
	budget.Store(int64(maxTasks))
	return &Limiter{sem: make(chan struct{}, maxParallel), budget: budget}
}

// Remaining returns how many more sub-agents may be started.
func (l *Limiter) Remaining() int {
	return int(max(l.budget.Load(), 0))
}

// reserve takes n tasks from the budget, all or nothing.
func (l *Limiter) reserve(n int) bool {
	for {
		cur := l.budget.Load()
		if cur < int64(n) {
			return false
		}
		if l.budget.CompareAndSwap(cur, cur-int64(n)) {
			return true
		}
	}
}

// acquire waits for a free slot.
func (l *Limiter) acquire(ctx context.Context) error {
	// a select picks randomly among ready cases, so a cancelled context must win explicitly
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limiter) release() {
	<-l.sem
}

// renewed returns a limiter with a fresh task budget which shares the slots of l.
func (l *Limiter) renewed(maxTasks int) *Limiter {
	if maxTasks <= 0 {
		maxTasks = DefaultDelegateMaxTasksPerRun
	}

	budget := &atomic.Int64{}
	budget.Store(int64(maxTasks))
	return &Limiter{sem: l.sem, budget: budget}
}

// nested returns a limiter for the sub-agents of a sub-agent. It shares the task budget but has its own slots:
// a sub-agent holds a slot while it waits for its own sub-agents, so sharing the slots could deadlock.
func (l *Limiter) nested() *Limiter {
	return &Limiter{sem: make(chan struct{}, cap(l.sem)), budget: l.budget}
}

// DelegateConfig configures [NewDelegateTool] and [NewTaskTools].
//
// A sub-agent gets the tools of the parent, reduced step by step, and never more:
//
//  1. Tools (or the tools of the calling run when Tools is nil),
//  2. without tools which await the user ([Tool.AwaitsUser]) or must not be delegated ([Tool.NoDelegate]),
//  3. without mutating tools, unless AllowMutating is set and neither ConfirmMutating nor the calling run
//     confirms mutations - a sub-agent would bypass the approval. If only marked tools are confirmed
//     (ConfirmMarked), the sub-agent gets the mutating tools which do not require approval,
//  4. intersected with AllowedTools, if set,
//  5. intersected with the tools the model requested for the task, if any. Requesting a tool outside of this
//     set fails the task rather than widening anything.
type DelegateConfig struct {
	// Completions runs the sub-agents. Nil means the provider of the calling run.
	Completions Completions

	// Model is the default model of the sub-agents. Empty means the model of the calling run.
	Model model.ID

	// Models is the allowlist of models the model may pick per task. Empty means it cannot pick any.
	Models []model.ID

	// System is appended to the fixed sub-agent preamble. Empty means the system prompt of the calling run.
	System string

	// Tools is the tool set the sub-agent tools are derived from. Nil means the tools of the calling run; a
	// non-nil empty slice means no tools at all.
	Tools []Tool

	// AllowedTools restricts the sub-agent tools to these names. Empty means no additional restriction.
	AllowedTools []string

	// AllowMutating lets sub-agents use [Tool.Mutating] tools. It has no effect when ConfirmMutating is set or
	// the calling run confirms mutations ([RunOptions.ConfirmMutating]), because a sub-agent cannot ask anyone
	// and would bypass the approval. When only marked tools are confirmed (ConfirmMarked or
	// [RunOptions.ConfirmMarked]), the tools marked [Tool.RequiresApproval] stay reserved for the calling run.
	// Default is read-only.
	AllowMutating bool

	// ConfirmMutating states that the parent confirms mutations. See AllowMutating.
	ConfirmMutating bool

	// ConfirmMarked states that the parent confirms the calls of tools marked [Tool.RequiresApproval]. See
	// AllowMutating.
	ConfirmMarked bool

	// OnBeforeToolCall is consulted before every tool call of a sub-agent. Nil means the one of the calling run.
	OnBeforeToolCall BeforeToolCallFunc

	// FileUploader is used by file tools of a sub-agent. Nil means the one of the calling run.
	FileUploader FileUploader

	// MaxDepth bounds nesting: the sub-agents of a run are at depth 1, and they receive a delegate tool of their
	// own only while their depth is below MaxDepth. Zero means [DefaultDelegateMaxDepth], so sub-agents cannot
	// delegate.
	MaxDepth int

	// MaxParallel, MaxTasksPerCall, MaxTasksPerRun, SubMaxTurns and SubTimeout are the limits; zero means the
	// respective Default* constant. MaxParallel and MaxTasksPerRun are ignored when Limiter is set.
	MaxParallel     int
	MaxTasksPerCall int
	MaxTasksPerRun  int
	SubMaxTurns     int
	SubTimeout      time.Duration

	// Limiter shares the limits between the delegate tool and the task tools of one run. Nil creates one per
	// constructed tool set. With background tasks, use [TaskGroup.Limiter], because the tasks outlive the run.
	Limiter *Limiter

	// Context additionally cancels the sub-agents of the synchronous delegate tool, besides the context of the
	// calling run. Background tasks are bound to their [TaskGroup] instead. Optional.
	Context context.Context

	// Runner runs one sub-agent. Nil means [DefaultSubRunner].
	Runner SubRunner

	// OnEvent observes the delegated tasks. It is called from the goroutines of the sub-agents and must be
	// goroutine-safe. Optional.
	OnEvent func(SubEvent)

	// OnUsage receives the usage of every finished sub-agent. It is called from the goroutines of the
	// sub-agents and must be goroutine-safe. Optional.
	OnUsage func(Usage)

	// depth is the depth of the calling run: 0 for the top level.
	depth int
}

func (cfg DelegateConfig) withDefaults() DelegateConfig {
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = DefaultDelegateMaxDepth
	}
	if cfg.MaxParallel <= 0 {
		cfg.MaxParallel = DefaultDelegateMaxParallel
	}
	if cfg.MaxTasksPerCall <= 0 {
		cfg.MaxTasksPerCall = DefaultDelegateMaxTasksPerCall
	}
	if cfg.MaxTasksPerRun <= 0 {
		cfg.MaxTasksPerRun = DefaultDelegateMaxTasksPerRun
	}
	if cfg.SubMaxTurns <= 0 {
		cfg.SubMaxTurns = DefaultDelegateSubMaxTurns
	}
	if cfg.SubTimeout <= 0 {
		cfg.SubTimeout = DefaultDelegateSubTimeout
	}
	if cfg.Limiter == nil {
		cfg.Limiter = NewLimiter(cfg.MaxParallel, cfg.MaxTasksPerRun)
	}
	if cfg.Runner == nil {
		cfg.Runner = DefaultSubRunner
	}
	return cfg
}

// subAgentPreamble is the fixed part of every sub-agent system prompt.
const subAgentPreamble = `You are a sub-agent. Another AI assistant delegated exactly one task to you and waits for your result.
- No user is present. Do not ask questions and do not wait for confirmation; decide sensibly on your own and state your assumptions.
- Solve exactly the given task, nothing more. Use your tools where they help.
- The caller sees only your final answer, not your tool calls or intermediate steps. Make the final answer concise, self-contained and complete: include every fact, number and identifier the caller needs.
- If the task cannot be solved, say so clearly and explain why.`

// delegateTaskIn is one task as the model describes it.
type delegateTaskIn struct {
	Title   string   `json:"title"`
	Task    string   `json:"task"`
	Context string   `json:"context,omitempty"`
	Tools   []string `json:"tools,omitempty"`
	Model   string   `json:"model,omitempty"`
}

type delegateIn struct {
	Tasks []delegateTaskIn `json:"tasks"`
}

type delegateOut struct {
	Results []TaskResult `json:"results"`
}

// delegator is the shared implementation of the delegation tools.
type delegator struct {
	cfg DelegateConfig
}

// mayMutate reports whether a sub-agent may end up with a mutating tool, as far as it is known at construction.
func (d *delegator) mayMutate() bool {
	if !d.cfg.AllowMutating || d.cfg.ConfirmMutating {
		return false
	}
	if d.cfg.Tools == nil {
		// inherited at call time; assume the worst
		return true
	}
	return slices.ContainsFunc(subTools(d.cfg.Tools, d.cfg.AllowedTools, true, !d.cfg.ConfirmMarked), func(t Tool) bool { return t.Mutating })
}

// subTools derives the tool set of a sub-agent, see [DelegateConfig]. allowMarked is only considered together
// with allowMutating and admits the mutating tools marked [Tool.RequiresApproval].
func subTools(base []Tool, allowed []string, allowMutating, allowMarked bool) []Tool {
	out := make([]Tool, 0, len(base))
	for _, t := range base {
		switch {
		case t.AwaitsUser, t.NoDelegate:
			continue
		case t.Mutating && !allowMutating:
			continue
		case t.Mutating && t.RequiresApproval && !allowMarked:
			continue
		case len(allowed) > 0 && !slices.Contains(allowed, t.Def.Name):
			continue
		}
		out = append(out, t)
	}
	return out
}

// taskSchema is the JSON schema of the task list shared by delegate and start_tasks.
func (d *delegator) taskSchema() map[string]any {
	toolsDesc := "optional names of the tools the sub-agent needs; omit to give it all tools available to sub-agents. Tools outside of that set are refused."
	if d.cfg.Tools != nil {
		var names []string
		for _, t := range subTools(d.cfg.Tools, d.cfg.AllowedTools, d.cfg.AllowMutating && !d.cfg.ConfirmMutating, !d.cfg.ConfirmMarked) {
			names = append(names, t.Def.Name)
		}
		if len(names) == 0 {
			toolsDesc = "must be omitted: sub-agents have no tools and can only reason."
		} else {
			toolsDesc += " Available: " + strings.Join(names, ", ") + "."
		}
	}

	modelDesc := "must be omitted: sub-agents always use the current model."
	if len(d.cfg.Models) > 0 {
		names := make([]string, 0, len(d.cfg.Models))
		for _, m := range d.cfg.Models {
			names = append(names, string(m))
		}
		modelDesc = "optional model for this task; omit to use the current model. Allowed: " + strings.Join(names, ", ") + "."
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tasks": map[string]any{
				"type":        "array",
				"description": fmt.Sprintf("the independent tasks, at most %d", d.cfg.MaxTasksPerCall),
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":   map[string]any{"type": "string", "description": "a short label of the task, shown to the user"},
						"task":    map[string]any{"type": "string", "description": "the exact assignment, phrased so it can be solved without any further question"},
						"context": map[string]any{"type": "string", "description": "everything the sub-agent needs to know: it does not see this conversation, only task and context"},
						"tools":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": toolsDesc},
						"model":   map[string]any{"type": "string", "description": modelDesc},
					},
					"required":             []string{"title", "task"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"tasks"},
		"additionalProperties": false,
	}
}

// NewDelegateTool creates the delegate tool: the model hands several independent tasks to sub-agents which
// work on them in parallel, and waits for all of them. Each task is answered with its status (completed,
// failed, timeout or cancelled) and the final answer of its sub-agent. The call is only reported as an error
// if the input is invalid or every task failed, so one failure keeps the results of the others.
//
// The sub-agents run without user interaction and with restricted tools, see [DelegateConfig]. Their usage is
// reported via [DelegateConfig.OnUsage], not in the [Outcome.Usage] of the calling run.
func NewDelegateTool(cfg DelegateConfig) Tool {
	d := &delegator{cfg: cfg.withDefaults()}

	desc := "Delegates independent tasks to sub-agents which work on them in parallel, and waits until all of them are done. " +
		"Use it to speed up work that splits into parts which do not depend on each other, e.g. researching several topics, " +
		"checking several records or drafting several sections. Do not delegate a task which needs the result of another one; " +
		"do those one after another. A sub-agent does not see this conversation and cannot ask anything: put everything it " +
		"needs into task and context. Each result contains the sub-agent's final answer; check it before you rely on it."
	if d.cfg.depth == 0 {
		desc += " If you want to continue working while the sub-agents run, use start_tasks instead (when available)."
	}

	return Tool{
		Def:              newToolDefFromSchema(DelegateToolName, desc, d.taskSchema()),
		NoDelegate:       true,
		indirectMutating: d.mayMutate(),
		run:              d.runDelegate,
	}
}

// taskPlan is a validated task, ready to run.
type taskPlan struct {
	title string
	req   SubRunRequest
	err   error
	// reportUsage accounts the usage of the task to the run which delegates it, see [SubRunner].
	reportUsage func(Usage)
}

// plan validates one task of the model and assembles its sub-agent.
func (d *delegator) plan(env toolEnv, in delegateTaskIn, callID string, index int) taskPlan {
	cfg := d.cfg
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = fmt.Sprintf("task %d", index+1)
	}

	p := taskPlan{title: title}
	fail := func(format string, args ...any) taskPlan {
		p.err = fmt.Errorf(format, args...)
		return p
	}

	if strings.TrimSpace(in.Task) == "" {
		return fail("the task must not be empty")
	}

	parent := env.opts
	if parent == nil {
		parent = &RunOptions{}
	}

	mdl := cfg.Model
	if mdl == "" {
		mdl = parent.Model
	}
	if in.Model != "" {
		if !slices.Contains(cfg.Models, model.ID(in.Model)) {
			return fail("model %q is not allowed for sub-agents", in.Model)
		}
		mdl = model.ID(in.Model)
	}

	comps := cfg.Completions
	if comps == nil {
		comps = env.completions
	}
	if comps == nil {
		return fail("no provider configured for sub-agents")
	}

	baseSystem := cfg.System
	if baseSystem == "" {
		// The calling run may itself be a sub-agent; never stack the preamble.
		baseSystem = strings.TrimSpace(strings.TrimPrefix(parent.System, subAgentPreamble))
	}
	system := strings.TrimSpace(subAgentPreamble + "\n\n" + baseSystem)

	base := cfg.Tools
	if base == nil {
		base = parent.Tools
	}
	confirm := cfg.ConfirmMutating || parent.ConfirmMutating
	confirmMarked := cfg.ConfirmMarked || parent.ConfirmMarked
	tools := subTools(base, cfg.AllowedTools, cfg.AllowMutating && !confirm, !confirmMarked)

	if len(in.Tools) > 0 {
		requested := make([]Tool, 0, len(in.Tools))
		for _, name := range in.Tools {
			idx := slices.IndexFunc(tools, func(t Tool) bool { return t.Def.Name == name })
			if idx < 0 {
				// the tool description is static, so it may offer a mutating tool which this run withholds
				if (confirm || confirmMarked) && slices.ContainsFunc(subTools(base, cfg.AllowedTools, cfg.AllowMutating, true), func(t Tool) bool { return t.Def.Name == name }) {
					return fail("tool %q changes data and is not available to sub-agents in this conversation, because changes need the approval of the user; do it yourself", name)
				}
				return fail("tool %q is not available to sub-agents", name)
			}
			if !slices.ContainsFunc(requested, func(t Tool) bool { return t.Def.Name == name }) {
				requested = append(requested, tools[idx])
			}
		}
		tools = requested
	}

	// A sub-agent may delegate on its own only while it is above the depth limit. Its sub-agents are
	// transient and not observed by the parent's UI. Their usage is accounted to the run of the sub-agent,
	// see reportUsage, which is why the nested tool reports nothing on its own.
	if d.cfg.depth+1 < cfg.MaxDepth {
		nested := cfg
		nested.depth = cfg.depth + 1
		nested.Tools = slices.Clone(tools)
		nested.Completions = comps
		nested.Model = mdl
		nested.System = baseSystem
		nested.Limiter = cfg.Limiter.nested()
		nested.Runner = DefaultSubRunner
		nested.OnEvent = nil
		nested.OnUsage = nil
		tools = append(tools, NewDelegateTool(nested))
	}

	// The usage of a nested sub-agent belongs to the run of the sub-agent which delegates it, see [SubRunner].
	// A top level run accounts its sub-agents through [DelegateConfig.OnUsage] instead, because
	// [RunOptions.OnUsage] of a conversation only covers its own completions, like [Outcome.Usage].
	if d.cfg.depth > 0 {
		p.reportUsage = parent.OnUsage
	}

	uploader := cfg.FileUploader
	if uploader == nil {
		uploader = parent.FileUploader
	}
	before := cfg.OnBeforeToolCall
	if before == nil {
		before = parent.OnBeforeToolCall
	}

	text := strings.TrimSpace(in.Task)
	if c := strings.TrimSpace(in.Context); c != "" {
		text += "\n\nContext:\n" + c
	}

	p.req = SubRunRequest{
		Title:        title,
		ParentCallID: callID,
		Completions:  &retryCompletions{Completions: comps, backoff: rateLimitBackoff},
		Options: Options{
			Model:          mdl,
			System:         system,
			MaxTokens:      parent.MaxTokens,
			Thinking:       parent.Thinking,
			ThinkingBudget: parent.ThinkingBudget,
		},
		Input:            []Content{Text{Text: text}},
		Tools:            tools,
		MaxTurns:         cfg.SubMaxTurns,
		FileUploader:     uploader,
		OnBeforeToolCall: before,
	}

	return p
}

// planAll validates the tasks of a call and reserves the budget for the valid ones.
func (d *delegator) planAll(env toolEnv, args json.RawMessage, callID string) ([]taskPlan, error) {
	var in delegateIn
	if err := json.Unmarshal(args, &in); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if len(in.Tasks) == 0 {
		return nil, fmt.Errorf("tasks must not be empty")
	}

	if len(in.Tasks) > d.cfg.MaxTasksPerCall {
		return nil, fmt.Errorf("too many tasks: at most %d per call", d.cfg.MaxTasksPerCall)
	}

	plans := make([]taskPlan, len(in.Tasks))
	valid := 0
	for i, t := range in.Tasks {
		plans[i] = d.plan(env, t, callID, i)
		if plans[i].err == nil {
			valid++
		}
	}

	if valid > 0 && !d.cfg.Limiter.reserve(valid) {
		return nil, fmt.Errorf("the sub-agent budget of this run is exhausted: %d more may be started, %d requested; do the remaining work yourself", d.cfg.Limiter.Remaining(), valid)
	}

	return plans, nil
}

// runDelegate is the synchronous fan-out of the delegate tool.
func (d *delegator) runDelegate(env toolEnv, call ToolCall) ToolResult {
	plans, err := d.planAll(env, call.Arguments, call.ID)
	if err != nil {
		return errorResult(err)
	}

	parent := env.ctx
	if parent == nil {
		parent = context.Background()
	}
	if d.cfg.Context != nil {
		var cancel context.CancelFunc
		parent, cancel = context.WithCancel(parent)
		defer cancel()
		stop := context.AfterFunc(d.cfg.Context, cancel)
		defer stop()
	}

	results := make([]TaskResult, len(plans))
	var wg sync.WaitGroup
	for i, p := range plans {
		taskID := fmt.Sprintf("%s/%d", call.ID, i)
		if p.err != nil {
			results[i] = TaskResult{Title: p.title, Status: TaskFailed, Error: p.err.Error()}
			d.emit(SubEvent{Kind: SubStarted, TaskID: taskID, Title: p.title})
			d.emit(SubEvent{Kind: SubFinished, TaskID: taskID, Title: p.title, Result: &results[i]})
			continue
		}

		wg.Go(func() {
			results[i] = d.execute(parent, env.subject, p, taskID, false)
		})
	}
	wg.Wait()

	allFailed := !slices.ContainsFunc(results, func(r TaskResult) bool { return r.Status == TaskCompleted })

	buf, err := json.Marshal(delegateOut{Results: results})
	if err != nil {
		return errorResult(err)
	}

	return ToolResult{IsError: allFailed, Content: []Content{Text{Text: string(buf)}}}
}

// execute runs one planned task to its end and never panics.
func (d *delegator) execute(parent context.Context, subject auth.Subject, p taskPlan, taskID string, background bool) (result TaskResult) {
	result = TaskResult{Title: p.title}
	if background {
		result.ID = taskID
	}

	d.emit(SubEvent{Kind: SubStarted, TaskID: taskID, Title: p.title, Background: background})
	defer func() {
		if r := recover(); r != nil {
			slog.Error("sub-agent panicked", "task", taskID, "panic", r, "stack", string(debug.Stack()))
			result.Status = TaskFailed
			result.Error = fmt.Sprintf("the sub-agent crashed: %v", r)
		}
		d.emit(SubEvent{Kind: SubFinished, TaskID: taskID, Title: p.title, Background: background, Result: &result})
	}()

	if err := d.cfg.Limiter.acquire(parent); err != nil {
		result.Status = TaskCancelled
		result.Error = "cancelled before the sub-agent started"
		return result
	}
	defer d.cfg.Limiter.release()

	ctx, cancel := context.WithTimeout(parent, d.cfg.SubTimeout)
	defer cancel()

	req := p.req
	req.OnProgress = func(pr Progress) {
		d.emit(SubEvent{Kind: SubProgressed, TaskID: taskID, Title: p.title, Background: background, Progress: &pr})
	}

	res, err := d.cfg.Runner(ctx, subject, req)
	result.Usage = res.Usage
	result.SessionID = res.SessionID
	if !res.Usage.IsZero() {
		if d.cfg.OnUsage != nil {
			d.cfg.OnUsage(res.Usage)
		}
		if p.reportUsage != nil {
			p.reportUsage(res.Usage)
		}
	}

	switch {
	case err == nil:
		result.Status = TaskCompleted
		result.Answer = res.Answer
		if strings.TrimSpace(result.Answer) == "" {
			result.Answer = "(the sub-agent gave no answer)"
		}
	case parent.Err() != nil:
		result.Status = TaskCancelled
		result.Error = "cancelled: the calling run was stopped"
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		result.Status = TaskTimeout
		result.Error = fmt.Sprintf("the sub-agent did not finish within %s", d.cfg.SubTimeout)
	default:
		result.Status = TaskFailed
		result.Error = toolErrorText(err)
	}

	return result
}

// emit reports e to [DelegateConfig.OnEvent]. The callback belongs to the caller, e.g. a UI, and runs on the
// goroutine of a sub-agent, where a panic would crash the whole process. Thus, it is recovered and logged.
func (d *delegator) emit(e SubEvent) {
	if d.cfg.OnEvent == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			slog.Error("OnEvent of a sub-agent panicked", "task", e.TaskID, "kind", e.Kind, "panic", r, "stack", string(debug.Stack()))
		}
	}()

	d.cfg.OnEvent(e)
}

// errorResult is a tool error with the given cause.
func errorResult(err error) ToolResult {
	return ToolResult{IsError: true, Content: []Content{Text{Text: toolErrorText(err)}}}
}

// rateLimitBackoff are the pauses between the attempts of a rate limited sub-agent request.
var rateLimitBackoff = []time.Duration{time.Second, 3 * time.Second, 8 * time.Second}

// retryCompletions retries requests which failed with [TooManyRequests]. Parallel sub-agents hit the rate limit
// of a provider far more easily than a single conversation, and failing a whole sub-agent on that would waste
// what it already did.
type retryCompletions struct {
	Completions
	backoff []time.Duration
}

func (r *retryCompletions) Complete(ctx context.Context, subject auth.Subject, opts Options) (Result, error) {
	for attempt := 0; ; attempt++ {
		res, err := r.Completions.Complete(ctx, subject, opts)
		if err == nil || !errors.Is(err, TooManyRequests) || attempt >= len(r.backoff) {
			return res, err
		}

		if err := r.wait(ctx, attempt, err); err != nil {
			return res, err
		}
	}
}

// Stream retries like Complete, but only as long as nothing has been delivered yet.
func (r *retryCompletions) Stream(ctx context.Context, subject auth.Subject, opts Options) iter.Seq2[Delta, error] {
	return func(yield func(Delta, error) bool) {
		for attempt := 0; ; attempt++ {
			delivered := false
			var retry error
			for d, err := range r.Completions.Stream(ctx, subject, opts) {
				if err != nil && !delivered && errors.Is(err, TooManyRequests) && attempt < len(r.backoff) {
					retry = err
					break
				}

				delivered = true
				if !yield(d, err) {
					return
				}
			}

			if retry == nil {
				return
			}

			if err := r.wait(ctx, attempt, retry); err != nil {
				yield(Delta{}, err)
				return
			}
		}
	}
}

// wait pauses before the next attempt. A cancelled context is reported as such, not as the rate limit.
func (r *retryCompletions) wait(ctx context.Context, attempt int, cause error) error {
	timer := time.NewTimer(r.backoff[attempt])
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%w (while waiting to retry: %w)", ctx.Err(), cause)
	}
}

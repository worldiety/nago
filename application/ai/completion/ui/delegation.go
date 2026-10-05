// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accordion"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// DelegationOptions lets the model of a [Chat] hand independent tasks to sub-agents, which work on them in
// parallel and without user interaction (see [completion.NewDelegateTool]). With BackgroundTasks the model may
// also start them in the background and continue meanwhile (see [completion.NewTaskTools]).
//
// Sub-agents get the tools of the agent, minus the ones which need the user (ask_user, [ScreenTool] and
// everything marked [completion.Tool.NoDelegate]) and minus the mutating ones, unless AllowMutating is set. With
// [ChatOptions.ConfirmMutations] or [ChatOptions.ReadOnly] they are always read-only, because a sub-agent cannot
// ask anyone for approval.
//
// With [ChatOptions.History] each sub-agent is persisted as a child session, which the user can open from the
// conversation; the history dialog does not list them.
type DelegationOptions struct {
	// Models the model may pick per task. Empty means sub-agents always use the model of the agent.
	Models []model.ID

	// System replaces the system prompt of the agent for sub-agents. A fixed sub-agent preamble is always added.
	// Optional.
	System string

	// AllowedTools restricts the tools of sub-agents to these names. Empty means every tool that may be delegated.
	AllowedTools []string

	// AllowMutating lets sub-agents use mutating tools. Ignored with ConfirmMutations or ReadOnly.
	AllowMutating bool

	// BackgroundTasks additionally offers start_tasks, await_tasks and cancel_tasks.
	BackgroundTasks bool

	// MaxDepth, MaxParallel, MaxTasksPerCall, MaxTasksPerRun, SubMaxTurns and SubTimeout are the limits, see
	// [completion.DelegateConfig]. Zero means the respective default.
	MaxDepth        int
	MaxParallel     int
	MaxTasksPerCall int
	MaxTasksPerRun  int
	SubMaxTurns     int
	SubTimeout      time.Duration
}

// fallbackTasks keeps the background tasks of chats whose [ChatOptions.Sessions] carry no task registry, e.g.
// transient chats wired with a hand-made session.UseCases.
var fallbackTasks = completion.NewTaskRegistry()

// taskRegistry returns the registry for background tasks of a chat.
func taskRegistry(sessions session.UseCases) *completion.TaskRegistry {
	if sessions.Tasks != nil {
		return sessions.Tasks
	}
	return fallbackTasks
}

// delegationRun is what the delegation tools of one run need from the chat.
type delegationRun struct {
	opts    ChatOptions
	model   model.ID
	system  string
	tools   []completion.Tool
	confirm bool
	// confirmMarked states that the run holds the calls of tools which require approval
	confirmMarked bool
	fileUploader  completion.FileUploader
	sessionID     session.ID
	group         *completion.TaskGroup
	// renew states that this is a new run rather than the continuation of a suspended one, see
	// [completion.TaskGroup.Limiter].
	renew   bool
	onEvent func(completion.SubEvent)
	// onUsage receives the usage of every sub-agent, see [completion.DelegateConfig.OnUsage]. Nil with
	// History, because the session books it then.
	onUsage func(completion.Usage)
}

// delegationTools builds the delegation tools of one run and, with background tasks, the hook which joins
// forgotten tasks before the run ends. It returns nothing when delegation is off.
func delegationTools(r delegationRun) ([]completion.Tool, func(ctx context.Context) string) {
	d := r.opts.Delegation
	if d == nil {
		return nil, nil
	}

	cfg := completion.DelegateConfig{
		Completions:     r.opts.Completions,
		Model:           r.model,
		Models:          d.Models,
		System:          d.System,
		Tools:           slices.Clone(r.tools),
		AllowedTools:    d.AllowedTools,
		AllowMutating:   d.AllowMutating && !r.opts.ReadOnly,
		ConfirmMutating: r.confirm || r.opts.ConfirmMutations,
		ConfirmMarked:   r.confirmMarked || r.opts.ConfirmMarked,
		FileUploader:    r.fileUploader,
		MaxDepth:        d.MaxDepth,
		MaxParallel:     d.MaxParallel,
		MaxTasksPerCall: d.MaxTasksPerCall,
		MaxTasksPerRun:  d.MaxTasksPerRun,
		SubMaxTurns:     d.SubMaxTurns,
		SubTimeout:      d.SubTimeout,
		OnEvent:         r.onEvent,
		OnUsage:         r.onUsage,
	}

	// The limits hold for the whole run, even across a suspension, and background tasks even outlive the run,
	// thus they are kept by the group of the conversation.
	if r.group != nil {
		cfg.Limiter = r.group.Limiter(d.MaxParallel, d.MaxTasksPerRun, r.renew)
	} else {
		cfg.Limiter = completion.NewLimiter(d.MaxParallel, d.MaxTasksPerRun)
	}

	if r.opts.History && r.sessionID != "" {
		cfg.Runner = session.NewSubRunner(r.opts.Sessions, r.sessionID)
	}

	tools := []completion.Tool{completion.NewDelegateTool(cfg)}
	if !d.BackgroundTasks || r.group == nil {
		return tools, nil
	}

	tools = append(tools, completion.NewTaskTools(cfg, r.group)...)
	return tools, r.group.BeforeFinish(d.SubTimeout)
}

// taskProgressLabel is the status line while sub-agents work.
func taskProgressLabel(finished, started int) string {
	return fmt.Sprintf("… %d von %d Teilaufgaben erledigt", finished, started)
}

// tasksOf returns the tasks of a delegation call: the results once they exist, otherwise the tasks the model
// asked for, still running. A call which failed as a whole, e.g. because the budget was exhausted, shows its
// tasks as failed with the reason. latest holds the newest known state of the background tasks by id, see
// latestTasks, which replaces the state a task had when it was started.
func tasksOf(call completion.ToolCall, res completion.ToolResult, done bool, latest map[string]completion.TaskResult) []completion.TaskResult {
	var tasks []completion.TaskResult
	if done {
		if parsed, ok := completion.ParseTaskResults(res); ok {
			tasks = parsed
		}
	}

	if tasks == nil {
		var in struct {
			Tasks []struct {
				Title string `json:"title"`
			} `json:"tasks"`
			IDs []string `json:"ids"`
		}
		if err := json.Unmarshal(call.Arguments, &in); err != nil {
			return nil
		}

		status := completion.TaskRunning
		reason := ""
		if done {
			// the call is over, but it produced no tasks: it has been refused as a whole
			status = completion.TaskFailed
			reason = strings.TrimSpace(extractText(res.Content))
		}

		for _, t := range in.Tasks {
			tasks = append(tasks, completion.TaskResult{Title: t.Title, Status: status, Error: reason})
		}
		for _, id := range in.IDs {
			tasks = append(tasks, completion.TaskResult{ID: id, Status: status, Error: reason})
		}
	}

	for i, t := range tasks {
		if l, ok := latest[t.ID]; ok && t.ID != "" && t.Status == completion.TaskRunning {
			if l.Title == "" {
				l.Title = t.Title
			}
			tasks[i] = l
		}
	}

	return tasks
}

// extractText joins the texts of the given contents.
func extractText(contents []completion.Content) string {
	var sb strings.Builder
	for _, c := range contents {
		if t, ok := c.(completion.Text); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String()
}

// latestTasks collects the newest known state of every background task of the history by id: from the results
// of start_tasks, await_tasks and cancel_tasks, and from the results which [completion.TaskGroup.BeforeFinish]
// handed over in a hidden prompt. Without that, a task which the model never awaited would stay "running"
// forever, and its transcript could not be opened.
func latestTasks(history []completion.Message) map[string]completion.TaskResult {
	latest := map[string]completion.TaskResult{}
	for _, m := range history {
		if tasks, ok := completion.HandedOverTasks(m); ok {
			for _, t := range tasks {
				if t.ID != "" {
					latest[t.ID] = t
				}
			}
			continue
		}

		for _, c := range m.Content {
			r, ok := c.(completion.ToolResult)
			if !ok {
				continue
			}
			tasks, ok := completion.ParseTaskResults(r)
			if !ok {
				continue
			}
			for _, t := range tasks {
				if t.ID == "" {
					continue
				}
				// a lost task tells nothing new about its state
				if cur, known := latest[t.ID]; known && t.Status == completion.TaskLost && cur.Status != completion.TaskRunning {
					continue
				}
				latest[t.ID] = t
			}
		}
	}

	return latest
}

// tasksHeading is the label of the sub task list of a delegation call.
func tasksHeading(name string, n int) string {
	switch name {
	case completion.StartTasksToolName:
		return fmt.Sprintf("Teilaufgaben gestartet (%d)", n)
	case completion.AwaitTasksToolName:
		return fmt.Sprintf("Ergebnisse der Teilaufgaben (%d)", n)
	case completion.CancelTasksToolName:
		return fmt.Sprintf("Teilaufgaben abgebrochen (%d)", n)
	default:
		return fmt.Sprintf("Teilaufgaben (%d)", n)
	}
}

// taskStatusLabel is the German label of a task status.
func taskStatusLabel(s completion.TaskStatus) string {
	switch s {
	case completion.TaskRunning:
		return "läuft"
	case completion.TaskCompleted:
		return "erledigt"
	case completion.TaskFailed:
		return "fehlgeschlagen"
	case completion.TaskTimeout:
		return "Zeit überschritten"
	case completion.TaskCancelled:
		return "abgebrochen"
	case completion.TaskLost:
		return "verloren"
	default:
		return string(s)
	}
}

func taskStatusIcon(s completion.TaskStatus) core.SVG {
	switch s {
	case completion.TaskRunning:
		return icons.Clock
	case completion.TaskCompleted:
		return icons.CheckCircle
	case completion.TaskTimeout:
		return icons.Hourglass
	case completion.TaskCancelled:
		return icons.Ban
	case completion.TaskLost:
		return icons.ExclamationCircle
	default:
		return icons.CloseCircle
	}
}

// tasksView renders a delegation call as a collapsible list of its sub tasks. Returns nil when the call carries
// no recognizable tasks, so the caller falls back to a plain tool hint.
func tasksView(wnd core.Window, hv historyView, call completion.ToolCall, res completion.ToolResult, done bool, latest map[string]completion.TaskResult) core.View {
	tasks := tasksOf(call, res, done, latest)
	if len(tasks) == 0 {
		return nil
	}

	rows := make([]core.View, 0, len(tasks))
	for _, t := range tasks {
		rows = append(rows, taskRow(hv, t))
	}

	header := ui.Text("→ " + tasksHeading(call.Name, len(tasks))).Font(ui.Small)
	body := ui.VStack(rows...).Gap(ui.L8).FullWidth().Alignment(ui.Leading)

	var box core.View
	if wnd == nil {
		box = ui.VStack(header, body).Alignment(ui.Leading)
	} else {
		open := core.StateOf[bool](wnd, hv.idPrefix+"uicompletion-tasks-"+call.ID)
		box = accordion.Accordion(header, body, open).Small().HideSeparator().Frame(ui.Frame{MaxWidth: "85%"})
	}

	return ui.HStack(box, ui.Spacer()).FullWidth()
}

// taskRow renders one sub task: status, title, a button to its transcript and its answer or error.
func taskRow(hv historyView, t completion.TaskResult) core.View {
	title := t.Title
	if title == "" {
		title = t.ID
	}

	var open core.View
	if hv.openChild != nil && t.SessionID != "" {
		sid := session.ID(t.SessionID)
		open = ui.TertiaryButton(func() { hv.openChild(sid) }).PreIcon(icons.Eye).AccessibilityLabel("Verlauf der Teilaufgabe anzeigen")
	}

	var detail core.View
	switch {
	case t.Status == completion.TaskCompleted && strings.TrimSpace(t.Answer) != "":
		detail = ui.VStack(markdown.RichText(t.Answer)).Alignment(ui.Leading).Font(ui.Small)
	case t.Error != "":
		detail = ui.Text(t.Error).Font(ui.Small)
	}

	return ui.VStack(
		ui.HStack(
			ui.ImageIcon(taskStatusIcon(t.Status)).AccessibilityLabel(taskStatusLabel(t.Status)),
			ui.Text(title).Font(ui.Small),
			ui.Text("• "+taskStatusLabel(t.Status)).Font(ui.Small),
			ui.Spacer(),
			ui.If(open != nil, open),
		).Gap(ui.L4).FullWidth().Alignment(ui.Center),
		ui.If(detail != nil, detail),
	).Gap(ui.L4).
		FullWidth().
		Alignment(ui.Leading).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))
}

// childDialog shows the persisted transcript of a sub-agent read-only while present is set.
func childDialog(wnd core.Window, sessions session.UseCases, childID *core.State[session.ID], present *core.State[bool]) core.View {
	id := childID.Get()
	if !present.Get() || id == "" {
		return nil
	}

	child, ok := reloadSession(wnd.Subject(), sessions, id)
	if !ok {
		return alert.Dialog("Teilaufgabe", ui.Text("Der Verlauf dieser Teilaufgabe ist nicht (mehr) verfügbar.").Font(ui.BodySmall), present, alert.Closeable())
	}

	title := child.Title
	if title == "" {
		title = "Teilaufgabe"
	}

	body := conversationView(wnd, child.Messages, "Die Teilaufgabe hat noch keinen Verlauf.", ui.L400, historyView{
		idPrefix: "child-" + string(id) + "-",
		// A sub-agent may have delegated on its own; its children open in the same dialog.
		openChild: func(next session.ID) {
			childID.Set(next)
		},
	})

	return alert.Dialog(title, body, present, alert.Closeable(), alert.Larger())
}

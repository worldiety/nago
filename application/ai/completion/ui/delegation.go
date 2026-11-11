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
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/progress"
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

// tasksHeading is the label of the tasks of a delegation call.
func tasksHeading(name string, n int) string {
	tasks := fmt.Sprintf("%d Hintergrundaufgaben", n)
	if n == 1 {
		tasks = "1 Hintergrundaufgabe"
	}

	switch name {
	case completion.StartTasksToolName:
		return tasks + " gestartet"
	case completion.AwaitTasksToolName:
		if n == 1 {
			return "Ergebnis einer Hintergrundaufgabe"
		}
		return fmt.Sprintf("Ergebnisse von %d Hintergrundaufgaben", n)
	case completion.CancelTasksToolName:
		return tasks + " abgebrochen"
	default:
		return tasks
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
		return "Zeitlimit erreicht"
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

func taskStatusColor(s completion.TaskStatus) ui.Color {
	switch s {
	case completion.TaskCompleted:
		return ui.ColorSemanticGood
	case completion.TaskTimeout:
		return ui.ColorSemanticWarn
	case completion.TaskFailed:
		return ui.ColorSemanticError
	default:
		return ui.M7
	}
}

// shownTask is a task of a delegation call with what the chat knows about its sub-agent, see taskBoard.
type shownTask struct {
	completion.TaskResult
	live    liveTask
	hasLive bool
}

// status is the state of the task. The sub-agent of a task of the delegate tool is done before the call, which
// reports the results of all of them.
func (t shownTask) status() completion.TaskStatus {
	if t.Status == completion.TaskRunning && t.hasLive && t.live.Status != completion.TaskRunning {
		return t.live.Status
	}
	return t.Status
}

// working tells, whether the sub-agent of the task works right now, rather than waiting for a free slot.
func (t shownTask) working() bool {
	return t.status() == completion.TaskRunning && t.hasLive && !t.live.Waiting
}

// taskKey is the id, under which the events of the i-th task of a call report it, see
// [completion.SubEvent.TaskID]: the id of a background task, or the call and the index for the delegate tool.
func taskKey(call completion.ToolCall, i int, t completion.TaskResult) string {
	if t.ID != "" {
		return t.ID
	}
	if call.Name == completion.DelegateToolName {
		return fmt.Sprintf("%s/%d", call.ID, i)
	}
	return ""
}

// The look of the task list: rows like cards with a hairline on the background of the chat, the title a little larger
// than what tells about it.
var (
	tasksHeadingFont = ui.Font{Name: ui.DefaultFontName, Size: "0.8125rem", Weight: 500, LineHeight: "1.125rem"}
	taskTitleFont    = ui.Font{Name: ui.DefaultFontName, Size: "0.8125rem", Weight: ui.BodyFontWeight, LineHeight: "1.125rem"}
	taskTextFont     = ui.BodySmall
	taskRowGap       = ui.Length("0.375rem")
	taskBarHeight    = ui.Length("0.1875rem")
)

// tasksView renders a delegation call as a collapsible list of its tasks: a heading with how many are done and
// since when they work, and a row for each task with what its sub-agent does right now. Returns nil when the call
// carries no recognizable tasks, so the caller falls back to a plain tool hint.
func tasksView(wnd core.Window, hv historyView, call completion.ToolCall, res completion.ToolResult, done bool, latest map[string]completion.TaskResult) core.View {
	results := tasksOf(call, res, done, latest)
	if len(results) == 0 {
		return nil
	}

	tasks := make([]shownTask, len(results))
	for i, r := range results {
		tasks[i] = shownTask{TaskResult: r}
		if key := taskKey(call, i, r); key != "" {
			tasks[i].live, tasks[i].hasLive = hv.live[key]
		}
	}

	// the list is open while its tasks work, and closed once they are done when it shows up first, e.g. in the history
	id := hv.idPrefix + "uicompletion-tasks-" + call.ID
	open := true
	var toggle func()
	if wnd != nil {
		running := slices.ContainsFunc(tasks, func(t shownTask) bool { return t.status() == completion.TaskRunning })
		state := core.StateOf[bool](wnd, id).Init(func() bool { return running })
		open = state.Get()
		toggle = func() { state.Set(!state.Get()) }
	}

	meta := tasksMeta(tasks, hv.now)
	header := ui.HStack(
		ui.ImageIcon(chevron(open)).StrokeColor(ui.M7).Frame(ui.Frame{}.Size(ui.L12, ui.L12)),
		ui.Text(tasksHeading(call.Name, len(tasks))).Font(tasksHeadingFont),
		ui.If(meta != "", ui.Text("· "+meta).Font(taskTextFont).Color(ui.M7)),
	).Gap(taskRowGap).Alignment(ui.Center)
	if toggle != nil {
		// the key keeps a click, which refers to a former tree: the window renders every second while tasks work
		header = header.Action(toggle).Key("uicompletion-tasks", id)
	}

	lines := []core.View{header}
	if open {
		for i, t := range tasks {
			lines = append(lines, taskRow(wnd, hv, fmt.Sprintf("%suicompletion-task-%s-%d", hv.idPrefix, call.ID, i), t))
		}
	}

	return ui.HStack(
		ui.VStack(lines...).Gap(taskRowGap).Alignment(ui.Leading).Frame(ui.Frame{Width: ui.Full, MaxWidth: "88%"}),
		ui.Spacer(),
	).FullWidth()
}

func chevron(open bool) core.SVG {
	if open {
		return icons.ChevronUp
	}
	return icons.ChevronDown
}

// tasksMeta summarizes the tasks of a call: how many are done and since when they work, or how long they took.
func tasksMeta(tasks []shownTask, now time.Time) string {
	var completed, other, running int
	var first, last time.Time
	timed := true
	for _, t := range tasks {
		switch t.status() {
		case completion.TaskRunning:
			running++
		case completion.TaskCompleted:
			completed++
		default:
			other++
		}

		if !t.hasLive {
			timed = false
			continue
		}
		if first.IsZero() || t.live.Queued.Before(first) {
			first = t.live.Queued
		}
		if t.live.Ended.After(last) {
			last = t.live.Ended
		}
	}

	var parts []string
	if completed > 0 {
		parts = append(parts, fmt.Sprintf("%d erledigt", completed))
	}
	if other > 0 {
		parts = append(parts, fmt.Sprintf("%d nicht erledigt", other))
	}
	if timed && !first.IsZero() {
		if running > 0 {
			parts = append(parts, "seit "+clock(now.Sub(first)))
		} else if !last.IsZero() {
			parts = append(parts, "in "+clock(last.Sub(first)))
		}
	}

	return strings.Join(parts, " · ")
}

// taskRow renders one task as a card: status, title and how long it works or worked, the button to its transcript,
// and what its sub-agent does right now. A click opens its latest steps and its time limit, or its answer or error.
// The card is no accordion, because an accordion within the body of another one takes its styles, and the detail
// is only rendered while it is open, because the window renders every second while tasks work.
func taskRow(wnd core.Window, hv historyView, id string, t shownTask) core.View {
	title := t.Title
	if title == "" {
		title = t.ID
	}

	detail := taskDetail(hv, t)
	open := wnd == nil
	var state *core.State[bool]
	if wnd != nil && detail != nil {
		state = core.StateOf[bool](wnd, id)
		open = state.Get()
	}

	var transcript core.View
	if hv.openChild != nil && t.SessionID != "" {
		sid := session.ID(t.SessionID)
		transcript = ui.TertiaryButton(func() { hv.openChild(sid) }).
			PreIcon(icons.Eye).
			Key("uicompletion-task-transcript", t.SessionID).
			AccessibilityLabel("Verlauf der Hintergrundaufgabe anzeigen")
	}

	head := ui.HStack(
		taskIcon(t),
		ui.Text(title).Font(taskTitleFont),
		ui.Spacer(),
		ui.Text(taskMeta(t, hv.now)).Font(taskTextFont).Color(ui.M7).WhiteSpace(ui.WhiteSpaceNoWrap),
		ui.If(transcript != nil, transcript),
		ui.If(state != nil, ui.ImageIcon(chevron(open)).StrokeColor(ui.M7).Frame(ui.Frame{}.Size(ui.L12, ui.L12))),
	).Gap(ui.L8).FullWidth().Alignment(ui.Center)

	summary := []core.View{head}
	if activity := taskActivity(t); activity != "" {
		summary = append(summary, ui.Text(activity).Font(taskTextFont).Color(ui.M7).Padding(ui.Padding{Left: ui.L24}))
	}

	top := ui.VStack(summary...).Gap(ui.L2).FullWidth().Alignment(ui.Leading)
	if state != nil {
		top = top.Action(func() { state.Set(!state.Get()) }).Key("uicompletion-task", id)
	}

	lines := []core.View{top}
	if open && detail != nil {
		lines = append(lines, detail)
	}

	return ui.VStack(lines...).
		Gap(taskRowGap).
		FullWidth().
		Alignment(ui.Leading).
		BackgroundColor(ui.M1).
		Border(ui.Border{}.Radius(ui.L8).Width(ui.L1).Color(ui.M3)).
		Padding(ui.Padding{Top: ui.L8, Bottom: ui.L8, Left: "0.625rem", Right: "0.625rem"})
}

// taskIcon is a spinning ring while the sub-agent of the task works, otherwise the icon of its status.
func taskIcon(t shownTask) core.View {
	status := t.status()
	if t.working() {
		return ui.VStack(spinner(ui.L16)).AccessibilityLabel(taskStatusLabel(status))
	}

	return ui.ImageIcon(taskStatusIcon(status)).
		StrokeColor(taskStatusColor(status)).
		Frame(ui.Frame{}.Size(ui.L16, ui.L16)).
		AccessibilityLabel(taskStatusLabel(status))
}

// taskMeta tells how long a task works and in which step, or how it ended, e.g. "2:14 · Schritt 6".
func taskMeta(t shownTask, now time.Time) string {
	l := t.live
	switch status := t.status(); {
	case status == completion.TaskRunning && t.hasLive && l.Waiting:
		return "wartet auf einen freien Platz"
	case status == completion.TaskRunning && t.hasLive:
		meta := clock(l.elapsed(now))
		if l.Turn > 0 {
			meta += fmt.Sprintf(" · Schritt %d", l.Turn)
		}
		return meta
	case status == completion.TaskCompleted && t.hasLive && !l.Started.IsZero():
		return "erledigt nach " + clock(l.elapsed(now))
	case status == completion.TaskTimeout && t.hasLive && l.Timeout > 0:
		return "Zeitlimit von " + clock(l.Timeout) + " erreicht"
	default:
		return taskStatusLabel(status)
	}
}

// taskActivity tells what the sub-agent of a working task does right now, e.g. "führt save_view aus".
func taskActivity(t shownTask) string {
	if !t.working() {
		return ""
	}
	return t.live.Activity
}

// taskDetail shows the latest steps and the time limit of a working sub-agent, or the answer or error of a finished
// one. Returns nil when there is nothing to show.
func taskDetail(hv historyView, t shownTask) core.View {
	var lines []core.View
	l := t.live
	if t.working() {
		if len(l.Steps) > 0 {
			steps := make([]core.View, 0, len(l.Steps))
			for _, s := range l.Steps {
				steps = append(steps, stepView(s))
			}
			// a line on the left, like a quote of what the sub-agent did
			lines = append(lines, ui.VStack(steps...).
				Gap(ui.L2).
				FullWidth().
				Alignment(ui.Leading).
				Border(ui.Border{LeftWidth: ui.L2, LeftColor: ui.M3}).
				Padding(ui.Padding{Left: "0.625rem"}))
		}

		if l.Timeout > 0 {
			elapsed := l.elapsed(hv.now)
			lines = append(lines, ui.HStack(
				progress.LinearProgress().
					Progress(float64(elapsed)/float64(l.Timeout)).
					Color(ui.M7).
					BackgroundColor(ui.M3).
					Frame(ui.Frame{Height: taskBarHeight, Width: ui.Full}),
				ui.Text("Zeitlimit "+clock(l.Timeout)).Font(taskTextFont).Color(ui.M7).WhiteSpace(ui.WhiteSpaceNoWrap),
			).Gap(ui.L8).FullWidth().Alignment(ui.Center))
		}
	}

	switch status := t.status(); {
	case status == completion.TaskCompleted && strings.TrimSpace(t.Answer) != "":
		lines = append(lines, ui.VStack(modelText(t.Answer)).Alignment(ui.Leading).Font(taskTextFont))
	case status == completion.TaskTimeout:
		lines = append(lines, ui.Text("Der Unteragent wurde nach Ablauf seines Zeitlimits beendet. Was er bis dahin getan hat, steht im Verlauf.").Font(taskTextFont))
	case status == completion.TaskRunning || status == completion.TaskCancelled:
	case t.Error != "":
		lines = append(lines, ui.Text(t.Error).Font(taskTextFont))
	}

	if len(lines) == 0 {
		return nil
	}

	return ui.VStack(lines...).
		Gap(taskRowGap).
		Alignment(ui.Leading).
		FullWidth().
		Padding(ui.Padding{Left: ui.L24})
}

// stepView shows what a sub-agent wrote or a tool it called, with its duration once it is done.
func stepView(s liveStep) core.View {
	if s.Tool == "" {
		return ui.Text(s.Text).Font(taskTextFont).Color(ui.M7)
	}

	if !s.Done {
		return ui.HStack(spinner(ui.L12), ui.Text(s.Tool).Font(taskTextFont).Color(ui.M7)).Gap(ui.L4).Alignment(ui.Center)
	}

	label := "→ " + s.Tool + " · " + seconds(s.Took)
	if s.Failed {
		label = "→ " + s.Tool + " · fehlgeschlagen nach " + seconds(s.Took)
	}
	return ui.Text(label).Font(taskTextFont).Color(ui.M7)
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

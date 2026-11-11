// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"go.wdy.de/nago/application/ai/completion"
)

// maxLiveSteps bounds the steps of a task, which the chat shows while it works.
const maxLiveSteps = 6

// maxStepText bounds the text of a step in runes.
const maxStepText = 160

// liveTask is what a delegated task does, as far as the chat knows it from the events of its sub-agent, see
// taskBoard. A conversation only keeps the results of its tasks, thus this is lost with the window.
type liveTask struct {
	Status completion.TaskStatus
	// Waiting is set while the task waits for a free slot.
	Waiting bool
	// Queued, Started and Ended are the moments the task was accepted, got its slot and ended.
	Queued, Started, Ended time.Time
	// Timeout is the time limit of the sub-agent, which runs from Started.
	Timeout time.Duration
	// Turn is the turn of the sub-agent, counted from 1.
	Turn int
	// Activity tells what the sub-agent does right now, e.g. "führt read_table aus".
	Activity string
	// Steps are the latest steps of the sub-agent.
	Steps []liveStep
}

// liveStep is something a sub-agent wrote or a tool it called.
type liveStep struct {
	// Text is what the sub-agent wrote before it called its tools.
	Text string
	// Tool is the name of a called tool.
	Tool   string
	callID string
	// Started is when the call started, Took how long it took once it is Done.
	Started time.Time
	Took    time.Duration
	Done    bool
	Failed  bool
}

// elapsed returns how long the task works, or worked, at now. A waiting task has not started yet.
func (t liveTask) elapsed(now time.Time) time.Duration {
	if t.Started.IsZero() {
		return 0
	}
	if !t.Ended.IsZero() {
		return t.Ended.Sub(t.Started)
	}
	return now.Sub(t.Started)
}

// taskBoard collects what the delegated tasks of a chat do, from the events of their sub-agents, so that the chat
// shows it while they work: in the window which started the run and in every window which follows it, see
// activeRun. It is keyed by the id of the task, see [completion.SubEvent.TaskID], which is unique, thus one board
// serves all runs of a window. It is safe for concurrent use.
type taskBoard struct {
	mu    sync.Mutex
	tasks map[string]*liveTask
	// last is the tool which a sub-agent called last
	last string
	// waits is set while the run waits on its tasks, i.e. while it runs delegate or await_tasks
	waits bool
	// since is the start of the current run, whose tasks the progress line counts
	since time.Time
	// version counts the changes, see pulse
	version uint64
	// clock is time.Now, except in tests
	clock func() time.Time
}

func newTaskBoard() *taskBoard {
	return &taskBoard{tasks: map[string]*liveTask{}, clock: time.Now}
}

// boardRef holds the board, which a window shows. A window which follows a run shows the board of the run instead
// of its own. The board is never replaced through [core.State.Set], whose comparison would read the board while
// sub-agents write it.
type boardRef struct {
	mu    sync.Mutex
	board *taskBoard
}

func newBoardRef() *boardRef {
	return &boardRef{board: newTaskBoard()}
}

func (r *boardRef) get() *taskBoard {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.board
}

func (r *boardRef) set(b *taskBoard) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.board = b
}

// observe takes the event of a task.
func (b *taskBoard) observe(e completion.SubEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock()
	t := b.tasks[e.TaskID]
	if t == nil || e.Kind == completion.SubStarted {
		t = &liveTask{Status: completion.TaskRunning, Waiting: true, Queued: now}
		b.tasks[e.TaskID] = t
	}

	switch e.Kind {
	case completion.SubStarted:
	case completion.SubRunning:
		t.Waiting = false
		t.Started = now
		t.Timeout = e.Timeout
		t.Activity = "denkt nach"
	case completion.SubProgressed:
		if e.Progress == nil {
			return
		}
		t.Waiting = false
		if t.Started.IsZero() {
			t.Started = now
		}
		b.progress(t, *e.Progress, now)
	case completion.SubFinished:
		t.Waiting = false
		t.Status = completion.TaskFailed
		if e.Result != nil {
			t.Status = e.Result.Status
		}
		t.Ended = now
		t.Activity = ""
	default:
		return
	}

	b.version++
}

func (b *taskBoard) progress(t *liveTask, p completion.Progress, now time.Time) {
	switch p.Phase {
	case completion.PhaseTurnStarted:
		t.Turn = p.Turn + 1
		t.Activity = "denkt nach"
	case completion.PhaseModelResponded:
		// the final answer is the result of the task, not one of its steps
		if p.Result == nil || !slices.ContainsFunc(p.Result.Message.Content, isToolCall) {
			return
		}
		if text := stepText(p.Result.Message.Content); text != "" {
			t.add(liveStep{Text: text})
		}
	case completion.PhaseToolStarted:
		if p.ToolCall == nil {
			return
		}
		t.add(liveStep{Tool: p.ToolCall.Name, callID: p.ToolCall.ID, Started: now})
		t.Activity = fmt.Sprintf("führt %s aus", p.ToolCall.Name)
		b.last = p.ToolCall.Name
	case completion.PhaseToolCompleted:
		if p.ToolCall == nil {
			return
		}
		for i := len(t.Steps) - 1; i >= 0; i-- {
			s := &t.Steps[i]
			if s.callID == p.ToolCall.ID && !s.Done {
				s.Done = true
				s.Took = now.Sub(s.Started)
				s.Failed = p.ToolResult != nil && p.ToolResult.IsError
				break
			}
		}
		t.Activity = "denkt nach"
	case completion.PhaseCompacted:
		t.Activity = "fasst den bisherigen Verlauf zusammen"
	}
}

func isToolCall(c completion.Content) bool {
	_, ok := c.(completion.ToolCall)
	return ok
}

// add appends a step and keeps the latest ones.
func (t *liveTask) add(s liveStep) {
	t.Steps = append(t.Steps, s)
	if n := len(t.Steps); n > maxLiveSteps {
		t.Steps = slices.Clone(t.Steps[n-maxLiveSteps:])
	}
}

// stepText returns the texts of a model turn as one line, shortened to maxStepText runes.
func stepText(contents []completion.Content) string {
	var sb strings.Builder
	for _, c := range contents {
		if t, ok := c.(completion.Text); ok {
			sb.WriteString(t.Text)
			sb.WriteByte(' ')
		}
	}

	text := strings.Join(strings.Fields(sb.String()), " ")
	if utf8.RuneCountInString(text) <= maxStepText {
		return text
	}

	runes := []rune(text)
	return strings.TrimSpace(string(runes[:maxStepText-1])) + "…"
}

// begin starts a new run: the progress line counts its tasks only.
func (b *taskBoard) begin() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.since = b.clock()
	b.last = ""
	b.waits = false
	b.version++
}

// setWaits states, whether the run waits on its tasks.
func (b *taskBoard) setWaits(waits bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.waits != waits {
		b.waits = waits
		b.version++
	}
}

// snapshot returns a copy of the tasks by id.
func (b *taskBoard) snapshot() map[string]liveTask {
	b.mu.Lock()
	defer b.mu.Unlock()

	res := make(map[string]liveTask, len(b.tasks))
	for id, t := range b.tasks {
		c := *t
		c.Steps = slices.Clone(t.Steps)
		res[id] = c
	}
	return res
}

// pulse returns the version of the board and how many tasks work, so that a window renders as long as tasks work,
// because their clocks move on, and once more after a change.
func (b *taskBoard) pulse() (version uint64, working int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.version, b.working()
}

func (b *taskBoard) working() int {
	n := 0
	for _, t := range b.tasks {
		if t.Status == completion.TaskRunning {
			n++
		}
	}
	return n
}

// status returns the progress line of a run, which waits on its tasks, while any of them works.
func (b *taskBoard) status() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.waits {
		return "", false
	}

	var working, waiting, finished int
	for _, t := range b.tasks {
		switch {
		case t.Queued.Before(b.since):
		case t.Status == completion.TaskRunning && t.Waiting:
			waiting++
		case t.Status == completion.TaskRunning:
			working++
		default:
			finished++
		}
	}
	if working+waiting == 0 {
		return "", false
	}

	return tasksStatusLabel(working, waiting, finished, b.last), true
}

// pulseTasks calls render every second while tasks of the board work, because their clocks move on, and once after
// any other change, until ctx is done.
func pulseTasks(ctx context.Context, b *taskBoard, render func()) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var seen uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		version, working := b.pulse()
		if version == seen && working == 0 {
			continue
		}
		seen = version
		render()
	}
}

// tasksStatusLabel is the progress line while sub-agents work: how many work, wait for a free slot and are done, and
// the tool which a sub-agent called last.
func tasksStatusLabel(working, waiting, finished int, last string) string {
	var label string
	switch working {
	case 0:
		label = fmt.Sprintf("… %d %s auf einen freien Platz", waiting, plural(waiting, "Hintergrundaufgabe wartet", "Hintergrundaufgaben warten"))
		waiting = 0
	case 1:
		label = "… 1 Hintergrundaufgabe arbeitet"
	default:
		label = fmt.Sprintf("… %d Hintergrundaufgaben arbeiten", working)
	}
	if waiting > 0 {
		label += fmt.Sprintf(", %d %s", waiting, plural(waiting, "wartet", "warten"))
	}
	if finished > 0 {
		label += fmt.Sprintf(", %d erledigt", finished)
	}
	if last != "" {
		label += ", zuletzt: " + last
	}
	return label
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// waitsOnTasks tells, whether a run, which calls the named tool, waits on its tasks.
func waitsOnTasks(name string) bool {
	return name == completion.DelegateToolName || name == completion.AwaitTasksToolName
}

// clock formats a duration like a stopwatch, e.g. 2:05 or 1:02:05.
func clock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	s := int(d / time.Second)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// seconds formats the duration of a tool call, e.g. 0,8 s or 12 s.
func seconds(d time.Duration) string {
	if d < 10*time.Second {
		return strings.Replace(fmt.Sprintf("%.1f s", d.Seconds()), ".", ",", 1)
	}
	return fmt.Sprintf("%d s", int(d.Round(time.Second)/time.Second))
}

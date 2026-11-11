// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
)

func progressEvent(id string, p completion.Progress) completion.SubEvent {
	return completion.SubEvent{Kind: completion.SubProgressed, TaskID: id, Progress: &p}
}

func toolEvents(id, callID, name string) (started, completed completion.SubEvent) {
	call := completion.ToolCall{ID: callID, Name: name, Arguments: json.RawMessage(`{}`)}
	started = progressEvent(id, completion.Progress{Phase: completion.PhaseToolStarted, ToolCall: &call})
	completed = progressEvent(id, completion.Progress{Phase: completion.PhaseToolCompleted, ToolCall: &call, ToolResult: &completion.ToolResult{ToolCallID: callID}})
	return started, completed
}

// The board follows a sub-agent from its slot to its end: its turn, what it does, its latest steps and its time.
func TestTaskBoard_FollowsASubAgent(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	b := newTaskBoard()
	b.clock = func() time.Time { return now }
	b.begin()
	b.setWaits(true)

	b.observe(completion.SubEvent{Kind: completion.SubStarted, TaskID: "d1/0", Title: "Lesen"})
	if task := b.snapshot()["d1/0"]; !task.Waiting || task.Status != completion.TaskRunning {
		t.Fatalf("an accepted task waits for its slot: %+v", task)
	}
	if label, ok := b.status(); !ok || label != "… 1 Hintergrundaufgabe wartet auf einen freien Platz" {
		t.Fatalf("label = %q", label)
	}

	now = now.Add(2 * time.Second)
	b.observe(completion.SubEvent{Kind: completion.SubRunning, TaskID: "d1/0", Timeout: 5 * time.Minute})
	b.observe(progressEvent("d1/0", completion.Progress{Phase: completion.PhaseTurnStarted, Turn: 0}))
	b.observe(progressEvent("d1/0", completion.Progress{Phase: completion.PhaseModelResponded, Result: &completion.Result{Message: completion.Message{
		Role: completion.Assistant,
		Content: []completion.Content{
			completion.Text{Text: "Ich lese\n  die Tabelle."},
			completion.ToolCall{ID: "c1", Name: "read_table"},
		},
	}}}))
	started, completed := toolEvents("d1/0", "c1", "read_table")
	b.observe(started)
	if task := b.snapshot()["d1/0"]; task.Activity != "führt read_table aus" || task.Turn != 1 {
		t.Fatalf("activity = %q, turn = %d", task.Activity, task.Turn)
	}
	if label, _ := b.status(); label != "… 1 Hintergrundaufgabe arbeitet, zuletzt: read_table" {
		t.Fatalf("label = %q", label)
	}

	now = now.Add(800 * time.Millisecond)
	b.observe(completed)
	task := b.snapshot()["d1/0"]
	if task.Waiting || task.Timeout != 5*time.Minute || task.elapsed(now) != 800*time.Millisecond {
		t.Fatalf("running task = %+v", task)
	}
	if len(task.Steps) != 2 || task.Steps[0].Text != "Ich lese die Tabelle." || task.Steps[1].Tool != "read_table" || !task.Steps[1].Done || task.Steps[1].Took != 800*time.Millisecond {
		t.Fatalf("steps = %+v", task.Steps)
	}

	// the final answer is the result of the task, not a step
	b.observe(progressEvent("d1/0", completion.Progress{Phase: completion.PhaseModelResponded, Result: &completion.Result{Message: completion.Message{
		Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "42 Zeilen"}},
	}}}))
	if steps := b.snapshot()["d1/0"].Steps; len(steps) != 2 {
		t.Fatalf("the answer became a step: %+v", steps)
	}

	for i := range 10 {
		started, completed := toolEvents("d1/0", "c"+string(rune('a'+i)), "list_columns")
		b.observe(started)
		b.observe(completed)
	}
	if steps := b.snapshot()["d1/0"].Steps; len(steps) != maxLiveSteps || steps[0].Tool != "list_columns" {
		t.Fatalf("expected the latest %d steps, got %+v", maxLiveSteps, steps)
	}

	now = now.Add(time.Minute)
	b.observe(completion.SubEvent{Kind: completion.SubFinished, TaskID: "d1/0", Result: &completion.TaskResult{Status: completion.TaskCompleted}})
	task = b.snapshot()["d1/0"]
	if task.Status != completion.TaskCompleted || task.Activity != "" || task.elapsed(now.Add(time.Hour)) != time.Minute+800*time.Millisecond {
		t.Fatalf("finished task = %+v", task)
	}
	if _, ok := b.status(); ok {
		t.Fatal("no progress line without working tasks")
	}

	// the progress line of the next run counts its own tasks only
	b.begin()
	b.setWaits(true)
	b.observe(completion.SubEvent{Kind: completion.SubStarted, TaskID: "d2/0"})
	b.observe(completion.SubEvent{Kind: completion.SubRunning, TaskID: "d2/0"})
	if label, _ := b.status(); label != "… 1 Hintergrundaufgabe arbeitet" {
		t.Fatalf("label = %q", label)
	}
	b.setWaits(false)
	if _, ok := b.status(); ok {
		t.Fatal("no progress line while the run does not wait on its tasks")
	}
}

// pulse renders as long as tasks work, and once after any other change.
func TestTaskBoard_Pulse(t *testing.T) {
	b := newTaskBoard()
	v0, working := b.pulse()
	if working != 0 {
		t.Fatalf("working = %d", working)
	}

	b.observe(completion.SubEvent{Kind: completion.SubStarted, TaskID: "t"})
	v1, working := b.pulse()
	if v1 == v0 || working != 1 {
		t.Fatalf("version %d -> %d, working = %d", v0, v1, working)
	}

	b.observe(completion.SubEvent{Kind: completion.SubFinished, TaskID: "t", Result: &completion.TaskResult{Status: completion.TaskFailed}})
	if v2, working := b.pulse(); v2 == v1 || working != 0 {
		t.Fatalf("version %d -> %d, working = %d", v1, v2, working)
	}
}

func TestTaskMeta(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	running := liveTask{Status: completion.TaskRunning, Queued: now.Add(-3 * time.Minute), Started: now.Add(-134 * time.Second), Turn: 6, Activity: "führt save_view aus", Timeout: 5 * time.Minute}

	cases := []struct {
		task shownTask
		want string
	}{
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskRunning}, live: running, hasLive: true}, "2:14 · Schritt 6"},
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskRunning}, live: liveTask{Status: completion.TaskRunning, Waiting: true}, hasLive: true}, "wartet auf einen freien Platz"},
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskRunning}}, "läuft"},
		// the sub-agent is done before the call reports its result
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskRunning}, live: liveTask{Status: completion.TaskCompleted, Started: now.Add(-108 * time.Second), Ended: now}, hasLive: true}, "erledigt nach 1:48"},
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskTimeout}, live: liveTask{Status: completion.TaskTimeout, Timeout: 5 * time.Minute}, hasLive: true}, "Zeitlimit von 5:00 erreicht"},
		{shownTask{TaskResult: completion.TaskResult{Status: completion.TaskCompleted}}, "erledigt"},
	}
	for _, c := range cases {
		if got := taskMeta(c.task, now); got != c.want {
			t.Errorf("taskMeta = %q, want %q", got, c.want)
		}
	}
	if got := taskActivity(cases[0].task); got != "führt save_view aus" {
		t.Errorf("taskActivity = %q", got)
	}
	if got := taskActivity(cases[1].task); got != "" {
		t.Errorf("a waiting task does nothing yet, got %q", got)
	}

	tasks := []shownTask{cases[0].task, cases[3].task}
	if got := tasksMeta(tasks, now); got != "1 erledigt" {
		// the second task carries no Queued, thus no time is known for all of them
		t.Errorf("tasksMeta = %q", got)
	}
	cases[3].task.live.Queued = now.Add(-2 * time.Minute)
	tasks = []shownTask{cases[0].task, cases[3].task}
	if got := tasksMeta(tasks, now); got != "1 erledigt · seit 3:00" {
		t.Errorf("tasksMeta = %q", got)
	}
}

func TestClockAndSeconds(t *testing.T) {
	for d, want := range map[time.Duration]string{0: "0:00", 65 * time.Second: "1:05", 3725 * time.Second: "1:02:05", -time.Second: "0:00"} {
		if got := clock(d); got != want {
			t.Errorf("clock(%s) = %q, want %q", d, got, want)
		}
	}
	for d, want := range map[time.Duration]string{800 * time.Millisecond: "0,8 s", 12400 * time.Millisecond: "12 s"} {
		if got := seconds(d); got != want {
			t.Errorf("seconds(%s) = %q, want %q", d, got, want)
		}
	}

	long := stepText([]completion.Content{completion.Text{Text: strings.Repeat("a ", 200)}})
	if n := len([]rune(long)); n != maxStepText || !strings.HasSuffix(long, "…") {
		t.Fatalf("shortened step has %d runes: %q", n, long)
	}
}

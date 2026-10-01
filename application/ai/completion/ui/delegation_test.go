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

	"go.wdy.de/nago/application/ai/completion"
)

func delegationHistory(result string) []completion.Message {
	h := []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "compare"}}},
		{Role: completion.Assistant, Content: []completion.Content{
			completion.ToolCall{ID: "d1", Name: completion.DelegateToolName, Arguments: json.RawMessage(`{"tasks":[{"title":"one","task":"A"},{"title":"two","task":"B"}]}`)},
		}},
	}
	if result != "" {
		h = append(h,
			completion.Message{Role: completion.User, Content: []completion.Content{
				completion.ToolResult{ToolCallID: "d1", Content: []completion.Content{completion.Text{Text: result}}},
			}},
			completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "both done"}}},
		)
	}
	return h
}

func TestTasksOf_LiveAndFinished(t *testing.T) {
	h := delegationHistory("")
	call := h[1].Content[0].(completion.ToolCall)

	running := tasksOf(call, completion.ToolResult{}, false, nil)
	if len(running) != 2 || running[0].Title != "one" || running[0].Status != completion.TaskRunning {
		t.Fatalf("live tasks = %+v", running)
	}

	res := completion.ToolResult{ToolCallID: "d1", Content: []completion.Content{completion.Text{
		Text: `{"results":[{"title":"one","status":"completed","answer":"x","sessionId":"c1"},{"title":"two","status":"failed","error":"boom"}]}`,
	}}}
	done := tasksOf(call, res, true, nil)
	if len(done) != 2 || done[0].Status != completion.TaskCompleted || done[0].SessionID != "c1" || done[1].Error != "boom" {
		t.Fatalf("finished tasks = %+v", done)
	}
}

func TestRenderHistory_DelegationAsTaskList(t *testing.T) {
	views := renderHistory(nil, delegationHistory(`{"results":[{"title":"one","status":"completed","answer":"x"},{"title":"two","status":"timeout"}]}`), historyView{})
	// user prompt, one task list, final reply; the tool result itself stays hidden
	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

func TestTaskLabels(t *testing.T) {
	if got := taskProgressLabel(1, 3); got != "… 1 von 3 Teilaufgaben erledigt" {
		t.Fatalf("label = %q", got)
	}
	if !strings.HasPrefix(tasksHeading(completion.StartTasksToolName, 2), "Teilaufgaben gestartet") {
		t.Fatal("unexpected heading")
	}
	for _, s := range []completion.TaskStatus{completion.TaskRunning, completion.TaskCompleted, completion.TaskFailed, completion.TaskTimeout, completion.TaskCancelled, completion.TaskLost} {
		if taskStatusLabel(s) == string(s) {
			t.Fatalf("status %q has no German label", s)
		}
	}
}

func TestDelegationTools(t *testing.T) {
	if tools, hook := delegationTools(delegationRun{}); tools != nil || hook != nil {
		t.Fatal("delegation must be off without options")
	}

	tools, hook := delegationTools(delegationRun{opts: ChatOptions{Delegation: &DelegationOptions{}}})
	if len(tools) != 1 || tools[0].Def.Name != completion.DelegateToolName || hook != nil {
		t.Fatalf("expected only the delegate tool, got %d", len(tools))
	}

	group := completion.NewTaskRegistry().Group("k")
	tools, hook = delegationTools(delegationRun{opts: ChatOptions{Delegation: &DelegationOptions{BackgroundTasks: true}}, group: group})
	if len(tools) != 4 || hook == nil {
		t.Fatalf("expected delegate and the task tools, got %d", len(tools))
	}
}

func TestScreenToolIsNeverDelegated(t *testing.T) {
	if !ScreenTool(nil, ScreenToolOptions{}).NoDelegate || !ScreenTool(nil, ScreenToolOptions{DisableImage: true}).NoDelegate {
		t.Fatal("the screen tool must be marked NoDelegate")
	}
}

// A delegation call which was refused as a whole shows its tasks as failed with the reason.
func TestTasksOf_RefusedCall(t *testing.T) {
	h := delegationHistory("")
	call := h[1].Content[0].(completion.ToolCall)

	res := completion.ToolResult{ToolCallID: "d1", IsError: true, Content: []completion.Content{completion.Text{Text: "the sub-agent budget of this run is exhausted"}}}
	tasks := tasksOf(call, res, true, nil)
	if len(tasks) != 2 || tasks[0].Status != completion.TaskFailed || !strings.Contains(tasks[0].Error, "budget") {
		t.Fatalf("refused tasks = %+v", tasks)
	}
}

// Tasks the model never awaited get their final state from the results BeforeFinish handed over.
func TestLatestTasks_HandedOver(t *testing.T) {
	start := completion.ToolCall{ID: "s1", Name: completion.StartTasksToolName, Arguments: json.RawMessage(`{"tasks":[{"title":"one","task":"A"}]}`)}
	history := []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "go"}}},
		{Role: completion.Assistant, Content: []completion.Content{start}},
		{Role: completion.User, Content: []completion.Content{completion.ToolResult{ToolCallID: "s1", Content: []completion.Content{completion.Text{
			Text: `{"tasks":[{"id":"task_1","title":"one","status":"running"}]}`,
		}}}}},
		{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "answer without awaiting"}}},
		{Role: completion.User, Content: []completion.Content{completion.Text{
			Text: "[system note, not written by the user]\nBackground tasks you started finished (or are still running), but you answered without awaiting them. Their results follow.\n" +
				`{"finished":[{"id":"task_1","title":"one","status":"completed","answer":"42","sessionId":"c1"}]}`,
		}}},
		{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "the answer is 42"}}},
	}

	latest := latestTasks(history)
	if got := latest["task_1"]; got.Status != completion.TaskCompleted || got.SessionID != "c1" {
		t.Fatalf("latest state = %+v", got)
	}

	res := history[2].Content[0].(completion.ToolResult)
	tasks := tasksOf(start, res, true, latest)
	if len(tasks) != 1 || tasks[0].Status != completion.TaskCompleted || tasks[0].SessionID != "c1" || tasks[0].Title != "one" {
		t.Fatalf("merged tasks = %+v", tasks)
	}

	// user prompt, task list, answer, answer; the hidden prompt stays hidden
	if views := renderHistory(nil, history, historyView{}); len(views) != 4 {
		t.Fatalf("expected 4 views, got %d", len(views))
	}
}

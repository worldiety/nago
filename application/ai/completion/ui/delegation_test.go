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

	running := tasksOf(call, completion.ToolResult{}, false)
	if len(running) != 2 || running[0].Title != "one" || running[0].Status != completion.TaskRunning {
		t.Fatalf("live tasks = %+v", running)
	}

	res := completion.ToolResult{ToolCallID: "d1", Content: []completion.Content{completion.Text{
		Text: `{"results":[{"title":"one","status":"completed","answer":"x","sessionId":"c1"},{"title":"two","status":"failed","error":"boom"}]}`,
	}}}
	done := tasksOf(call, res, true)
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

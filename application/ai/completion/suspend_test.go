// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"errors"
	"testing"

	"go.wdy.de/nago/auth"
)

type suspendFixture struct {
	reads, writes int
	writeArgs     []string
	tools         []Tool
}

func newSuspendFixture() *suspendFixture {
	f := &suspendFixture{}
	read := NewTool("read", "reads", func(struct{}) (struct{}, error) { f.reads++; return struct{}{}, nil })
	write := Tool{
		Def: ToolDef{Name: "write", Description: "writes", Schema: json.RawMessage(`{}`)},
		Invoke: func(_ auth.Subject, args json.RawMessage) (json.RawMessage, error) {
			f.writes++
			f.writeArgs = append(f.writeArgs, string(args))
			return json.RawMessage(`{"ok":true}`), nil
		},
	}.AsMutating("writes something")
	f.tools = []Tool{read, write, NewAskUserTool()}
	return f
}

func toolUse(calls ...ToolCall) Result {
	content := make([]Content, 0, len(calls))
	for _, c := range calls {
		content = append(content, c)
	}
	return Result{Message: Message{Role: Assistant, Content: content}, StopReason: StopToolUse}
}

func askCall(id string) ToolCall {
	return ToolCall{ID: id, Name: AskUserToolName, Arguments: json.RawMessage(`{"question":"which one?","options":["a","b"]}`)}
}

// lastUserResults returns the tool results of the request's last message, keyed by call id.
func lastUserResults(t *testing.T, opts Options) map[string]ToolResult {
	t.Helper()
	last := opts.Messages[len(opts.Messages)-1]
	if last.Role != User {
		t.Fatalf("expected a user turn, got %s", last.Role)
	}
	out := map[string]ToolResult{}
	for _, c := range last.Content {
		if r, ok := c.(ToolResult); ok {
			out[r.ToolCallID] = r
		}
	}
	return out
}

func TestStart_QuestionSuspendsWithoutBlocking(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{
		toolUse(
			ToolCall{ID: "r", Name: "read", Arguments: json.RawMessage(`{}`)},
			askCall("q"),
			ToolCall{ID: "w", Name: "write", Arguments: json.RawMessage(`{"x":1}`)},
		),
		assistantText(StopEndTurn, Text{Text: "done"}),
	}}

	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: f.tools})
	if err != nil {
		t.Fatal(err)
	}
	if out.Suspended == nil || len(out.Suspended.Pending) != 1 || out.Suspended.Pending[0].Question != "which one?" {
		t.Fatalf("expected a suspended question, got %+v", out.Suspended)
	}
	if f.reads != 1 {
		t.Fatalf("reading sibling call must run before suspending, reads=%d", f.reads)
	}
	if f.writes != 0 {
		t.Fatalf("a mutating call next to a question must not run")
	}
	if last := out.History[len(out.History)-1]; last.Role != Assistant {
		t.Fatalf("history must end with the suspended assistant turn, got %+v", last)
	}
	if fake.calls != 1 {
		t.Fatalf("model must not be called again while suspended, calls=%d", fake.calls)
	}

	// resume after a round trip through JSON, as a persisted session would
	buf, err := json.Marshal(*out.Suspended)
	if err != nil {
		t.Fatal(err)
	}
	var cont Continuation
	if err := json.Unmarshal(buf, &cont); err != nil {
		t.Fatal(err)
	}

	final, err := Continue(nil, fake, RunOptions{Options: Options{Messages: out.History}, Tools: f.tools}, cont,
		[]Resolution{{CallID: "q", Answer: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if final.Suspended != nil || final.Result.StopReason != StopEndTurn {
		t.Fatalf("expected the run to finish, got %+v", final)
	}
	if f.reads != 1 || f.writes != 0 {
		t.Fatalf("completed calls must not run again on resume, reads=%d writes=%d", f.reads, f.writes)
	}

	results := lastUserResults(t, fake.reqs[1])
	if len(results) != 3 {
		t.Fatalf("every call of the turn needs a result, got %d", len(results))
	}
	if got := results["q"].Content[0].(Text).Text; got != `{"answer":"b"}` {
		t.Fatalf("unexpected answer payload %s", got)
	}
	if !results["w"].IsError {
		t.Fatalf("deferred mutating call must be reported as not executed")
	}
	if hasDanglingToolUse(final.History) {
		t.Fatalf("history has a dangling tool_use: %+v", final.History)
	}
}

func TestStart_ConfirmMutatingApproveExecutesStoredArguments(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{
		toolUse(ToolCall{ID: "w", Name: "write", Arguments: json.RawMessage(`{"x":42}`)}),
		assistantText(StopEndTurn, Text{Text: "written"}),
	}}
	opts := RunOptions{Options: Options{Messages: userMsg("write")}, Tools: f.tools, ConfirmMutating: true}

	out, err := Start(nil, fake, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out.Suspended == nil || out.Suspended.Pending[0].Kind != PendingApproval || out.Suspended.Pending[0].Effect != "writes something" {
		t.Fatalf("expected a pending approval, got %+v", out.Suspended)
	}
	if f.writes != 0 {
		t.Fatalf("mutating call must wait for approval")
	}

	opts.Messages = out.History
	if _, err := Continue(nil, fake, opts, *out.Suspended, []Resolution{{CallID: "w", Approved: true}}); err != nil {
		t.Fatal(err)
	}
	if f.writes != 1 || f.writeArgs[0] != `{"x":42}` {
		t.Fatalf("approved call must run once with the stored arguments, got %v", f.writeArgs)
	}
}

func TestContinue_RejectDoesNotExecute(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{
		toolUse(ToolCall{ID: "w", Name: "write", Arguments: json.RawMessage(`{}`)}),
		assistantText(StopEndTurn, Text{Text: "ok"}),
	}}
	opts := RunOptions{Options: Options{Messages: userMsg("write")}, Tools: f.tools, ConfirmMutating: true}

	out, err := Start(nil, fake, opts)
	if err != nil {
		t.Fatal(err)
	}
	opts.Messages = out.History
	if _, err := Continue(nil, fake, opts, *out.Suspended, []Resolution{{CallID: "w", Message: "declined"}}); err != nil {
		t.Fatal(err)
	}
	if f.writes != 0 {
		t.Fatalf("rejected call must not run")
	}
	if r := lastUserResults(t, fake.reqs[1])["w"]; !r.IsError || r.Content[0].(Text).Text != "declined" {
		t.Fatalf("unexpected rejection result %+v", r)
	}
}

func TestContinue_RejectsMismatchedResolutions(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{toolUse(askCall("q"))}}
	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: f.tools})
	if err != nil {
		t.Fatal(err)
	}
	opts := RunOptions{Options: Options{Messages: out.History}, Tools: f.tools}

	cases := map[string][]Resolution{
		"missing": nil,
		"twice":   {{CallID: "q", Answer: "a"}, {CallID: "q", Answer: "b"}},
		"unknown": {{CallID: "q", Answer: "a"}, {CallID: "x", Answer: "b"}},
	}
	for name, res := range cases {
		if _, err := Continue(nil, fake, opts, *out.Suspended, res); !errors.Is(err, ErrContinuationMismatch) {
			t.Errorf("%s: expected ErrContinuationMismatch, got %v", name, err)
		}
	}

	// history that already moved on no longer fits the continuation
	moved := RunOptions{Options: Options{Messages: append(out.History, userMsg("other")...)}, Tools: f.tools}
	if _, err := Continue(nil, fake, moved, *out.Suspended, []Resolution{{CallID: "q", Answer: "a"}}); !errors.Is(err, ErrContinuationMismatch) {
		t.Errorf("expected mismatch for a moved-on history, got %v", err)
	}
}

func TestDismiss_ProducesValidHistoryWithoutModelCall(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{toolUse(
		ToolCall{ID: "r", Name: "read", Arguments: json.RawMessage(`{}`)},
		askCall("q"),
	)}}
	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: f.tools})
	if err != nil {
		t.Fatal(err)
	}

	history, err := Dismiss(out.History, *out.Suspended)
	if err != nil {
		t.Fatal(err)
	}
	if hasDanglingToolUse(history) {
		t.Fatalf("dismissed history has a dangling tool_use: %+v", history)
	}
	if fake.calls != 1 {
		t.Fatalf("dismiss must not call the model")
	}
}

func TestRun_FailsOnSuspension(t *testing.T) {
	f := newSuspendFixture()
	fake := &fakeCompletions{results: []Result{toolUse(askCall("q"))}}
	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: f.tools})
	if !errors.Is(err, ErrSuspendUnsupported) {
		t.Fatalf("expected ErrSuspendUnsupported, got %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("unexpected history %+v", history)
	}
}

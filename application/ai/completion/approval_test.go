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
	"slices"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/auth"
)

// markedFixture has a reading tool, a mutating tool and a mutating tool which requires approval.
func markedFixture() []Tool {
	read := NewTool("read", "reads", func(struct{}) (struct{}, error) { return struct{}{}, nil })
	write := NewTool("write", "writes", func(struct{}) (struct{}, error) { return struct{}{}, nil }).AsMutating("writes")
	drop := NewTool("drop", "drops", func(struct{}) (struct{}, error) { return struct{}{}, nil }).AsMutating("drops everything").RequireApproval()
	return []Tool{read, write, drop}
}

// With ConfirmMarked, only the call of the marked tool waits for approval.
func TestStart_ConfirmMarkedHoldsOnlyMarkedTools(t *testing.T) {
	f := newSuspendFixture()
	drops := 0
	drop := NewTool("drop", "drops", func(struct{}) (struct{}, error) { drops++; return struct{}{}, nil }).AsMutating("drops everything").RequireApproval()
	tools := append(slices.Clone(f.tools), drop)

	fake := &fakeCompletions{results: []Result{
		toolUse(ToolCall{ID: "w", Name: "write", Arguments: json.RawMessage(`{}`)}, ToolCall{ID: "d", Name: "drop", Arguments: json.RawMessage(`{}`)}),
		assistantText(StopEndTurn, Text{Text: "done"}),
	}}

	opts := RunOptions{Options: Options{Messages: userMsg("go")}, Tools: tools, ConfirmMarked: true}
	out, err := Start(nil, fake, opts)
	if err != nil {
		t.Fatal(err)
	}

	if out.Suspended == nil || len(out.Suspended.Pending) != 1 || out.Suspended.Pending[0].Call.ID != "d" || out.Suspended.Pending[0].Effect != "drops everything" {
		t.Fatalf("expected only the marked call to wait, got %+v", out.Suspended)
	}

	if f.writes != 1 || drops != 0 {
		t.Fatalf("the unmarked call must run and the marked one wait, writes=%d drops=%d", f.writes, drops)
	}

	opts.Messages = out.History
	if _, err := Continue(nil, fake, opts, *out.Suspended, []Resolution{{CallID: "d", Approved: true}}); err != nil {
		t.Fatal(err)
	}

	if drops != 1 || f.writes != 1 {
		t.Fatalf("the approved call must run once, writes=%d drops=%d", f.writes, drops)
	}
}

// ConfirmMutating takes precedence over ConfirmMarked, and the mark alone does not hold anything.
func TestNeedsApproval(t *testing.T) {
	tools := markedFixture()
	cases := []struct {
		name string
		opts RunOptions
		want []string
	}{
		{name: "nothing", opts: RunOptions{}, want: nil},
		{name: "all", opts: RunOptions{ConfirmMutating: true}, want: []string{"write", "drop"}},
		{name: "marked", opts: RunOptions{ConfirmMarked: true}, want: []string{"drop"}},
		{name: "both", opts: RunOptions{ConfirmMutating: true, ConfirmMarked: true}, want: []string{"write", "drop"}},
	}

	for _, tc := range cases {
		var got []string
		for _, tool := range tools {
			if need, _, _ := tc.opts.approval(nil, tool, ToolCall{}); need {
				got = append(got, tool.Def.Name)
			}
		}

		if !slices.Equal(got, tc.want) {
			t.Fatalf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}

	// a mark without Mutating has no effect
	readOnly := tools[0]
	readOnly.RequiresApproval = true
	if need, _, _ := (RunOptions{ConfirmMarked: true}).approval(nil, readOnly, ToolCall{}); need {
		t.Fatal("a reading tool must never wait")
	}
}

func TestDelegate_SubToolSetWithMarkedTools(t *testing.T) {
	cases := []struct {
		name          string
		cfg           DelegateConfig
		confirm       bool
		confirmMarked bool
		want          []string
	}{
		{name: "no confirmation", cfg: DelegateConfig{AllowMutating: true}, want: []string{"drop", "read", "write"}},
		{name: "marked by the parent", cfg: DelegateConfig{AllowMutating: true}, confirmMarked: true, want: []string{"read", "write"}},
		{name: "marked in config", cfg: DelegateConfig{AllowMutating: true, ConfirmMarked: true}, want: []string{"read", "write"}},
		{name: "marked without mutating", cfg: DelegateConfig{}, confirmMarked: true, want: []string{"read"}},
		{name: "all wins over marked", cfg: DelegateConfig{AllowMutating: true}, confirm: true, confirmMarked: true, want: []string{"read"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}))
			fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
				if !isSub(opts) {
					return parent(opts), nil
				}
				return echoSub(ctx, opts)
			}}

			res, _, _, err := runDelegation(t, fake, RunOptions{
				Options:         Options{Messages: userMsg("go")},
				Tools:           append(markedFixture(), NewDelegateTool(tc.cfg)),
				ConfirmMutating: tc.confirm,
				ConfirmMarked:   tc.confirmMarked,
			})
			if err != nil {
				t.Fatal(err)
			}
			if res.Results[0].Status != TaskCompleted {
				t.Fatalf("task failed: %+v", res.Results[0])
			}

			subs := fake.subRequests()
			if len(subs) != 1 {
				t.Fatalf("expected one sub request, got %d", len(subs))
			}
			if got := toolNames(subs[0].Tools); !slices.Equal(got, tc.want) {
				t.Fatalf("sub tools = %v, want %v", got, tc.want)
			}
		})
	}
}

// A sub-agent which asks for a marked tool fails with the hint to do it in the calling run.
func TestDelegate_MarkedToolIsReservedForTheCaller(t *testing.T) {
	parent := parentScript(delegateCall("d1",
		delegateTaskIn{Title: "drop", Task: "A", Tools: []string{"drop"}},
		delegateTaskIn{Title: "write", Task: "B", Tools: []string{"write"}},
	))
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}
		return echoSub(ctx, opts)
	}}

	res, _, _, err := runDelegation(t, fake, RunOptions{
		Options:       Options{Messages: userMsg("go")},
		Tools:         append(markedFixture(), NewDelegateTool(DelegateConfig{AllowMutating: true})),
		ConfirmMarked: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.Results[0].Status != TaskFailed || !strings.Contains(res.Results[0].Error, "approval of the user") {
		t.Fatalf("the marked tool must be refused: %+v", res.Results[0])
	}

	if res.Results[1].Status != TaskCompleted {
		t.Fatalf("the unmarked mutating tool must be available: %+v", res.Results[1])
	}
}

// conditionalDelete deletes n items and needs approval only for more than one.
func conditionalDelete(deleted *[]int) Tool {
	type in struct {
		N int `json:"n"`
	}

	t := NewTool("delete", "deletes items", func(args in) (struct{}, error) {
		*deleted = append(*deleted, args.N)
		return struct{}{}, nil
	}).AsMutating("deletes items")

	t.ApprovalFor = func(subject auth.Subject, call ToolCall) (bool, string, error) {
		var args in
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return false, "", err
		}

		if args.N < 0 {
			return false, "", errors.New("n must not be negative")
		}

		if args.N > 1 {
			return true, fmt.Sprintf("deletes %d items", args.N), nil
		}

		return false, "", nil
	}

	return t
}

// ApprovalFor decides per call: a single item is deleted right away, many wait with their own effect, and an
// error is reported without executing anything.
func TestStart_ApprovalForDecidesPerCall(t *testing.T) {
	var deleted []int
	tools := []Tool{conditionalDelete(&deleted)}

	fake := &fakeCompletions{results: []Result{
		toolUse(
			ToolCall{ID: "one", Name: "delete", Arguments: json.RawMessage(`{"n":1}`)},
			ToolCall{ID: "many", Name: "delete", Arguments: json.RawMessage(`{"n":5}`)},
			ToolCall{ID: "bad", Name: "delete", Arguments: json.RawMessage(`{"n":-1}`)},
		),
		assistantText(StopEndTurn, Text{Text: "done"}),
	}}

	opts := RunOptions{Options: Options{Messages: userMsg("go")}, Tools: tools, ConfirmMarked: true}
	out, err := Start(nil, fake, opts)
	if err != nil {
		t.Fatal(err)
	}

	if out.Suspended == nil || len(out.Suspended.Pending) != 1 || out.Suspended.Pending[0].Call.ID != "many" || out.Suspended.Pending[0].Effect != "deletes 5 items" {
		t.Fatalf("expected only the large deletion to wait, got %+v", out.Suspended)
	}

	if !slices.Equal(deleted, []int{1}) {
		t.Fatalf("expected only the single deletion, got %v", deleted)
	}

	opts.Messages = out.History
	if _, err := Continue(nil, fake, opts, *out.Suspended, []Resolution{{CallID: "many", Approved: true}}); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(deleted, []int{1, 5}) {
		t.Fatalf("the approved deletion must run once, got %v", deleted)
	}

	results := lastUserResults(t, fake.reqs[1])
	if r := results["bad"]; !r.IsError || !strings.Contains(r.Content[0].(Text).Text, "negative") {
		t.Fatalf("the error of ApprovalFor must be the result, got %+v", r)
	}
}

// ApprovalFor is only consulted when the run confirms marked tools.
func TestApprovalForOnlyInMarkedMode(t *testing.T) {
	var deleted []int
	tool := conditionalDelete(&deleted)
	call := ToolCall{Arguments: json.RawMessage(`{"n":5}`)}

	if need, _, _ := (RunOptions{}).approval(nil, tool, call); need {
		t.Fatal("a run without confirmation must not ask")
	}

	if need, effect, _ := (RunOptions{ConfirmMutating: true}).approval(nil, tool, ToolCall{Arguments: json.RawMessage(`{"n":1}`)}); !need || effect != "deletes items" {
		t.Fatalf("confirming every change must ask with the static effect, got %v %q", need, effect)
	}

	silent := tool
	silent.ApprovalFor = func(auth.Subject, ToolCall) (bool, string, error) { return true, "", nil }
	if need, effect, _ := (RunOptions{ConfirmMarked: true}).approval(nil, silent, call); !need || effect != "deletes items" {
		t.Fatalf("an empty effect falls back to Confirm, got %v %q", need, effect)
	}

	marked := tool.RequireApproval()
	if need, effect, _ := (RunOptions{ConfirmMarked: true}).approval(nil, marked, ToolCall{Arguments: json.RawMessage(`{"n":1}`)}); !need || effect != "deletes items" {
		t.Fatalf("a marked tool always asks, got %v %q", need, effect)
	}
}

// A sub-agent gets a tool with ApprovalFor, but a call which needs approval fails with the hint to report back.
func TestDelegate_ApprovalForRefusesInSubAgent(t *testing.T) {
	var deleted []int
	tools := []Tool{conditionalDelete(&deleted)}

	parent := parentScript(delegateCall("d1", delegateTaskIn{Title: "a", Task: "A"}))
	var subResults []ToolResult
	var mutex sync.Mutex
	fake := &scriptFake{respond: func(ctx context.Context, opts Options) (Result, error) {
		if !isSub(opts) {
			return parent(opts), nil
		}

		if r, ok := lastToolResult(opts); ok {
			mutex.Lock()
			subResults = append(subResults, r)
			mutex.Unlock()
			if len(subResults) == 1 {
				return Result{Message: Message{Role: Assistant, Content: []Content{ToolCall{ID: "s2", Name: "delete", Arguments: json.RawMessage(`{"n":1}`)}}}, StopReason: StopToolUse}, nil
			}
			return textResult("sub done", Usage{}), nil
		}

		return Result{Message: Message{Role: Assistant, Content: []Content{ToolCall{ID: "s1", Name: "delete", Arguments: json.RawMessage(`{"n":5}`)}}}, StopReason: StopToolUse}, nil
	}}

	res, _, _, err := runDelegation(t, fake, RunOptions{
		Options:       Options{Messages: userMsg("go")},
		Tools:         append(tools, NewDelegateTool(DelegateConfig{AllowMutating: true})),
		ConfirmMarked: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.Results[0].Status != TaskCompleted {
		t.Fatalf("task failed: %+v", res.Results[0])
	}

	if len(subResults) != 2 || !subResults[0].IsError || !strings.Contains(subResults[0].Content[0].(Text).Text, "approval of the user") {
		t.Fatalf("the large deletion must be refused with the hint, got %+v", subResults)
	}

	if !slices.Equal(deleted, []int{1}) {
		t.Fatalf("only the single deletion may run in the sub-agent, got %v", deleted)
	}
}

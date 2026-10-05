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
	"slices"
	"strings"
	"testing"
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
			if tc.opts.needsApproval(tool) {
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
	if (RunOptions{ConfirmMarked: true}).needsApproval(readOnly) {
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

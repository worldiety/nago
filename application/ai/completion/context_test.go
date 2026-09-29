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
	"sync"
	"testing"

	"go.wdy.de/nago/auth"
)

func TestRun_CancelBetweenTurnsKeepsValidHistory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The tool cancels the run while it executes, so the loop must stop before the next model turn.
	stopper := NewTool("stop", "stops", func(struct{}) (struct{}, error) { cancel(); return struct{}{}, nil })

	fake := &fakeCompletions{results: []Result{
		toolUse(ToolCall{ID: "1", Name: "stop", Arguments: json.RawMessage(`{}`)}),
		assistantText(StopEndTurn, Text{Text: "never reached"}),
	}}

	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: []Tool{stopper}, Context: ctx})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancellation, got %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("the model was asked again after cancellation: %d calls", fake.calls)
	}
	if len(out.History) != 3 || hasDanglingToolUse(out.History) || !out.Progressed {
		t.Fatalf("invalid history after cancellation: %+v", out.History)
	}
}

func TestRun_CancelWithinTurnAnswersRemainingCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ran := 0
	stopper := NewTool("stop", "stops", func(struct{}) (struct{}, error) { ran++; cancel(); return struct{}{}, nil })

	fake := &fakeCompletions{results: []Result{
		toolUse(
			ToolCall{ID: "1", Name: "stop", Arguments: json.RawMessage(`{}`)},
			ToolCall{ID: "2", Name: "stop", Arguments: json.RawMessage(`{}`)},
		),
	}}

	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Tools: []Tool{stopper}, Context: ctx})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancellation, got %v", err)
	}
	if ran != 1 {
		t.Fatalf("a call ran after cancellation: %d", ran)
	}
	if hasDanglingToolUse(out.History) {
		t.Fatalf("dangling tool use: %+v", out.History)
	}

	results := lastUserResults(t, Options{Messages: out.History})
	if results["1"].IsError || !results["2"].IsError || extractText(Message{Content: results["2"].Content}) != cancelledText {
		t.Fatalf("unexpected results %+v", results)
	}
}

func TestRun_AlreadyCancelledDoesNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fake := &fakeCompletions{results: []Result{assistantText(StopEndTurn, Text{Text: "hi"})}}
	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}, Context: ctx})
	if !errors.Is(err, context.Canceled) || fake.calls != 0 || out.Progressed {
		t.Fatalf("err=%v calls=%d progressed=%v", err, fake.calls, out.Progressed)
	}
}

func TestRun_UsageSummedOverAllCompletions(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) { return addOut{Sum: in.A + in.B}, nil })

	first := toolUse(ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":1,"b":2}`)})
	first.Usage = Usage{InputTokens: 100, OutputTokens: 10, CacheReadTokens: 50}
	final := assistantText(StopEndTurn, Text{Text: "3"})
	final.Usage = Usage{InputTokens: 120, OutputTokens: 5, CacheWriteTokens: 7}

	var mu sync.Mutex
	var reported []Usage
	out, err := Start(nil, &fakeCompletions{results: []Result{first, final}}, RunOptions{
		Options: Options{Messages: userMsg("go")},
		Tools:   []Tool{tool},
		OnUsage: func(u Usage) {
			mu.Lock()
			reported = append(reported, u)
			mu.Unlock()
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := Usage{InputTokens: 220, OutputTokens: 15, CacheReadTokens: 50, CacheWriteTokens: 7}
	if out.Usage != want {
		t.Fatalf("usage = %+v, want %+v", out.Usage, want)
	}
	if out.Result.Usage != final.Usage {
		t.Fatalf("result usage must stay the last turn: %+v", out.Result.Usage)
	}
	if len(reported) != 2 {
		t.Fatalf("OnUsage called %d times", len(reported))
	}
}

func TestRun_UsageIncludesCompaction(t *testing.T) {
	fake := &compactFake{overflowUntil: 1}
	out, err := Start(nil, usageEvery{fake, Usage{InputTokens: 1}}, RunOptions{Options: Options{Messages: longHistory(20)}})
	if err != nil {
		t.Fatal(err)
	}
	// one summary plus the successful main turn; the failed main turn reports nothing
	if out.Usage.InputTokens != fake.summaryCalls+1 {
		t.Fatalf("usage %+v after %d summaries", out.Usage, fake.summaryCalls)
	}
}

// usageEvery adds a fixed usage to every successful completion.
type usageEvery struct {
	*compactFake
	usage Usage
}

func (u usageEvery) Complete(ctx context.Context, subject auth.Subject, opts Options) (Result, error) {
	res, err := u.compactFake.Complete(ctx, subject, opts)
	res.Usage = u.usage
	return res, err
}

func TestRun_BeforeFinishContinuesTheLoop(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopEndTurn, Text{Text: "first"}),
		assistantText(StopEndTurn, Text{Text: "second"}),
	}}

	hooks := 0
	out, err := Start(nil, fake, RunOptions{
		Options: Options{Messages: userMsg("go")},
		BeforeFinish: func(ctx context.Context) string {
			hooks++
			if hooks == 1 {
				return "late results"
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if hooks != 2 || fake.calls != 2 || FinalAnswer(out.History) != "second" {
		t.Fatalf("hooks=%d calls=%d history=%+v", hooks, fake.calls, out.History)
	}

	injected := out.History[2]
	if !IsLoopPrompt(injected) || extractText(injected) != finishPromptMarker+"late results" {
		t.Fatalf("injected turn is not a hidden loop prompt: %+v", injected)
	}
}

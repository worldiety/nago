// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

// compactFake distinguishes the summarization request (identified by the summary system prompt) from the
// main agentic turn. The first overflowUntil main-turn calls fail with ContextWindowExceeded, the rest
// succeed. Summarization calls always succeed with a short summary.
type compactFake struct {
	overflowUntil int
	mainCalls     int
	summaryCalls  int
}

func (f *compactFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *compactFake) Complete(_ context.Context, _ auth.Subject, opts Options) (Result, error) {
	if opts.System == defaultSummaryPrompt {
		f.summaryCalls++
		return Result{
			Message:    Message{Role: Assistant, Content: []Content{Text{Text: "SUMMARY"}}},
			StopReason: StopEndTurn,
		}, nil
	}

	f.mainCalls++
	if f.mainCalls <= f.overflowUntil {
		return Result{}, ContextWindowError{Limit: 100, Tokens: 200}
	}

	return Result{
		Message:    Message{Role: Assistant, Content: []Content{Text{Text: "final answer"}}},
		StopReason: StopEndTurn,
	}, nil
}

func (f *compactFake) Stream(context.Context, auth.Subject, Options) iter.Seq2[Delta, error] {
	return nil
}

func longHistory(n int) []Message {
	out := make([]Message, 0, n)
	for i := 0; i < n; i++ {
		role := User
		if i%2 == 1 {
			role = Assistant
		}
		out = append(out, Message{Role: role, Content: []Content{
			Text{Text: fmt.Sprintf("this is message number %d with some filler content", i)},
		}})
	}
	return out
}

func TestContextWindowError_Is(t *testing.T) {
	err := error(ContextWindowError{Limit: 200000, Tokens: 215534})
	if !errors.Is(err, ContextWindowExceeded) {
		t.Fatalf("expected errors.Is to match ContextWindowExceeded")
	}

	var cwe ContextWindowError
	if !errors.As(err, &cwe) || cwe.Limit != 200000 || cwe.Tokens != 215534 {
		t.Fatalf("expected to extract details, got %+v", cwe)
	}
}

func TestRun_CompactsOnContextWindowExceeded(t *testing.T) {
	history := longHistory(12)
	before := runeLen(history)

	fake := &compactFake{overflowUntil: 1}

	res, out, err := Run(nil, fake, RunOptions{
		Options:   Options{Messages: history},
		Compactor: NewSummaryCompactor(SummaryCompactorConfig{}),
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if fake.summaryCalls != 1 {
		t.Fatalf("expected exactly one summarization call, got %d", fake.summaryCalls)
	}

	if res.StopReason != StopEndTurn {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}

	if runeLen(out) >= before {
		t.Fatalf("expected compacted history to shrink (before=%d, after=%d)", before, runeLen(out))
	}

	first := out[0]
	if first.Role != User || len(first.Content) != 1 {
		t.Fatalf("expected leading summary message, got %+v", first)
	}
	if txt, ok := first.Content[0].(Text); !ok || !strings.Contains(txt.Text, "SUMMARY") {
		t.Fatalf("leading message should carry the summary, got %+v", first.Content[0])
	}
}

func TestRun_CompactsByDefault(t *testing.T) {
	// Without an explicit Compactor, Run must initialize and use the default summarizing compactor to recover
	// from a context window overflow.
	fake := &compactFake{overflowUntil: 1}

	res, _, err := Run(nil, fake, RunOptions{
		Options: Options{Messages: longHistory(12)},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if fake.summaryCalls != 1 {
		t.Fatalf("expected the default compactor to summarize once, got %d", fake.summaryCalls)
	}
	if res.StopReason != StopEndTurn {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}
}

func TestSplitIndex_KeepsToolPairsIntact(t *testing.T) {
	// 0 user, 1 assistant(tool_use id=1), 2 user(tool_result id=1), 3 assistant final
	history := []Message{
		{Role: User, Content: []Content{Text{Text: "question"}}},
		{Role: Assistant, Content: []Content{ToolCall{ID: "1", Name: "add"}}},
		{Role: User, Content: []Content{ToolResult{ToolCallID: "1", Content: []Content{Text{Text: "3"}}}}},
		{Role: Assistant, Content: []Content{Text{Text: "the answer is 3"}}},
	}

	// keepLastN == 2 would naively start the tail at index 2, i.e. on the tool_result message. splitIndex
	// must push the boundary forward so the retained tail never begins with a dangling tool_result.
	start := splitIndex(history, 2)
	if hasToolResult(history[start]) {
		t.Fatalf("split index %d starts on a tool_result message", start)
	}
}

func TestSummaryCompactor_TruncatesWhenNoPrefix(t *testing.T) {
	// A single, oversized message has no older prefix to summarize; the compactor must still shrink it via
	// rune-safe truncation so the run can make progress.
	big := strings.Repeat("ü", 10_000) // multi-byte rune to prove unicode safety
	history := []Message{{Role: User, Content: []Content{Text{Text: big}}}}
	before := runeLen(history)

	fake := &compactFake{}
	out, err := NewSummaryCompactor(SummaryCompactorConfig{})(context.Background(), nil, fake, Options{}, history)
	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	if runeLen(out) >= before {
		t.Fatalf("expected truncation to shrink history (before=%d, after=%d)", before, runeLen(out))
	}
	if fake.summaryCalls != 0 {
		t.Fatalf("did not expect a summarization call for a single oversized message")
	}
}

// A history which overflows again right after a compaction consists of a summary and recent messages which are too
// large by themselves. The compactor has to cut those, also tool results, without paying for a summary of the summary;
// a message of normal size is left alone.
func TestSummaryCompactor_CutsTheRecentMessagesWhenNothingOlderIsLeft(t *testing.T) {
	big := strings.Repeat("x", 10_000)
	summary := Message{Role: User, Content: []Content{Text{Text: summaryMarker + "former"}}}
	call := Message{Role: Assistant, Content: []Content{ToolCall{ID: "1", Name: "read", Arguments: []byte(`{}`)}}}

	tests := []struct {
		name          string
		history       []Message
		wantSummaries int
		wantShrink    bool
	}{
		{
			name: "a large tool result behind a summary",
			history: []Message{
				summary, call,
				{Role: User, Content: []Content{ToolResult{ToolCallID: "1", Content: []Content{Text{Text: big}}}}},
				{Role: Assistant, Content: []Content{Text{Text: "a"}}},
				{Role: User, Content: []Content{Text{Text: "b"}}},
				{Role: Assistant, Content: []Content{Text{Text: "c"}}},
				{Role: User, Content: []Content{Text{Text: "d"}}},
			},
			wantShrink: true,
		},
		{
			name: "small messages behind a summary",
			history: []Message{
				summary, call,
				{Role: User, Content: []Content{ToolResult{ToolCallID: "1", Content: []Content{Text{Text: "tiny"}}}}},
				{Role: Assistant, Content: []Content{Text{Text: "a"}}},
				{Role: User, Content: []Content{Text{Text: "b"}}},
				{Role: Assistant, Content: []Content{Text{Text: "c"}}},
				{Role: User, Content: []Content{Text{Text: "d"}}},
			},
		},
		{
			name: "older messages are summarized",
			history: []Message{
				{Role: User, Content: []Content{Text{Text: big}}},
				{Role: Assistant, Content: []Content{Text{Text: "a"}}},
				{Role: User, Content: []Content{Text{Text: "b"}}},
				{Role: Assistant, Content: []Content{Text{Text: "c"}}},
				{Role: User, Content: []Content{Text{Text: "d"}}},
				{Role: Assistant, Content: []Content{Text{Text: "e"}}},
				{Role: User, Content: []Content{Text{Text: "f"}}},
			},
			wantSummaries: 1,
			wantShrink:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &compactFake{}
			out, err := NewSummaryCompactor(SummaryCompactorConfig{})(context.Background(), nil, fake, Options{}, tc.history)
			if err != nil {
				t.Fatal(err)
			}
			if fake.summaryCalls != tc.wantSummaries {
				t.Fatalf("%d summaries, want %d", fake.summaryCalls, tc.wantSummaries)
			}
			if got := runeLen(out) < runeLen(tc.history); got != tc.wantShrink {
				t.Fatalf("shrunk = %v (%d of %d runes), want %v", got, runeLen(out), runeLen(tc.history), tc.wantShrink)
			}
		})
	}
}

// A summary which is not smaller than what it summarizes does not help, so the largest message is cut instead.
func TestSummaryCompactor_SummaryWhichDoesNotShrink(t *testing.T) {
	history := longHistory(7)
	history[6].Content = []Content{Text{Text: strings.Repeat("x", 10_000)}}

	fake := &wordyFake{}
	out, err := NewSummaryCompactor(SummaryCompactorConfig{})(context.Background(), nil, fake, Options{}, history)
	if err != nil {
		t.Fatal(err)
	}
	if runeLen(out) >= runeLen(history) {
		t.Fatalf("the history did not shrink: %d of %d runes", runeLen(out), runeLen(history))
	}
	if fake.calls != 1 || hasText(out, summaryMarker) || len(out) != len(history) {
		t.Fatalf("the largest message must be cut instead of using the summary: %d calls, %d of %d messages", fake.calls, len(out), len(history))
	}
}

// wordyFake answers with a summary longer than what it summarizes.
type wordyFake struct{ calls int }

func (f *wordyFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *wordyFake) Complete(context.Context, auth.Subject, Options) (Result, error) {
	f.calls++
	return Result{Message: Message{Role: Assistant, Content: []Content{Text{Text: strings.Repeat("y", 3_000)}}}, StopReason: StopEndTurn}, nil
}

func (f *wordyFake) Stream(context.Context, auth.Subject, Options) iter.Seq2[Delta, error] {
	return nil
}

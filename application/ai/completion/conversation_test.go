// Copyright (c) 2026 worldiety GmbH
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
	"iter"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

// conversationFake records every request to the model and answers it with next, called with the number of the
// request; a request to summarize is answered by itself and neither recorded nor counted.
type conversationFake struct {
	requests [][]Message
	next     func(n int) (Result, error)
}

func (f *conversationFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *conversationFake) Complete(_ context.Context, _ auth.Subject, opts Options) (Result, error) {
	if opts.System == defaultSummaryPrompt {
		return Result{Message: Message{Role: Assistant, Content: []Content{Text{Text: "SUMMARY"}}}, StopReason: StopEndTurn}, nil
	}
	f.requests = append(f.requests, slices.Clone(opts.Messages))
	return f.next(len(f.requests))
}

func (f *conversationFake) Stream(context.Context, auth.Subject, Options) iter.Seq2[Delta, error] {
	return nil
}

func overflow() (Result, error) { return Result{}, ContextWindowError{Limit: 100, Tokens: 200} }

func finalAnswer(text string) (Result, error) {
	return Result{Message: Message{Role: Assistant, Content: []Content{Text{Text: text}}}, StopReason: StopEndTurn}, nil
}

func noopTool() Tool {
	return NewTool("noop", "does nothing", func(struct{}) (struct{}, error) { return struct{}{}, nil })
}

func noopCall(id string) Result {
	return toolUse(ToolCall{ID: id, Name: "noop", Arguments: json.RawMessage(`{}`)})
}

func hasText(msgs []Message, text string) bool {
	return slices.ContainsFunc(msgs, func(m Message) bool { return strings.Contains(extractText(m), text) })
}

// assertValid checks what a provider demands: every tool call has exactly one result and every result has a call.
func assertValid(t *testing.T, msgs []Message) {
	t.Helper()
	calls, results := map[string]int{}, map[string]int{}
	for _, m := range msgs {
		for _, c := range m.Content {
			switch v := c.(type) {
			case ToolCall:
				calls[v.ID]++
			case ToolResult:
				results[v.ToolCallID]++
			}
		}
	}
	if !reflect.DeepEqual(calls, results) {
		t.Fatalf("calls and results do not match: %v and %v", calls, results)
	}
}

// A compaction shortens only what the model reads: the whole conversation stays with all its turns and media.
func TestStart_CompactionKeepsTheConversation(t *testing.T) {
	media := Media{MimeType: file.Type("image/png"), Source: Source{Data: []byte("png")}}
	long := func() []Message {
		msgs := longHistory(12)
		msgs[0].Content = append(msgs[0].Content, media)
		return msgs
	}

	final := Message{Role: Assistant, Content: []Content{Text{Text: "done"}}}
	step := noopCall("c1").Message
	stepResult := Message{Role: User, Content: []Content{ToolResult{ToolCallID: "c1", Content: []Content{Text{Text: "{}"}}}}}

	tests := []struct {
		name string
		// script are the answers to the requests, in order
		script []func() (Result, error)
		// wantCompactions are the compactions, wantHistoryTurns the turns of History after the run, wantAdded what the
		// conversation got behind the input
		wantCompactions  int
		wantHistoryTurns int
		wantAdded        []Message
	}{
		{
			name:             "without an overflow there is only one list",
			script:           []func() (Result, error){func() (Result, error) { return finalAnswer("done") }},
			wantHistoryTurns: 13,
			wantAdded:        []Message{final},
		},
		{
			name:             "one overflow",
			script:           []func() (Result, error){overflow, func() (Result, error) { return finalAnswer("done") }},
			wantCompactions:  1,
			wantHistoryTurns: 8,
			wantAdded:        []Message{final},
		},
		{
			name: "two overflows",
			script: []func() (Result, error){
				overflow,
				func() (Result, error) { return noopCall("c1"), nil },
				overflow,
				func() (Result, error) { return finalAnswer("done") },
			},
			wantCompactions:  2,
			wantHistoryTurns: 8,
			wantAdded:        []Message{step, stepResult, final},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &conversationFake{next: func(n int) (Result, error) { return tc.script[n-1]() }}

			var compacted []Progress
			out, err := Start(nil, fake, RunOptions{
				Options: Options{Messages: long()},
				Tools:   []Tool{noopTool()},
				OnProgress: func(p Progress) {
					if p.Phase == PhaseCompacted {
						compacted = append(compacted, p)
					}
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(out.History) != tc.wantHistoryTurns {
				t.Fatalf("history for the model has %d turns, want %d", len(out.History), tc.wantHistoryTurns)
			}
			if len(compacted) != tc.wantCompactions {
				t.Fatalf("%d compactions reported, want %d", len(compacted), tc.wantCompactions)
			}

			if tc.wantCompactions == 0 {
				if out.Conversation != nil {
					t.Fatalf("without a compaction there must be one list, got %d conversation turns", len(out.Conversation))
				}
				return
			}

			want := append(long(), tc.wantAdded...)
			if !reflect.DeepEqual(out.Conversation, want) {
				t.Fatalf("the conversation changed:\n%+v\nwant\n%+v", out.Conversation, want)
			}
			if hasText(out.Conversation, "SUMMARY") {
				t.Fatal("a summary must not be part of the conversation")
			}
			if !hasText(out.History, "SUMMARY") || hasText(out.History, "message number 0 ") {
				t.Fatalf("the model must read the summary instead of the old turns: %+v", out.History)
			}

			last := fake.requests[len(fake.requests)-1]
			if !hasText(last, "SUMMARY") || hasText(last, "message number 0 ") {
				t.Fatalf("the request after the compaction holds the old turns: %+v", last)
			}

			// reported at once, with the media and the list the model reads from now on
			if p := compacted[0]; !slices.ContainsFunc(p.Conversation[0].Content, func(c Content) bool { _, ok := c.(Media); return ok }) || !hasText(p.History, "SUMMARY") {
				t.Fatalf("the report lacks a list: %+v", p)
			}
		})
	}
}

// A run continues with both lists: it adds to both, and the whole conversation is still whole after the next
// compaction, also if the shortened version gets shortened again.
func TestStart_ContinuesBothLists(t *testing.T) {
	whole := longHistory(12)
	short := append([]Message{{Role: User, Content: []Content{Text{Text: "[Summary of earlier conversation]\nOLD"}}}}, slices.Clone(whole[6:])...)
	input := Message{Role: User, Content: []Content{Text{Text: "next question"}}}

	tests := []struct {
		name      string
		overflows []int
		wantRead  string
	}{
		{name: "no further compaction", wantRead: "OLD"},
		{name: "a further compaction", overflows: []int{2}, wantRead: "SUMMARY"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &conversationFake{next: func(n int) (Result, error) {
				switch {
				case slices.Contains(tc.overflows, n):
					return overflow()
				case n == 1:
					return noopCall("c1"), nil
				}
				return finalAnswer("done")
			}}

			in := append(slices.Clone(short), input)
			inWhole := append(slices.Clone(whole), input)
			out, err := Start(nil, fake, RunOptions{
				Options:      Options{Messages: in},
				Conversation: inWhole,
				Tools:        []Tool{noopTool()},
			})
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(out.Conversation[:len(inWhole)], inWhole) {
				t.Fatal("the turns of the conversation changed")
			}
			if got := len(out.Conversation); got != len(inWhole)+3 {
				t.Fatalf("conversation has %d turns, want %d", got, len(inWhole)+3)
			}
			if !reflect.DeepEqual(out.Conversation[len(out.Conversation)-3:], out.History[len(out.History)-3:]) {
				t.Fatal("both lists must end alike")
			}
			assertValid(t, out.Conversation)
			assertValid(t, out.History)
			if !hasText(fake.requests[len(fake.requests)-1], tc.wantRead) {
				t.Fatalf("the model does not read %q", tc.wantRead)
			}
			if hasText(fake.requests[0], "message number 0 ") {
				t.Fatal("the model read the turns which were summarized before")
			}
		})
	}
}

// After a compaction, a suspended run is continued or dismissed on both lists, which stay valid.
func TestContinue_AndDismiss_KeepBothListsValid(t *testing.T) {
	whole := longHistory(12)
	short := append([]Message{{Role: User, Content: []Content{Text{Text: "[Summary of earlier conversation]\nOLD"}}}}, slices.Clone(whole[6:])...)
	turn := Message{Role: Assistant, Content: []Content{askCall("q")}}
	cont := Continuation{Pending: []PendingCall{{Kind: PendingQuestion, Call: askCall("q"), Question: "which one?"}}}

	fake := &conversationFake{next: func(int) (Result, error) { return finalAnswer("done") }}
	out, err := Continue(nil, fake, RunOptions{
		Options:      Options{Messages: append(slices.Clone(short), turn)},
		Conversation: append(slices.Clone(whole), turn),
		Tools:        []Tool{NewAskUserTool()},
	}, cont, []Resolution{{CallID: "q", Answer: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	assertValid(t, out.History)
	assertValid(t, out.Conversation)
	if len(out.Conversation) != len(whole)+3 || len(out.History) != len(short)+3 {
		t.Fatalf("got %d and %d turns", len(out.Conversation), len(out.History))
	}
	if hasText(fake.requests[0], "message number 0 ") || !hasText(fake.requests[0], "OLD") {
		t.Fatal("the model must read the shortened version")
	}

	dismissed, err := Dismiss(append(slices.Clone(short), turn), cont)
	if err != nil {
		t.Fatal(err)
	}
	assertValid(t, dismissed)
}

// The compactions per run are limited by MaxCompactions; the limit is reached with the overflow after the last one.
func TestStart_MaxCompactions(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		overflows int
		wantErr   bool
	}{
		{name: "default exactly reached", max: 0, overflows: DefaultMaxCompactions, wantErr: false},
		{name: "default exceeded", max: 0, overflows: DefaultMaxCompactions + 1, wantErr: true},
		{name: "two exactly reached", max: 2, overflows: 2, wantErr: false},
		{name: "two exceeded", max: 2, overflows: 3, wantErr: true},
		{name: "32 exactly reached", max: 32, overflows: 32, wantErr: false},
		{name: "32 exceeded", max: 32, overflows: 33, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compactions := 0
			fake := &conversationFake{next: func(n int) (Result, error) {
				if n <= tc.overflows {
					return overflow()
				}
				return finalAnswer("done")
			}}

			// every compaction shrinks a list by one turn, so it can always go on
			_, err := Start(nil, fake, RunOptions{
				Options:        Options{Messages: longHistory(64)},
				MaxCompactions: tc.max,
				Compactor: func(_ context.Context, _ auth.Subject, _ Completions, _ Options, history []Message) ([]Message, error) {
					compactions++
					return history[1:], nil
				},
			})

			if got := err != nil; got != tc.wantErr {
				t.Fatalf("error = %v, want error %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, ContextWindowExceeded) {
				t.Fatalf("unexpected error: %v", err)
			}
			wantCompactions := tc.overflows
			if tc.wantErr {
				wantCompactions--
			}
			if compactions != wantCompactions {
				t.Fatalf("%d compactions, want %d", compactions, wantCompactions)
			}
		})
	}
}

// A failure after a compaction keeps the reached state, a failure of the first request has nothing to keep, and a
// cancelled compaction never replaces the history.
func TestStart_FailureAfterCompaction(t *testing.T) {
	boom := errors.New("provider down")

	tests := []struct {
		name           string
		next           func(n int) (Result, error)
		wantProgressed bool
		wantTurns      int
		wantBoth       bool
	}{
		{name: "fails without a compaction", next: func(int) (Result, error) { return Result{}, boom }},
		{name: "fails after a compaction", wantProgressed: true, wantBoth: true, next: func(n int) (Result, error) {
			if n == 1 {
				return overflow()
			}
			return Result{}, boom
		}},
		{name: "fails after an answer and a compaction", wantProgressed: true, wantBoth: true, wantTurns: 1, next: func(n int) (Result, error) {
			switch n {
			case 1:
				return noopCall("c1"), nil
			case 2:
				return overflow()
			}
			return Result{}, boom
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := longHistory(13)
			fake := &conversationFake{next: tc.next}
			out, err := Start(nil, fake, RunOptions{Options: Options{Messages: input}, Tools: []Tool{noopTool()}})
			if !errors.Is(err, boom) {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Progressed != tc.wantProgressed || out.Turns != tc.wantTurns || (out.Conversation != nil) != tc.wantBoth {
				t.Fatalf("progressed=%v turns=%d conversation=%v", out.Progressed, out.Turns, out.Conversation != nil)
			}
			if tc.wantBoth {
				assertValid(t, out.Conversation)
				assertValid(t, out.History)
				if len(out.Conversation) < len(input) || !reflect.DeepEqual(out.Conversation[:len(input)], input) {
					t.Fatal("the conversation must keep every turn")
				}
			}
		})
	}
}

// A compaction which is cancelled does not change any list.
func TestStart_CancelledCompactionKeepsTheConversation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fake := &conversationFake{next: func(int) (Result, error) { return overflow() }}
	input := longHistory(13)
	out, err := Start(nil, fake, RunOptions{
		Options: Options{Messages: input},
		Context: ctx,
		Compactor: func(context.Context, auth.Subject, Completions, Options, []Message) ([]Message, error) {
			cancel()
			return nil, ctx.Err()
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Progressed || out.Conversation != nil || !reflect.DeepEqual(out.History, input) {
		t.Fatalf("a cancelled compaction changed the lists: %+v", out)
	}
}

// A stop in the middle of the tool calls of a turn answers them in both lists, so that the next request is valid.
func TestStart_StopAfterCompactionAnswersRunningCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := NewTool("stop", "stops", func(struct{}) (struct{}, error) { cancel(); return struct{}{}, nil })
	fake := &conversationFake{next: func(n int) (Result, error) {
		if n == 1 {
			return overflow()
		}
		return toolUse(ToolCall{ID: "c1", Name: "stop", Arguments: json.RawMessage(`{}`)}, ToolCall{ID: "c2", Name: "stop", Arguments: json.RawMessage(`{}`)}), nil
	}}

	out, err := Start(nil, fake, RunOptions{Options: Options{Messages: longHistory(13)}, Tools: []Tool{stop}, Context: ctx})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValid(t, out.History)
	assertValid(t, out.Conversation)
	if !reflect.DeepEqual(out.History[len(out.History)-1], out.Conversation[len(out.Conversation)-1]) {
		t.Fatal("both lists must end alike")
	}
}

// A turn limit keeps the reached state of both lists; a final answer in the last allowed turn is no limit.
func TestStart_TurnLimitAfterCompaction(t *testing.T) {
	tests := []struct {
		name    string
		last    func() (Result, error)
		wantErr bool
	}{
		{name: "the model answers in the last turn", last: func() (Result, error) { return finalAnswer("done") }},
		{name: "the model goes on in the last turn", last: func() (Result, error) { return noopCall("c2"), nil }, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &conversationFake{next: func(n int) (Result, error) {
				switch n {
				case 1:
					return overflow()
				case 2:
					return noopCall("c1"), nil
				}
				return tc.last()
			}}

			input := longHistory(13)
			out, err := Start(nil, fake, RunOptions{Options: Options{Messages: input}, Tools: []Tool{noopTool()}, MaxTurns: 2})
			if got := errors.Is(err, TurnLimitExceeded); got != tc.wantErr {
				t.Fatalf("limit = %v (%v), want %v", got, err, tc.wantErr)
			}
			if out.Turns != 2 {
				t.Fatalf("%d turns", out.Turns)
			}
			if len(out.Conversation) <= len(input) {
				t.Fatalf("the steps are missing: %d turns", len(out.Conversation))
			}
			assertValid(t, out.Conversation)
			assertValid(t, out.History)
		})
	}
}

// The hidden turns of the loop are not written by the user.
func TestIsLoopPrompt(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{TurnLimitPrompt, true},
		{continuePrompt, true},
		{"Bitte weitermachen", false},
		{fmt.Sprintf("%s ", TurnLimitPrompt), false},
	}

	for _, tc := range tests {
		if got := IsLoopPrompt(Message{Role: User, Content: []Content{Text{Text: tc.text}}}); got != tc.want {
			t.Errorf("IsLoopPrompt(%q) = %v", tc.text, got)
		}
	}
	if IsLoopPrompt(Message{Role: Assistant, Content: []Content{Text{Text: TurnLimitPrompt}}}) {
		t.Error("an answer of the model is no prompt of the loop")
	}
}

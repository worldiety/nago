// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// runFake records every request to the model and answers it with next, called with the number of the request; a
// request to summarize is answered by itself and neither recorded nor counted.
type runFake struct {
	mu       sync.Mutex
	requests [][]completion.Message
	next     func(n int) (completion.Result, error)
}

func (f *runFake) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *runFake) Complete(_ context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	if strings.HasPrefix(opts.System, "You are compacting") {
		return answerResult("SUMMARY"), nil
	}
	f.mu.Lock()
	f.requests = append(f.requests, slices.Clone(opts.Messages))
	n := len(f.requests)
	f.mu.Unlock()
	return f.next(n)
}

func (f *runFake) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func (f *runFake) last() []completion.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[len(f.requests)-1]
}

func (f *runFake) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func overflowResult() (completion.Result, error) {
	return completion.Result{}, completion.ContextWindowError{Limit: 100, Tokens: 200}
}

func toolCallResult(id string) completion.Result {
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{ID: id, Name: "noop", Arguments: stdjson.RawMessage(`{}`)}}},
		StopReason: completion.StopToolUse,
	}
}

func noop() completion.Tool {
	return completion.NewTool("noop", "does nothing", func(struct{}) (struct{}, error) { return struct{}{}, nil })
}

func userText(text string) []completion.Content {
	return []completion.Content{completion.Text{Text: text}}
}

func hasText(msgs []completion.Message, want string) bool {
	return slices.ContainsFunc(msgs, func(m completion.Message) bool {
		return slices.ContainsFunc(m.Content, func(c completion.Content) bool {
			t, ok := c.(completion.Text)
			return ok && strings.Contains(t.Text, want)
		})
	})
}

// assertRequestValid checks what a provider demands: every tool call has exactly one result and every result a call.
func assertRequestValid(t *testing.T, msgs []completion.Message) {
	t.Helper()
	calls, results := map[string]int{}, map[string]int{}
	for _, m := range msgs {
		for _, c := range m.Content {
			switch v := c.(type) {
			case completion.ToolCall:
				calls[v.ID]++
			case completion.ToolResult:
				results[v.ToolCallID]++
			}
		}
	}
	if !reflect.DeepEqual(calls, results) {
		t.Fatalf("invalid request: calls %v, results %v", calls, results)
	}
}

// longMessages is an old conversation; its first turn has a medium.
func longMessages(n int) []completion.Message {
	msgs := make([]completion.Message, 0, n)
	for i := range n {
		role := completion.User
		if i%2 == 1 {
			role = completion.Assistant
		}
		msgs = append(msgs, completion.Message{Role: role, Content: userText(fmt.Sprintf("old turn %d with some filler", i))})
	}
	msgs[0].Content = append(msgs[0].Content, completion.Media{MimeType: file.Type("image/png"), Source: completion.Source{Data: []byte("png")}})
	return msgs
}

// compactedSession stores a session which has been compacted by a run: it ends with a question for the model.
func compactedSession(t *testing.T) (UseCases, Session, []completion.Message) {
	t.Helper()
	uc, repo, _ := newTestUseCases(t)
	s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}
	s.Messages = longMessages(12)
	if err := repo.Save(s); err != nil {
		t.Fatal(err)
	}

	fake := &runFake{next: func(n int) (completion.Result, error) {
		if n == 1 {
			return overflowResult()
		}
		return answerResult("first answer"), nil
	}}
	s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("first question"), Tools: []completion.Tool{noop()}})
	if err != nil {
		t.Fatal(err)
	}

	if !hasText(s.Compacted, "SUMMARY") || hasText(s.Compacted, "old turn 0 ") {
		t.Fatalf("the compaction is missing: %+v", s.Compacted)
	}
	return uc, s, slices.Clone(s.Messages)
}

// A compacted conversation stays complete with its media, while each further way of going on sends the shortened
// version to the model and stays valid.
func TestCompaction_KeepsTheConversation(t *testing.T) {
	ask := completion.NewAskUserTool()

	tests := []struct {
		name string
		// goOn continues the compacted session and returns the session and the request it made
		goOn func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message)
		// wantAdded are the turns the conversation got
		wantAdded int
	}{
		{
			name:      "a further message with tools",
			wantAdded: 2,
			goOn: func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message) {
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("second answer"), nil }}
				s, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("second"), Tools: []completion.Tool{noop()}})
				if err != nil {
					t.Fatal(err)
				}
				return s, fake.last()
			},
		},
		{
			name:      "a further message without tools",
			wantAdded: 2,
			goOn: func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message) {
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("second answer"), nil }}
				s, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("second")})
				if err != nil {
					t.Fatal(err)
				}
				return s, fake.last()
			},
		},
		{
			name:      "stop, then a message",
			wantAdded: 5,
			goOn: func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				stopper := completion.NewTool("noop", "stops", func(struct{}) (struct{}, error) { cancel(); return struct{}{}, nil })
				running := &runFake{next: func(int) (completion.Result, error) {
					res := toolCallResult("c1")
					res.Message.Content = append(res.Message.Content, completion.ToolCall{ID: "c2", Name: "noop", Arguments: stdjson.RawMessage(`{}`)})
					return res, nil
				}}
				_, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: running, Input: userText("stop me"), Tools: []completion.Tool{stopper}, Context: ctx})
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected a stop, got %v", err)
				}
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("after stop"), nil }}
				s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("second"), Tools: []completion.Tool{noop()}})
				if err != nil {
					t.Fatal(err)
				}
				return s, fake.last()
			},
		},
		{
			name:      "the answer to a question",
			wantAdded: 4,
			goOn: func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message) {
				asking := &runFake{next: func(int) (completion.Result, error) { return askResult(), nil }}
				s, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: asking, Input: userText("ask me"), Tools: []completion.Tool{ask}})
				if err != nil {
					t.Fatal(err)
				}
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("thanks"), nil }}
				s, err = uc.Resolve(user.SU(), s.ID, ResolveOptions{
					Run:         AppendOptions{Completions: fake, Tools: []completion.Tool{ask}},
					Revision:    s.PendingRevision,
					Resolutions: []completion.Resolution{{CallID: "q1", Answer: "this"}},
				})
				if err != nil {
					t.Fatal(err)
				}
				return s, fake.last()
			},
		},
		{
			name:      "a dismissed question, then a message",
			wantAdded: 5,
			goOn: func(t *testing.T, uc UseCases, s Session) (Session, []completion.Message) {
				asking := &runFake{next: func(int) (completion.Result, error) { return askResult(), nil }}
				s, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: asking, Input: userText("ask me"), Tools: []completion.Tool{ask}})
				if err != nil {
					t.Fatal(err)
				}
				s, err = uc.Dismiss(user.SU(), s.ID, s.PendingRevision)
				if err != nil {
					t.Fatal(err)
				}
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("after"), nil }}
				s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("second"), Tools: []completion.Tool{ask}})
				if err != nil {
					t.Fatal(err)
				}
				return s, fake.last()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, s, before := compactedSession(t)
			got, request := tc.goOn(t, uc, s)

			if len(got.Messages) != len(before)+tc.wantAdded || !reflect.DeepEqual(got.Messages[:len(before)], before) {
				t.Fatalf("the conversation has %d turns, want %d and every former turn", len(got.Messages), len(before)+tc.wantAdded)
			}
			if !slices.ContainsFunc(got.Messages[0].Content, func(c completion.Content) bool { _, ok := c.(completion.Media); return ok }) {
				t.Fatal("the medium of the first turn is gone")
			}
			if hasText(got.Messages, "SUMMARY") {
				t.Fatal("a summary must not appear in the conversation")
			}

			if !hasText(request, "SUMMARY") || hasText(request, "old turn 0 ") {
				t.Fatalf("the request must hold the summary instead of the old turns: %+v", request)
			}
			assertRequestValid(t, request)

			if !reflect.DeepEqual(got.Messages[len(got.Messages)-1], got.Compacted[len(got.Compacted)-1]) {
				t.Fatal("both lists must end alike")
			}

			stored, err := uc.FindByID(user.SU(), got.ID)
			if err != nil || stored.IsNone() {
				t.Fatal(err)
			}
			if s := stored.Unwrap(); len(s.Messages) != len(got.Messages) || !hasText(s.Compacted, "SUMMARY") {
				t.Fatalf("both lists must be stored: %d, %d", len(s.Messages), len(s.Compacted))
			}
		})
	}
}

// Without a compaction there is one list. A conversation compacted before there were two lists stays as it is and
// goes on, a stored one without the new fields loads.
func TestCompaction_OneListWithoutACompaction(t *testing.T) {
	uc, repo, _ := newTestUseCases(t)
	s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("answer"), nil }}
	s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("question"), Tools: []completion.Tool{noop()}})
	if err != nil {
		t.Fatal(err)
	}
	if s.Compacted != nil || len(s.Messages) != 2 {
		t.Fatalf("expected one list, got %d and %d", len(s.Messages), len(s.Compacted))
	}

	buf, err := stdjson.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(buf), "compacted") || strings.Contains(string(buf), "turnLimited") {
		t.Fatalf("the new fields must be optional: %s", buf)
	}

	// a former compaction left its summary in Messages
	s.Messages = []completion.Message{{Role: completion.User, Content: userText("[Summary of earlier conversation]\nformer")}}
	if err := repo.Save(s); err != nil {
		t.Fatal(err)
	}
	s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("next"), Tools: []completion.Tool{noop()}})
	if err != nil {
		t.Fatal(err)
	}
	if !hasText(s.Messages, "former") || !hasText(fake.last(), "former") || s.Compacted != nil {
		t.Fatalf("a former compaction must stay as it is: %+v", s)
	}
}

// A sub-agent keeps summarizing in its own conversation.
func TestCompaction_ChildSessionKeepsTheShortVersion(t *testing.T) {
	tests := []struct {
		name        string
		child       bool
		wantSummary bool
	}{
		{name: "a child session", child: true, wantSummary: true},
		{name: "a conversation of the user"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, repo, _ := newTestUseCases(t)
			var parent ID
			if tc.child {
				p, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
				if err != nil {
					t.Fatal(err)
				}
				parent = p.ID
			}
			s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model", ParentID: parent})
			if err != nil {
				t.Fatal(err)
			}
			s.Messages = longMessages(12)
			if err := repo.Save(s); err != nil {
				t.Fatal(err)
			}

			fake := &runFake{next: func(n int) (completion.Result, error) {
				if n == 1 {
					return overflowResult()
				}
				return answerResult("done"), nil
			}}
			s, err = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("go"), Agentic: true})
			if err != nil {
				t.Fatal(err)
			}
			if got := hasText(s.Messages, "SUMMARY"); got != tc.wantSummary {
				t.Fatalf("summary in Messages = %v, want %v", got, tc.wantSummary)
			}
			if got := len(s.Compacted) > 0; got == tc.wantSummary {
				t.Fatalf("Compacted set = %v", got)
			}
		})
	}
}

// MaxCompactions reaches the loop of Append and Resolve; the limit is reached with the overflow after the last one.
func TestMaxCompactions_ReachesTheRun(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		overflows int
		wantErr   bool
	}{
		{name: "default exactly reached", overflows: completion.DefaultMaxCompactions},
		{name: "default exceeded", overflows: completion.DefaultMaxCompactions + 1, wantErr: true},
		{name: "two exceeded", max: 2, overflows: 3, wantErr: true},
		{name: "six exactly reached", max: 6, overflows: 6},
	}

	run := map[string]func(t *testing.T, uc UseCases, s Session, fake *runFake, max int) error{
		"append": func(t *testing.T, uc UseCases, s Session, fake *runFake, max int) error {
			_, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("go"), Tools: []completion.Tool{noop()}, MaxCompactions: max})
			return err
		},
		"resolve": func(t *testing.T, uc UseCases, s Session, fake *runFake, max int) error {
			asking := &runFake{next: func(int) (completion.Result, error) { return askResult(), nil }}
			ask := []completion.Tool{completion.NewAskUserTool(), noop()}
			s, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: asking, Input: userText("ask me"), Tools: ask})
			if err != nil {
				t.Fatal(err)
			}
			_, err = uc.Resolve(user.SU(), s.ID, ResolveOptions{
				Run:         AppendOptions{Completions: fake, Tools: ask, MaxCompactions: max},
				Revision:    s.PendingRevision,
				Resolutions: []completion.Resolution{{CallID: "q1", Answer: "this"}},
			})
			return err
		},
	}

	for way, fn := range run {
		for _, tc := range tests {
			t.Run(way+"/"+tc.name, func(t *testing.T) {
				uc, repo, _ := newTestUseCases(t)
				s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
				if err != nil {
					t.Fatal(err)
				}
				s.Messages = longMessages(400)
				if err := repo.Save(s); err != nil {
					t.Fatal(err)
				}

				// a step between two overflows gives the next compaction something to summarize
				fake := &runFake{next: func(n int) (completion.Result, error) {
					switch {
					case n > 2*tc.overflows:
						return answerResult("done"), nil
					case n%2 == 1:
						return overflowResult()
					}
					return toolCallResult(fmt.Sprintf("c%d", n)), nil
				}}
				err = fn(t, uc, s, fake, tc.max)
				if got := errors.Is(err, completion.ContextWindowExceeded); got != tc.wantErr {
					t.Fatalf("limit reached = %v (%v), want %v", got, err, tc.wantErr)
				}
			})
		}
	}
}

// turnLimitedSession stores a session whose run ended at its turn limit.
func turnLimitedSession(t *testing.T) (UseCases, Session) {
	t.Helper()
	uc, _, _ := newTestUseCases(t)
	s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &runFake{next: func(n int) (completion.Result, error) { return toolCallResult(fmt.Sprintf("c%d", n)), nil }}
	got, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText(strings.Repeat("a long task ", 20)), Tools: []completion.Tool{noop()}, MaxTurns: 4})
	if !errors.Is(err, completion.TurnLimitExceeded) {
		t.Fatalf("expected the limit, got %v", err)
	}
	if got.ID != s.ID || !got.TurnLimited || got.PendingRevision != 1 || len(got.Messages) != 9 {
		t.Fatalf("the failed run must return what it kept: %+v", got)
	}
	if fake.count() != 4 {
		t.Fatalf("%d requests, want exactly 4", fake.count())
	}
	return uc, got
}

// The hint of a turn limit is stored and ends with the first answer of the model in the next run. A run which fails
// before that does not change it.
func TestTurnLimited(t *testing.T) {
	boom := errors.New("provider down")
	tools := []completion.Tool{noop()}

	tests := []struct {
		name string
		next func(n int) (completion.Result, error)
		// wantErr is the failure of the run, wantLimited the hint afterwards, wantKept tells whether the run returns a
		// session
		wantErr     error
		wantLimited bool
		wantKept    bool
	}{
		{
			name:     "the next run answers",
			next:     func(int) (completion.Result, error) { return answerResult("done"), nil },
			wantKept: true,
		},
		{
			name:        "the next run fails before an answer",
			next:        func(int) (completion.Result, error) { return completion.Result{}, boom },
			wantErr:     boom,
			wantLimited: true,
		},
		{
			name: "the next run fails after a compaction before an answer",
			next: func(n int) (completion.Result, error) {
				if n == 1 {
					return overflowResult()
				}
				return completion.Result{}, boom
			},
			wantErr:     boom,
			wantLimited: true,
			wantKept:    true,
		},
		{
			name: "the next run fails after an answer",
			next: func(n int) (completion.Result, error) {
				if n == 1 {
					return toolCallResult("n1"), nil
				}
				return completion.Result{}, boom
			},
			wantErr:  boom,
			wantKept: true,
		},
		{
			name:        "the next run ends at the limit again",
			next:        func(n int) (completion.Result, error) { return toolCallResult(fmt.Sprintf("n%d", n)), nil },
			wantErr:     completion.TurnLimitExceeded,
			wantLimited: true,
			wantKept:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, s := turnLimitedSession(t)
			fake := &runFake{next: tc.next}
			got, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("next"), Tools: tools, MaxTurns: 2})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
			if (got.ID != "") != tc.wantKept {
				t.Fatalf("returned session %q, want one: %v", got.ID, tc.wantKept)
			}

			stored, ferr := uc.FindByID(user.SU(), s.ID)
			if ferr != nil || stored.IsNone() {
				t.Fatal(ferr)
			}
			if got := stored.Unwrap().TurnLimited; got != tc.wantLimited {
				t.Fatalf("TurnLimited = %v, want %v", got, tc.wantLimited)
			}
		})
	}
}

// A final answer in the last allowed turn is no turn limit.
func TestTurnLimited_NotWithAnAnswerInTheLastTurn(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	s, err := uc.Create(user.SU(), CreateOptions{Model: "fake-model"})
	if err != nil {
		t.Fatal(err)
	}

	fake := &runFake{next: func(n int) (completion.Result, error) {
		if n == 2 {
			return answerResult("done"), nil
		}
		return toolCallResult("c1"), nil
	}}
	got, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText("go"), Tools: []completion.Tool{noop()}, MaxTurns: 2})
	if err != nil || got.TurnLimited {
		t.Fatalf("expected a regular end, got %v, %+v", err, got)
	}
}

// Continuing needs a run ended at its limit with exactly the named revision.
func TestContinueRevision(t *testing.T) {
	tests := []struct {
		name     string
		revision func(s Session) int
		// prepare runs before the continuation
		prepare func(t *testing.T, uc UseCases, s Session)
		wantErr error
	}{
		{name: "the named revision", revision: func(s Session) int { return s.PendingRevision }},
		{name: "a stale revision", revision: func(s Session) int { return s.PendingRevision + 1 }, wantErr: ErrNoTurnLimit},
		{
			name:     "already continued",
			revision: func(s Session) int { return s.PendingRevision },
			prepare: func(t *testing.T, uc UseCases, s Session) {
				fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("first"), nil }}
				if _, err := uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText(completion.TurnLimitPrompt), ContinueRevision: s.PendingRevision, Tools: []completion.Tool{noop()}}); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: ErrNoTurnLimit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, s := turnLimitedSession(t)
			if tc.prepare != nil {
				tc.prepare(t, uc, s)
			}

			fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("done"), nil }}
			got, err := uc.Append(user.SU(), s.ID, AppendOptions{
				Completions:      fake,
				Input:            userText(completion.TurnLimitPrompt),
				ContinueRevision: tc.revision(s),
				Tools:            []completion.Tool{noop()},
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantErr != nil {
				if fake.count() != 0 {
					t.Fatalf("a refused continuation must not run, got %d requests", fake.count())
				}
				return
			}

			if len(fake.last()) != len(s.Messages)+1 {
				t.Fatalf("the first request must hold the kept conversation and the prompt: %d turns", len(fake.last()))
			}
			if !completion.IsLoopPrompt(fake.last()[len(fake.last())-1]) {
				t.Fatalf("the last turn of the request must be the prompt: %+v", fake.last())
			}
			if got.TurnLimited {
				t.Fatal("the hint must end")
			}
		})
	}
}

// Two tabs continuing the same limit start exactly one run.
func TestContinueRevision_TwoTabsStartOneRun(t *testing.T) {
	uc, s := turnLimitedSession(t)
	fake := &runFake{next: func(int) (completion.Result, error) { return answerResult("done"), nil }}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Go(func() {
			_, errs[i] = uc.Append(user.SU(), s.ID, AppendOptions{Completions: fake, Input: userText(completion.TurnLimitPrompt), ContinueRevision: s.PendingRevision, Tools: []completion.Tool{noop()}})
		})
	}
	wg.Wait()

	if fake.count() != 1 {
		t.Fatalf("%d runs, want one", fake.count())
	}
	if !(errs[0] == nil && errors.Is(errs[1], ErrNoTurnLimit)) && !(errs[1] == nil && errors.Is(errs[0], ErrNoTurnLimit)) {
		t.Fatalf("exactly one must be refused: %v", errs)
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

const limitHint = "Höchstzahl der Schritte"

// recording records every request to the model and answers it with next, called with the number of the request. A
// request to summarize or to name the conversation is answered by itself and neither recorded nor counted.
type recording struct {
	mu       sync.Mutex
	requests [][]completion.Message
	next     func(ctx context.Context, n int) (completion.Result, error)
}

func (f *recording) Models(auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) { yield(model.Model{ID: "fake"}, nil) }
}

func (f *recording) Complete(ctx context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	if strings.HasPrefix(opts.System, "You are compacting") {
		return text("SUMMARY"), nil
	}
	if opts.MaxTokens == 64 {
		return text("Titel"), nil
	}

	f.mu.Lock()
	f.requests = append(f.requests, slices.Clone(opts.Messages))
	n := len(f.requests)
	f.mu.Unlock()
	return f.next(ctx, n)
}

func (f *recording) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func (f *recording) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func (f *recording) last() []completion.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[len(f.requests)-1]
}

func overflowed() (completion.Result, error) {
	return completion.Result{}, completion.ContextWindowError{Limit: 100, Tokens: 200}
}

func step(id string) completion.Result { return call(id, "noop", `{}`) }

func stepWithText(id, label string) completion.Result {
	r := step(id)
	r.Message.Content = append([]completion.Content{completion.Text{Text: label}}, r.Message.Content...)
	return r
}

func noopAgent() []uicompletion.Agent {
	return []uicompletion.Agent{{Tools: []completion.Tool{
		completion.NewTool("noop", "does nothing", func(struct{}) (struct{}, error) { return struct{}{}, nil }),
	}}}
}

func limitChat(t *testing.T, comps completion.Completions, history bool, maxTurns int) (*nagotest.Window, session.UseCases, uicompletion.ChatOptions) {
	t.Helper()
	sessions := testSessions(t)
	opts := uicompletion.ChatOptions{
		Sessions:           sessions,
		History:            history,
		Completions:        comps,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		MaxTurns:           maxTurns,
		Agents:             noopAgent(),
		AskUser:            true,
	}
	return openChatOptions(t, opts), sessions, opts
}

func ask(w *nagotest.Window, question string) {
	w.Type(w.Find(nagotest.Label("Nachricht")), question)
	w.Click(w.Find(nagotest.Text("Senden")))
}

func waitIdle(w *nagotest.Window) { w.WaitFor(nagotest.Text("Senden"), 10*time.Second) }

func hasRequestText(msgs []completion.Message, want string) bool {
	return slices.ContainsFunc(msgs, func(m completion.Message) bool {
		return slices.ContainsFunc(m.Content, func(c completion.Content) bool {
			t, ok := c.(completion.Text)
			return ok && strings.Contains(t.Text, want)
		})
	})
}

// fieldWith matches a text field whose value contains want.
func fieldWith(want string) nagotest.Matcher {
	return nagotest.Where("TextField("+want+")", func(n nagotest.Node) bool {
		f, ok := n.Component.(*proto.TextField)
		return ok && strings.Contains(string(f.Value), want)
	})
}

// errorLog collects the records of the level error while a test runs.
type errorLog struct {
	slog.Handler
	mu      *sync.Mutex
	records *[]string
}

func (l errorLog) Enabled(_ context.Context, level slog.Level) bool { return level >= slog.LevelError }

func (l errorLog) Handle(_ context.Context, r slog.Record) error {
	text := r.Message
	r.Attrs(func(a slog.Attr) bool {
		text += " " + a.Value.String()
		return true
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	*l.records = append(*l.records, text)
	return nil
}

func captureErrors(t *testing.T) func() []string {
	t.Helper()
	log := errorLog{Handler: slog.DiscardHandler, mu: &sync.Mutex{}, records: &[]string{}}
	prev := slog.Default()
	slog.SetDefault(slog.New(log))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return func() []string {
		log.mu.Lock()
		defer log.mu.Unlock()
		return slices.Clone(*log.records)
	}
}

// A run at its limit keeps its steps and offers to continue, without any failure. An answer in the last allowed turn is
// no limit, and a failing provider stays a failure.
func TestChat_TurnLimit_ShowsAHintInsteadOfAFailure(t *testing.T) {
	boom := errors.New("provider down")

	tests := []struct {
		name     string
		last     func() (completion.Result, error)
		wantHint bool
		wantLog  bool
	}{
		{name: "the model goes on in the last turn", last: func() (completion.Result, error) { return stepWithText("c2", "Schritt zwei"), nil }, wantHint: true},
		{name: "the model answers in the last turn", last: func() (completion.Result, error) { return text("Alles fertig."), nil }},
		{name: "the provider fails", last: func() (completion.Result, error) { return completion.Result{}, boom }, wantLog: true},
	}

	for _, history := range []bool{false, true} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s/history=%v", tc.name, history), func(t *testing.T) {
				logged := captureErrors(t)
				fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
					if n == 1 {
						return stepWithText("c1", "Schritt eins"), nil
					}
					return tc.last()
				}}

				w, _, _ := limitChat(t, fake, history, 2)
				ask(w, "Bitte arbeiten")
				waitIdle(w)

				w.Find(richText("Schritt eins")).Exactly(1)
				if got := w.FindAll(nagotest.TextContains(limitHint)).Len() == 1; got != tc.wantHint {
					t.Fatalf("hint shown = %v, want %v", got, tc.wantHint)
				}
				if got := w.FindAll(nagotest.Text("Fortsetzen")).Len() == 1; got != tc.wantHint {
					t.Fatalf("button shown = %v, want %v", got, tc.wantHint)
				}
				if got := len(logged()) > 0; got != tc.wantLog {
					t.Fatalf("error logged = %v, want %v: %v", got, tc.wantLog, logged())
				}
				if fake.count() != 2 {
					t.Fatalf("%d requests, want exactly 2", fake.count())
				}
			})
		}
	}
}

// "Fortsetzen" starts a new run on the kept conversation with the full number of turns, and the hint is gone. The
// prompt which continues is not shown.
func TestChat_TurnLimit_Continue(t *testing.T) {
	for _, history := range []bool{false, true} {
		t.Run(fmt.Sprintf("history=%v", history), func(t *testing.T) {
			fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
				switch n {
				case 3:
					return text("Alles fertig."), nil
				default:
					return step(fmt.Sprintf("c%d", n)), nil
				}
			}}

			w, _, _ := limitChat(t, fake, history, 2)
			ask(w, "Bitte arbeiten")
			w.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)
			kept := len(fake.last()) + 2

			w.Click(w.Find(nagotest.Text("Fortsetzen")))
			w.WaitFor(richText("Alles fertig."), 10*time.Second)

			w.FindAll(nagotest.TextContains(limitHint)).None()
			w.FindAll(nagotest.Text("Fortsetzen")).None()
			w.FindAll(richText(completion.TurnLimitPrompt)).None()

			first := fake.last()
			if len(first) != kept+1 {
				t.Fatalf("the request holds %d turns, want the %d kept and the prompt", len(first), kept)
			}
			if !completion.IsLoopPrompt(first[len(first)-1]) {
				t.Fatalf("the last turn must be the prompt, got %+v", first[len(first)-1])
			}
		})
	}
}

// The hint ends with the next message and with a new chat; opened from the history, it is back.
func TestChat_TurnLimit_HintEnds(t *testing.T) {
	var answer atomicAnswer
	fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
		if a := answer.get(); a != "" {
			return text(a), nil
		}
		return step(fmt.Sprintf("c%d", n)), nil
	}}

	w, _, _ := limitChat(t, fake, true, 1)
	ask(w, "Bitte arbeiten")
	w.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)

	w.Click(w.Find(nagotest.Text("Neuer Chat")))
	w.FindAll(nagotest.TextContains(limitHint)).None()

	w.Click(w.Find(nagotest.Text("Verlauf")))
	w.WaitFor(nagotest.TextContains("Nachrichten"), 5*time.Second)
	w.Click(w.Find(nagotest.Text("Bitte arbeiten")))
	w.WaitFor(nagotest.TextContains(limitHint), 5*time.Second)

	answer.set("Antwort auf die Nachricht")
	ask(w, "Andere Nachricht")
	w.WaitFor(richText("Antwort auf die Nachricht"), 10*time.Second)
	w.FindAll(nagotest.TextContains(limitHint)).None()
	last := fake.last()
	if completion.IsLoopPrompt(last[len(last)-1]) || !hasRequestText(last, "Andere Nachricht") {
		t.Fatalf("the request must carry the message: %+v", last)
	}
}

type atomicAnswer struct {
	mu    sync.Mutex
	value string
}

func (a *atomicAnswer) get() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.value
}

func (a *atomicAnswer) set(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.value = v
}

// A failure before the model answered leaves the hint, a failure after an answer ends it.
func TestChat_TurnLimit_ContinueFails(t *testing.T) {
	boom := errors.New("provider down")
	tests := []struct {
		name     string
		next     func(n int) (completion.Result, error)
		wantHint bool
	}{
		{name: "before the model answered", wantHint: true, next: func(int) (completion.Result, error) { return completion.Result{}, boom }},
		{name: "after the model answered", next: func(n int) (completion.Result, error) {
			if n == 1 {
				return step("n1"), nil
			}
			return completion.Result{}, boom
		}},
	}

	for _, history := range []bool{false, true} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s/history=%v", tc.name, history), func(t *testing.T) {
				logged := captureErrors(t)
				var offset, limit int
				var mu sync.Mutex
				fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
					mu.Lock()
					defer mu.Unlock()
					if limit == 0 {
						return step(fmt.Sprintf("c%d", n)), nil
					}
					return tc.next(n - offset)
				}}

				w, _, _ := limitChat(t, fake, history, 2)
				ask(w, "Bitte arbeiten")
				w.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)
				mu.Lock()
				offset, limit = fake.count(), 1
				mu.Unlock()

				w.Click(w.Find(nagotest.Text("Fortsetzen")))
				waitIdle(w)

				if len(logged()) == 0 {
					t.Fatal("the failure must be reported as a failure")
				}
				if got := w.FindAll(nagotest.Text("Fortsetzen")).Len() == 1; got != tc.wantHint {
					t.Fatalf("button shown = %v, want %v", got, tc.wantHint)
				}
			})
		}
	}
}

// A second click and a second tab start exactly one run.
func TestChat_TurnLimit_ContinueStartsOneRun(t *testing.T) {
	t.Run("a working run offers no second click", func(t *testing.T) {
		gate := make(chan struct{})
		fake := &recording{next: func(ctx context.Context, n int) (completion.Result, error) {
			if n <= 2 {
				return step(fmt.Sprintf("c%d", n)), nil
			}
			select {
			case <-gate:
			case <-ctx.Done():
				return completion.Result{}, ctx.Err()
			}
			return text("Alles fertig."), nil
		}}
		w, _, _ := limitChat(t, fake, false, 2)
		ask(w, "Bitte arbeiten")
		w.Click(w.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second))
		w.WaitFor(nagotest.Text("Stopp"), 5*time.Second)
		w.FindAll(nagotest.Text("Fortsetzen")).None()
		w.FindAll(nagotest.TextContains(limitHint)).None()
		close(gate)
		w.WaitFor(richText("Alles fertig."), 10*time.Second)
		if fake.count() != 3 {
			t.Fatalf("%d requests, want 3", fake.count())
		}
	})

	t.Run("a second tab while the first one works", func(t *testing.T) {
		gate := make(chan struct{})
		fake := &recording{next: func(ctx context.Context, n int) (completion.Result, error) {
			if n <= 2 {
				return step(fmt.Sprintf("c%d", n)), nil
			}
			select {
			case <-gate:
			case <-ctx.Done():
				return completion.Result{}, ctx.Err()
			}
			return text("Alles fertig."), nil
		}}

		sessions := testSessions(t)
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.uicompletion.test")
			cfg.RootView("chat", func(wnd core.Window) core.View {
				return uicompletion.Chat(wnd, uicompletion.ChatOptions{
					Sessions: sessions, History: true, Completions: fake, Provider: echo.New("p", "p"),
					DisableCurrentTime: true, MaxTurns: 2, Agents: noopAgent(),
				})
			})
		})

		first := app.Open(t, user.SU(), "chat")
		ask(first, "Bitte arbeiten")
		first.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)

		second := app.Open(t, user.SU(), "chat")
		second.Click(second.Find(nagotest.Text("Verlauf")))
		second.WaitFor(nagotest.TextContains("Nachrichten"), 5*time.Second)
		second.Click(second.Find(nagotest.Text("Bitte arbeiten")))
		second.WaitFor(nagotest.Text("Fortsetzen"), 5*time.Second)

		first.Click(first.Find(nagotest.Text("Fortsetzen")))
		first.WaitFor(nagotest.Text("Stopp"), 5*time.Second)
		for fake.count() < 3 {
			time.Sleep(time.Millisecond)
		}

		// The second tab still shows the offer and joins the run instead of starting one. Its window settles only with
		// the end of that run, so the run is held until the tab shows it.
		button := second.Find(nagotest.Text("Fortsetzen"))
		clicked := make(chan struct{})
		go func() {
			defer close(clicked)
			second.Click(button)
		}()
		for second.FindAll(nagotest.Text("Stopp")).Len() == 0 {
			time.Sleep(time.Millisecond)
		}

		close(gate)
		<-clicked
		first.WaitFor(richText("Alles fertig."), 10*time.Second)
		second.WaitFor(richText("Alles fertig."), 10*time.Second)
		if fake.count() != 3 {
			t.Fatalf("%d requests, want 3", fake.count())
		}
	})

	t.Run("a second tab after the first one finished", func(t *testing.T) {
		logged := captureErrors(t)
		fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
			if n <= 2 {
				return step(fmt.Sprintf("c%d", n)), nil
			}
			return text("Alles fertig."), nil
		}}

		sessions := testSessions(t)
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.uicompletion.test")
			cfg.RootView("chat", func(wnd core.Window) core.View {
				return uicompletion.Chat(wnd, uicompletion.ChatOptions{
					Sessions: sessions, History: true, Completions: fake, Provider: echo.New("p", "p"),
					DisableCurrentTime: true, MaxTurns: 2, Agents: noopAgent(),
				})
			})
		})

		first := app.Open(t, user.SU(), "chat")
		ask(first, "Bitte arbeiten")
		first.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)

		second := app.Open(t, user.SU(), "chat")
		second.Click(second.Find(nagotest.Text("Verlauf")))
		second.WaitFor(nagotest.TextContains("Nachrichten"), 5*time.Second)
		second.Click(second.Find(nagotest.Text("Bitte arbeiten")))
		second.WaitFor(nagotest.Text("Fortsetzen"), 5*time.Second)

		first.Click(first.Find(nagotest.Text("Fortsetzen")))
		first.WaitFor(richText("Alles fertig."), 10*time.Second)

		second.Click(second.Find(nagotest.Text("Fortsetzen")))
		second.WaitFor(richText("Alles fertig."), 10*time.Second)
		second.FindAll(nagotest.Text("Fortsetzen")).None()
		if fake.count() != 3 {
			t.Fatalf("%d requests, want 3", fake.count())
		}
		if len(logged()) != 0 {
			t.Fatalf("the stale click is no failure: %v", logged())
		}
	})
}

// A kept conversation shows the hint on the next page, whose chat continues the conversation of its window state.
func TestChat_TurnLimit_FollowsToTheNextPage(t *testing.T) {
	fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
		return step(fmt.Sprintf("c%d", n)), nil
	}}

	sessions := testSessions(t)
	var keptMu sync.Mutex
	var kept session.ID
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.test")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			state := core.AutoState[session.ID](wnd).Init(func() session.ID {
				keptMu.Lock()
				defer keptMu.Unlock()
				return kept
			})
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions: sessions, History: true, Completions: fake, Provider: echo.New("p", "p"),
				DisableCurrentTime: true, MaxTurns: 2, Agents: noopAgent(), Session: state,
			})
		})
	})

	first := app.Open(t, user.SU(), "chat")
	ask(first, "Bitte arbeiten")
	first.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)
	for s, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		keptMu.Lock()
		kept = s.ID
		keptMu.Unlock()
	}

	// a new mount, as after a navigation
	next := app.Open(t, user.SU(), "chat")
	next.Find(nagotest.TextContains(limitHint)).Exactly(1)
	next.Find(nagotest.Text("Fortsetzen")).Exactly(1)

	// a window without the kept id starts empty, the conversation is found over the history
	keptMu.Lock()
	kept = ""
	keptMu.Unlock()
	empty := app.Open(t, user.SU(), "chat")
	empty.FindAll(nagotest.TextContains(limitHint)).None()
}

// The conversation shown stays whole after a compaction, while every further request holds the shortened version; in
// a transient chat as in a persisted one.
func TestChat_Compaction_ShowsTheWholeConversation(t *testing.T) {
	question := "ORIGINAL " + strings.Repeat("Ein langer Auftrag. ", 30)

	for _, history := range []bool{false, true} {
		t.Run(fmt.Sprintf("history=%v", history), func(t *testing.T) {
			fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
				switch {
				case n <= 4:
					return stepWithText(fmt.Sprintf("c%d", n), fmt.Sprintf("Schritt %d", n)), nil
				case n == 5:
					return overflowed()
				}
				return text(fmt.Sprintf("Antwort %d", n)), nil
			}}

			w, _, _ := limitChat(t, fake, history, 16)
			ask(w, question)
			w.WaitFor(richText("Antwort 6"), 10*time.Second)

			w.Find(richText("ORIGINAL")).Exactly(1)
			w.Find(richText("Schritt 1")).Exactly(1)
			w.FindAll(richText("SUMMARY")).None()
			if r := fake.last(); hasRequestText(r, "ORIGINAL") || !hasRequestText(r, "SUMMARY") {
				t.Fatalf("the model must read the summary: %+v", r)
			}

			ask(w, "Noch eine Frage")
			w.WaitFor(richText("Antwort 7"), 10*time.Second)
			w.Find(richText("ORIGINAL")).Exactly(1)
			w.FindAll(richText("SUMMARY")).None()
			r := fake.last()
			if hasRequestText(r, "ORIGINAL") || !hasRequestText(r, "SUMMARY") || !hasRequestText(r, "Noch eine Frage") {
				t.Fatalf("the further request must build on the shortened version: %+v", r)
			}
		})
	}
}

// A run which fails after a compaction keeps the reached state: the view does not jump back and the question does not
// return to the input. Without a compaction, a failing run hands the question back.
func TestChat_FailedRunAfterACompactionKeepsItsState(t *testing.T) {
	boom := errors.New("provider down")
	question := "ORIGINAL " + strings.Repeat("Ein langer Auftrag. ", 30)

	tests := []struct {
		name string
		next func(n int) (completion.Result, error)
		// wantKept tells whether the steps stay, wantField whether the question is back in the input
		wantKept  bool
		wantField bool
	}{
		{
			name: "fails after steps and a compaction",
			next: func(n int) (completion.Result, error) {
				switch {
				case n <= 4:
					return stepWithText(fmt.Sprintf("c%d", n), fmt.Sprintf("Schritt %d", n)), nil
				case n == 5:
					return overflowed()
				}
				return completion.Result{}, boom
			},
			wantKept: true,
		},
		{
			name:      "fails at once",
			next:      func(int) (completion.Result, error) { return completion.Result{}, boom },
			wantField: true,
		},
	}

	for _, history := range []bool{false, true} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s/history=%v", tc.name, history), func(t *testing.T) {
				fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) { return tc.next(n) }}
				w, _, _ := limitChat(t, fake, history, 16)
				ask(w, question)
				waitIdle(w)

				if got := w.FindAll(richText("Schritt 1")).Len() == 1; got != tc.wantKept {
					t.Fatalf("steps shown = %v, want %v", got, tc.wantKept)
				}
				if got := w.FindAll(nagotest.And(nagotest.Label("Nachricht"), fieldWith("ORIGINAL"))).Len() == 1; got != tc.wantField {
					t.Fatalf("question in the field = %v, want %v", got, tc.wantField)
				}
			})
		}
	}
}

// A run which ends at its limit after a compaction shows its steps and the hint.
func TestChat_TurnLimitAfterACompaction(t *testing.T) {
	question := "ORIGINAL " + strings.Repeat("Ein langer Auftrag. ", 30)
	fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
		if n == 5 {
			return overflowed()
		}
		return stepWithText(fmt.Sprintf("c%d", n), fmt.Sprintf("Schritt %d", n)), nil
	}}

	w, sessions, _ := limitChat(t, fake, true, 6)
	ask(w, question)
	w.WaitFor(nagotest.Text("Fortsetzen"), 10*time.Second)
	w.Find(richText("ORIGINAL")).Exactly(1)
	w.Find(richText("Schritt 1")).Exactly(1)
	w.FindAll(richText("SUMMARY")).None()
	w.FindAll(nagotest.And(nagotest.Label("Nachricht"), fieldWith("ORIGINAL"))).None()

	for s, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		if !s.TurnLimited || len(s.Compacted) == 0 || !reflect.DeepEqual(s.Messages[len(s.Messages)-1], s.Compacted[len(s.Compacted)-1]) {
			t.Fatalf("both lists must be stored and end alike: %+v", s)
		}
	}
}

// MaxCompactions reaches the run in a transient chat, in a persisted one and after the answer to a question.
func TestChat_MaxCompactions(t *testing.T) {
	question := "ORIGINAL " + strings.Repeat("Ein langer Auftrag. ", 30)
	// four steps, then every other request overflows, with a step in between
	script := func(overflows int) func(ctx context.Context, n int) (completion.Result, error) {
		return func(_ context.Context, n int) (completion.Result, error) {
			switch {
			case n <= 4:
				return stepWithText(fmt.Sprintf("c%d", n), fmt.Sprintf("Schritt %d", n)), nil
			case n > 4+2*overflows:
				return text("Alles fertig."), nil
			case (n-4)%2 == 1:
				return overflowed()
			}
			return step(fmt.Sprintf("c%d", n)), nil
		}
	}

	tests := []struct {
		name      string
		max       int
		overflows int
		wantDone  bool
	}{
		{name: "two exactly reached", max: 2, overflows: 2, wantDone: true},
		{name: "two exceeded", max: 2, overflows: 3},
		{name: "default exactly reached", overflows: completion.DefaultMaxCompactions, wantDone: true},
		{name: "default exceeded", overflows: completion.DefaultMaxCompactions + 1},
	}

	for _, history := range []bool{false, true} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s/history=%v", tc.name, history), func(t *testing.T) {
				logged := captureErrors(t)
				fake := &recording{next: script(tc.overflows)}
				opts := uicompletion.ChatOptions{
					Sessions: testSessions(t), History: history, Completions: fake, Provider: echo.New("p", "p"),
					DisableCurrentTime: true, MaxTurns: 64, MaxCompactions: tc.max, Agents: noopAgent(),
				}
				w := openChatOptions(t, opts)
				ask(w, question)
				waitIdle(w)

				if got := w.FindAll(richText("Alles fertig.")).Len() == 1; got != tc.wantDone {
					t.Fatalf("done = %v, want %v", got, tc.wantDone)
				}
				if got := len(logged()) > 0; got == tc.wantDone {
					t.Fatalf("a limit of the compactions stays a failure: logged=%v", logged())
				}
				w.FindAll(nagotest.TextContains(limitHint)).None()
			})
		}
	}

	t.Run("after the answer to a question", func(t *testing.T) {
		fake := &recording{next: func(_ context.Context, n int) (completion.Result, error) {
			switch {
			case n == 1:
				return call("q1", completion.AskUserToolName, `{"question":"Welche?"}`), nil
			case n <= 5:
				return stepWithText(fmt.Sprintf("c%d", n), fmt.Sprintf("Schritt %d", n)), nil
			case n == 6:
				return overflowed()
			}
			return text("Alles fertig."), nil
		}}
		opts := uicompletion.ChatOptions{
			Sessions: testSessions(t), History: true, Completions: fake, Provider: echo.New("p", "p"),
			DisableCurrentTime: true, MaxCompactions: 1, Agents: noopAgent(), AskUser: true,
		}
		w := openChatOptions(t, opts)
		ask(w, question)
		w.WaitFor(richText("Welche?"), 10*time.Second)
		w.Type(w.Find(nagotest.Label("Eigene Antwort")), "Diese")
		w.Click(w.Find(nagotest.Text("Antworten")))
		w.WaitFor(richText("Alles fertig."), 10*time.Second)
	})
}

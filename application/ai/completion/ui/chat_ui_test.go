// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	stdjson "encoding/json"
	"iter"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// chatFake scripts the parent conversation and the sub-agents independently; it is safe for concurrent use.
type chatFake struct {
	parent func(ctx context.Context, opts completion.Options) (completion.Result, error)
	calls  atomic.Int32
}

func (f *chatFake) Models(auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) { yield(model.Model{ID: "fake"}, nil) }
}

func (f *chatFake) Complete(ctx context.Context, _ auth.Subject, opts completion.Options) (completion.Result, error) {
	f.calls.Add(1)
	if strings.HasPrefix(opts.System, "You are a sub-agent.") {
		return text("Antwort auf " + completion.FinalAnswer([]completion.Message{{Role: completion.Assistant, Content: opts.Messages[0].Content}})), nil
	}
	return f.parent(ctx, opts)
}

func (f *chatFake) Stream(context.Context, auth.Subject, completion.Options) iter.Seq2[completion.Delta, error] {
	return nil
}

func text(s string) completion.Result {
	return completion.Result{
		Message:    completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: s}}},
		StopReason: completion.StopEndTurn,
	}
}

// richText matches markdown views, which is how chat bubbles render.
func richText(contains string) nagotest.Matcher {
	return nagotest.Where("RichText("+contains+")", func(n nagotest.Node) bool {
		rt, ok := n.Component.(*proto.RichText)
		return ok && strings.Contains(string(rt.Value), contains)
	})
}

func openChat(t *testing.T, comps completion.Completions) *nagotest.Window {
	t.Helper()
	return openChatWith(t, comps, session.UseCases{}, false)
}

// testSessions wires in-memory session use cases, including the ReBAC rules the owner grant needs.
func testSessions(t *testing.T) session.UseCases {
	t.Helper()
	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatal(err)
	}

	rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Owner, Target: session.Namespace})
	for _, pid := range session.InstancePermissions {
		rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Relation(pid), Target: session.Namespace})
	}

	repo := session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](mem.NewBlobStore(string(session.Namespace))))
	return session.NewUseCases(repo, rdb)
}

func openChatWith(t *testing.T, comps completion.Completions, sessions session.UseCases, history bool) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.test")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions:           sessions,
				History:            history,
				Completions:        comps,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
				Delegation:         &uicompletion.DelegationOptions{},
			})
		})
	})

	return app.Open(t, user.SU(), "chat")
}

// delegatingParent delegates two tasks once and then answers.
func delegatingParent(ctx context.Context, opts completion.Options) (completion.Result, error) {
	last := opts.Messages[len(opts.Messages)-1]
	for _, c := range last.Content {
		if _, ok := c.(completion.ToolResult); ok {
			return text("Alles erledigt."), nil
		}
	}

	return completion.Result{
		Message: completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{
			ID: "d1", Name: completion.DelegateToolName,
			Arguments: stdjson.RawMessage(`{"tasks":[{"title":"Erste","task":"A"},{"title":"Zweite","task":"B"}]}`),
		}}},
		StopReason: completion.StopToolUse,
	}, nil
}

func TestChat_DelegationShowsSubTasks(t *testing.T) {
	fake := &chatFake{parent: delegatingParent}

	w := openChat(t, fake)
	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte parallel arbeiten")
	w.Click(w.Find(nagotest.Text("Senden")))

	w.WaitFor(richText("Alles erledigt."), 10*time.Second)
	w.Find(nagotest.TextContains("Teilaufgaben (2)")).Visible()
	if got := fake.calls.Load(); got != 4 {
		t.Fatalf("expected 2 parent and 2 sub requests, got %d", got)
	}
}

func TestChat_StopCancelsTheRun(t *testing.T) {
	cancelled := make(chan struct{})
	fake := &chatFake{parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
		<-ctx.Done()
		close(cancelled)
		return completion.Result{}, ctx.Err()
	}}

	w := openChat(t, fake)
	w.Type(w.Find(nagotest.Label("Nachricht")), "Das dauert")
	w.Click(w.Find(nagotest.Text("Senden")))

	w.WaitFor(nagotest.Text("Stopp"), 5*time.Second)
	w.Click(w.Find(nagotest.Text("Stopp")))

	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("stop did not cancel the provider request")
	}

	// Nothing happened yet, so the input is handed back and the chat is ready again.
	w.WaitFor(nagotest.Text("Senden"), 5*time.Second)
}

func TestChat_HistoryOpensSubTaskTranscript(t *testing.T) {
	sessions := testSessions(t)
	w := openChatWith(t, &chatFake{parent: delegatingParent}, sessions, true)

	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte parallel arbeiten")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("Alles erledigt."), 10*time.Second)

	// Each sub task was persisted as a child session and can be opened from the task list.
	open := w.FindAll(nagotest.Label("Verlauf der Teilaufgabe anzeigen")).Exactly(2)
	w.FindAll(nagotest.Type[*proto.Modal]()).None()
	w.Click(open.At(0))
	w.WaitFor(nagotest.And(nagotest.Type[*proto.Modal](), nagotest.Containing(richText("Antwort auf "))), 5*time.Second)

	// The history dialog lists only the conversation itself, never its sub tasks.
	n := 0
	for _, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("expected only the parent session, got %d", n)
	}
}

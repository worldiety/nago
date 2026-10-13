// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
)

// admin may do everything, like user.SU, but has an id, as a real administrator.
type admin struct {
	user.Subject
	id user.ID
}

func (a admin) ID() user.ID { return a.id }

func openChatAs(t *testing.T, subject user.Subject, comps completion.Completions, sessions session.UseCases) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.titletest")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Title:              "Assistent",
				Sessions:           sessions,
				History:            true,
				Completions:        comps,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
			})
		})
	})

	return app.Open(t, subject, "chat")
}

func onlySession(t *testing.T, sessions session.UseCases) session.Session {
	t.Helper()
	var res []session.Session
	for s, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		res = append(res, s)
	}

	if len(res) != 1 {
		t.Fatalf("expected one session, got %d", len(res))
	}

	return res[0]
}

// The conversation is summarized into its title once, after its first meaningful turn, and never again.
func TestChat_TitleIsSummarizedOnce(t *testing.T) {
	var titleCalls atomic.Int32
	comps := &chatFake{parent: func(_ context.Context, opts completion.Options) (completion.Result, error) {
		if strings.Contains(opts.System, "Titel einer Unterhaltung") {
			titleCalls.Add(1)
			return text("„Rechnung von gestern prüfen“."), nil
		}
		return text("ok"), nil
	}}

	sessions := testSessions(t)
	me := admin{Subject: user.SU(), id: "me"}
	w := openChatAs(t, me, comps, sessions)

	send := func(msg string) {
		w.Type(w.Find(nagotest.Label("Nachricht")), msg)
		w.Click(w.Find(nagotest.Text("Senden")))
		w.WaitFor(nagotest.Text("Senden"), 5*time.Second)
	}

	send("Hallo")
	time.Sleep(100 * time.Millisecond)
	if titleCalls.Load() != 0 || onlySession(t, sessions).Title != "" {
		t.Fatal("a greeting has no topic yet")
	}

	send("Bitte prüfe die Rechnung von gestern")
	deadline := time.Now().Add(5 * time.Second)
	for onlySession(t, sessions).Title != "Rechnung von gestern prüfen" {
		if time.Now().After(deadline) {
			t.Fatalf("expected the summarized title, got %q", onlySession(t, sessions).Title)
		}
		time.Sleep(10 * time.Millisecond)
	}

	send("Und bitte auch die Rechnung von vorgestern")
	time.Sleep(100 * time.Millisecond)
	if titleCalls.Load() != 1 {
		t.Fatalf("expected exactly one summary, got %d", titleCalls.Load())
	}
}

// The history lists only the own conversations, also for an administrator, who may read every session.
func TestChat_HistoryShowsOnlyOwnConversations(t *testing.T) {
	comps := &chatFake{parent: func(context.Context, completion.Options) (completion.Result, error) {
		return text("ok"), nil
	}}

	sessions := testSessions(t)
	foreign, err := sessions.Create(admin{Subject: user.SU(), id: "someone-else"}, session.CreateOptions{Model: "fake"})
	if err != nil {
		t.Fatal(err)
	}

	me := admin{Subject: user.SU(), id: "me"}
	w := openChatAs(t, me, comps, sessions)
	w.Type(w.Find(nagotest.Label("Nachricht")), "Hallo")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("ok"), 5*time.Second)

	w.Click(w.Find(nagotest.Text("Verlauf")))
	w.WaitFor(nagotest.Text("Hallo"), 5*time.Second)
	w.FindAll(nagotest.TextContains(string(foreign.ID))).None()
	w.Find(nagotest.Label("Verlauf löschen")).Exactly(1)
}

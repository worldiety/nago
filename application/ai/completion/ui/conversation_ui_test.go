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
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
)

// openConversationChat opens a persisted chat whose OnAttach rejects old Excel workbooks as an error of the user.
func openConversationChat(t *testing.T) (*nagotest.Window, session.UseCases) {
	t.Helper()
	sessions := testSessions(t)
	comps := &chatFake{parent: func(context.Context, completion.Options) (completion.Result, error) {
		return text("ok"), nil
	}}

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.conversationtest")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions:           sessions,
				History:            true,
				Completions:        comps,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
				FileUpload:         true,
				OnAttach: func(_ auth.Subject, att uicompletion.Attachment) ([]completion.Content, bool, error) {
					if strings.HasSuffix(att.Name, ".xls") {
						return nil, false, std.NewLocalizedError("Format nicht unterstützt", "Bitte "+att.Name+" als .xlsx speichern.")
					}
					return []completion.Content{completion.Text{Text: "abgelegt"}}, true, nil
				},
			})
		})
	})

	return app.Open(t, user.SU(), "chat"), sessions
}

func attach(w *nagotest.Window, name string) {
	w.Click(w.Find(nagotest.Text("Datei")))
	imports := w.Imports()
	w.Upload(string(imports[len(imports)-1].ID), nagotest.File(name, []byte("x")))
}

func countSessions(t *testing.T, sessions session.UseCases) int {
	t.Helper()
	n := 0
	for _, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	return n
}

func waitForSessions(t *testing.T, sessions session.UseCases, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for countSessions(t, sessions) != want {
		if time.Now().After(deadline) {
			t.Fatalf("expected %d sessions, got %d", want, countSessions(t, sessions))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A refused first turn shows the reason of the user error as a plain notice, keeps the file for a correction and
// leaves no empty conversation in the history.
func TestChat_RefusedFirstTurnLeavesNoConversation(t *testing.T) {
	w, sessions := openConversationChat(t)
	attach(w, "Alt.xls")
	w.Click(w.Find(nagotest.Text("Senden")))

	w.WaitFor(nagotest.Text("Bitte Alt.xls als .xlsx speichern."), 5*time.Second)
	w.Find(nagotest.Text("Format nicht unterstützt")).Exactly(1)
	w.FindAll(nagotest.TextContains("Code:")).None()
	w.Find(nagotest.TextContains("Alt.xls (1 B)")).Exactly(1)
	waitForSessions(t, sessions, 0)
}

// Files picked for one conversation are never sent with another one, and a conversation can be deleted from the
// history.
func TestChat_NewChatAndDeleteFromHistory(t *testing.T) {
	w, sessions := openConversationChat(t)
	w.Type(w.Find(nagotest.Label("Nachricht")), "Hallo")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("ok"), 5*time.Second)
	waitForSessions(t, sessions, 1)

	attach(w, "liste.txt")
	w.Find(nagotest.TextContains("liste.txt (1 B)")).Exactly(1)
	w.Click(w.Find(nagotest.Text("Neuer Chat")))
	w.FindAll(nagotest.TextContains("liste.txt (1 B)")).None()

	w.Click(w.Find(nagotest.Text("Verlauf")))
	w.Click(w.Find(nagotest.Label("Verlauf löschen")))
	w.Click(w.Find(nagotest.Text("Löschen")))
	waitForSessions(t, sessions, 0)
	w.WaitFor(nagotest.Text("Es gibt noch keine gespeicherten Verläufe."), 5*time.Second)
}

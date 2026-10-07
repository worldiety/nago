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
	"sync"
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
	"go.wdy.de/nago/presentation/core"
)

// The application takes over a workbook, which nago cannot send, even with a provider without a Files
// capability, and the notice next to the upload button tells the user where attachments go.
func TestChat_OnAttachAndUploadHint(t *testing.T) {
	var mutex sync.Mutex
	var sent []completion.Message
	comps := &chatFake{parent: func(_ context.Context, opts completion.Options) (completion.Result, error) {
		mutex.Lock()
		sent = opts.Messages
		mutex.Unlock()
		return text("ok"), nil
	}}

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.attachtest")
		cfg.RootView("chat", func(wnd core.Window) core.View {
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions:           session.UseCases{},
				Completions:        comps,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
				FileUpload:         true,
				UploadHint:         "Anhänge gehen an den KI-Anbieter.",
				OnAttach: func(_ auth.Subject, att uicompletion.Attachment) ([]completion.Content, bool, error) {
					if !strings.HasSuffix(att.Name, ".xlsx") {
						return nil, false, nil
					}
					return []completion.Content{completion.Text{Text: "Mappe " + att.Name + " abgelegt"}}, true, nil
				},
			})
		})
	})

	w := app.Open(t, user.SU(), "chat")
	w.Find(nagotest.Text("Anhänge gehen an den KI-Anbieter.")).Exactly(1)
	w.Click(w.Find(nagotest.Text("Datei")))
	imports := w.Imports()
	if len(imports) == 0 {
		t.Fatal("the upload button must request a file import")
	}

	w.Upload(string(imports[len(imports)-1].ID), nagotest.File("inventur.xlsx", []byte{0x50, 0x4b, 0x03, 0x04}))
	w.Type(w.Find(nagotest.Label("Nachricht")), "Bitte einlesen")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("ok"), 5*time.Second)

	mutex.Lock()
	defer mutex.Unlock()
	last := completion.FinalAnswer([]completion.Message{{Role: completion.Assistant, Content: sent[len(sent)-1].Content}})
	if !strings.Contains(last, "Mappe inventur.xlsx abgelegt") || !strings.Contains(last, "Bitte einlesen") {
		t.Fatalf("expected the content of the application in the user turn, got %q", last)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/nagotest"
)

// While a sub-agent works, its row tells what it does and the progress line counts the tasks, so that a long
// delegation visibly makes progress. The open row shows its latest steps and its time limit.
func TestChat_DelegationShowsWhatTheSubAgentDoes(t *testing.T) {
	arrived := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once

	type readIn struct {
		Name string `json:"name"`
	}
	type readOut struct {
		Rows int `json:"rows"`
	}
	reader := completion.NewTool("read_table", "reads a table", func(in readIn) (readOut, error) {
		once.Do(func() { close(arrived) })
		<-release
		return readOut{Rows: 42}, nil
	})

	fake := &scripted{
		parent: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				return text("Fertig."), nil
			}
			return call("d1", completion.DelegateToolName, `{"tasks":[{"title":"Lesen","task":"Lies die Tabelle"}]}`), nil
		},
		sub: func(ctx context.Context, opts completion.Options) (completion.Result, error) {
			if hasToolResult(opts) {
				return text("42 Zeilen"), nil
			}
			res := call("c1", "read_table", `{"name":"kunden"}`)
			res.Message.Content = append([]completion.Content{completion.Text{Text: "Ich lese die Tabelle."}}, res.Message.Content...)
			return res, nil
		},
	}

	w := openChatOptions(t, uicompletion.ChatOptions{
		Completions:        fake,
		Provider:           echo.New("p", "p"),
		DisableCurrentTime: true,
		Delegation:         &uicompletion.DelegationOptions{},
		Agents:             []uicompletion.Agent{{ID: "a", Name: "A", Tools: []completion.Tool{reader}}},
	})

	w.Type(w.Find(nagotest.Label("Nachricht")), "Lies")
	w.Click(w.Find(nagotest.Text("Senden")))
	select {
	case <-arrived:
	case <-time.After(10 * time.Second):
		t.Fatal("the sub-agent did not call its tool")
	}

	shown(t, w, "the task list", nagotest.Text("1 Hintergrundaufgabe"))
	shown(t, w, "its time and step", nagotest.TextContains(" · Schritt 1"))
	shown(t, w, "what the sub-agent does", nagotest.Text("führt read_table aus"))
	shown(t, w, "the progress line", nagotest.TextContains("1 Hintergrundaufgabe arbeitet, zuletzt: read_table"))

	// the steps are only rendered while the row is open
	w.FindAll(nagotest.Text("Ich lese die Tabelle.")).None()
	w.Click(w.Find(nagotest.Text("Lesen")))
	shown(t, w, "the step", nagotest.Text("Ich lese die Tabelle."))
	shown(t, w, "the running call", nagotest.Text("read_table"))
	shown(t, w, "the time limit", nagotest.Text("Zeitlimit 5:00"))

	close(release)
	w.WaitFor(richText("Fertig."), 10*time.Second)
	w.Find(nagotest.TextContains("1 erledigt"))
	w.Find(richText("42 Zeilen"))
	w.FindAll(nagotest.Text("Zeitlimit 5:00")).None()
	w.FindAll(nagotest.Text("führt read_table aus")).None()
}

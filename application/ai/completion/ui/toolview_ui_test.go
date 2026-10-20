// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"encoding/json"
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
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type suggestIn struct {
	Name string `json:"name"`
}

type suggestOut struct {
	Path string `json:"path"`
}

// suggestViewTool proposes a page, which the user opens by a button in the chat.
func suggestViewTool() completion.Tool {
	tool := completion.NewTool("suggest_view", "proposes a page to the user", func(in suggestIn) (suggestOut, error) {
		return suggestOut{Path: "details"}, nil
	})

	tool.View = func(wnd core.Window, args, result json.RawMessage) core.View {
		var in suggestIn
		var out suggestOut
		if json.Unmarshal(args, &in) != nil || json.Unmarshal(result, &out) != nil {
			return nil
		}

		return ui.SecondaryButton(func() {
			wnd.Navigation().ForwardTo(core.NavigationPath(out.Path), nil)
		}).Title("Ansicht „"+in.Name+"“ öffnen").Key("suggest_view", string(args))
	}

	return tool
}

// suggestingParent proposes a page once and then answers.
func suggestingParent(_ context.Context, opts completion.Options) (completion.Result, error) {
	last := opts.Messages[len(opts.Messages)-1]
	for _, c := range last.Content {
		if _, ok := c.(completion.ToolResult); ok {
			return text("Dort findest du es."), nil
		}
	}

	return completion.Result{
		Message: completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{
			ID: "s1", Name: "suggest_view", Arguments: json.RawMessage(`{"name":"Rechnungen"}`),
		}}},
		StopReason: completion.StopToolUse,
	}, nil
}

// The view of a tool call shows up below the collapsed tool section, and again when the conversation is restored.
func TestToolView(t *testing.T) {
	sessions := testSessions(t)
	comps := &chatFake{parent: suggestingParent}

	var restore session.ID
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uicompletion.toolviewtest")
		chat := func(wnd core.Window, kept *core.State[session.ID]) core.View {
			return uicompletion.Chat(wnd, uicompletion.ChatOptions{
				Sessions:           sessions,
				History:            true,
				Completions:        comps,
				Provider:           echo.New("p", "p"),
				DisableCurrentTime: true,
				Agents:             []uicompletion.Agent{{Tools: []completion.Tool{suggestViewTool()}}},
				Session:            kept,
			})
		}

		cfg.RootView("chat", func(wnd core.Window) core.View {
			return chat(wnd, nil)
		})
		cfg.RootView("restored", func(wnd core.Window) core.View {
			return chat(wnd, core.StateOf[session.ID](wnd, "kept").Init(func() session.ID { return restore }))
		})
		cfg.RootView("details", func(wnd core.Window) core.View {
			return ui.Text("details")
		})
	})

	w := app.Open(t, user.SU(), "chat")
	w.Type(w.Find(nagotest.Label("Nachricht")), "Wo sind die Rechnungen?")
	w.Click(w.Find(nagotest.Text("Senden")))
	answer := w.WaitFor(richText("Dort findest du es."), 5*time.Second)
	if answer.Node().Component.(*proto.RichText).ExternalLinkConfirmation == "" {
		t.Fatal("a link written by the model must not open an external address without asking")
	}
	w.Find(nagotest.Text("Werkzeug: suggest_view"))
	w.Click(w.Find(nagotest.Text("Ansicht „Rechnungen“ öffnen")).Visible())
	w.Find(nagotest.Text("details"))

	for s, err := range sessions.FindAll(user.SU(), session.FindAllOptions{}) {
		if err != nil {
			t.Fatal(err)
		}
		restore = s.ID
	}

	r := app.Open(t, user.SU(), "restored")
	r.Find(richText("Dort findest du es."))
	r.Find(nagotest.Text("Ansicht „Rechnungen“ öffnen"))
}

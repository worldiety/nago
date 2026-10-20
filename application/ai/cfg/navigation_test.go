// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// richText matches the rich text views, e.g. a chat bubble, which contain the given text.
func richText(text string) nagotest.Matcher {
	return nagotest.Where("RichText("+text+")", func(n nagotest.Node) bool {
		rt, ok := n.Component.(*proto.RichText)
		return ok && strings.Contains(string(rt.Value), text)
	})
}

// The floating assistant follows the user from page to page: it stays open and keeps its draft and conversation,
// although a navigation destroys the window together with its states.
func TestAssistantFollowsNavigation(t *testing.T) {
	llm := newFakeOpenAI(t, "Alpha", "alpha-1")

	var mgmt Management
	app := nagotest.New(t, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.assistantnavigationtest")
		mgmt = std.Must(Enable(c))
		for _, page := range []struct{ path, next core.NavigationPath }{{"a", "b"}, {"b", "a"}} {
			c.RootView(page.path, func(wnd core.Window) core.View {
				return ui.VStack(
					ui.Text("page "+string(page.path)),
					ui.PrimaryButton(func() {
						wnd.Navigation().ForwardTo(page.next, nil)
					}).Title("to "+string(page.next)),
					mgmt.Assistant.Button(wnd, AssistantOptions{Label: "Assistent", History: true}),
				)
			})
		}
	})

	cfg := app.Configurator()
	secrets := std.Must(cfg.SecretManagement())
	id := std.Must(secrets.UseCases.CreateSecret(user.SU(), openai.Settings{Name: "Alpha", BaseURL: llm.URL + "/v1"}))
	if err := secrets.UseCases.UpdateMySecretGroups(user.SU(), id, []group.ID{group.System}); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		if c, _ := mgmt.Assistant.Providers(user.SU()); len(c) == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("provider not loaded")
		}
		time.Sleep(10 * time.Millisecond)
	}

	w := app.Open(t, cfg.SysUser(), "a")
	w.Click(w.Find(nagotest.Label("Assistent")))
	w.Type(w.Find(nagotest.Label("Nachricht")), "Hallo")
	w.Click(w.Find(nagotest.Text("Senden")))
	w.WaitFor(richText("Antwort von Alpha"), 10*time.Second)
	w.Type(w.Find(nagotest.Label("Nachricht")), "Noch etwas")

	w.Click(w.Find(nagotest.Text("to b")))
	w.Find(nagotest.Text("page b"))
	w.Find(richText("Hallo"))
	w.Find(richText("Antwort von Alpha"))
	if field := w.Find(nagotest.Label("Nachricht")).Node().Component.(*proto.TextField); field.Value != "Noch etwas" {
		t.Fatalf("the draft is lost: %q", field.Value)
	}

	// closed stays closed, and opening it again still shows the conversation
	w.Click(w.Find(nagotest.Label("Schließen")))
	w.Click(w.Find(nagotest.Text("to a")))
	w.Find(nagotest.Text("page a"))
	w.FindAll(nagotest.Label("Nachricht")).None()
	w.Click(w.Find(nagotest.Label("Assistent")))
	w.Find(richText("Antwort von Alpha"))

	// another tab is another scope, which starts without the assistant
	other := app.Open(t, cfg.SysUser(), "a")
	other.FindAll(nagotest.Label("Nachricht")).None()
}

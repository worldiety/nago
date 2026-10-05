// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"fmt"
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

// A page opens the floating assistant with a suggested message, which the user still has to send.
func TestOpenWithDraft(t *testing.T) {
	llm := newFakeOpenAI(t, "Alpha", "alpha-1")

	var mgmt Management
	var opened bool
	app := nagotest.New(t, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.assistantopentest")
		mgmt = std.Must(Enable(c))
		c.RootView("page", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text(fmt.Sprintf("available: %v", mgmt.Assistant.Available(wnd))),
				ui.PrimaryButton(func() {
					opened = mgmt.Assistant.Open(wnd, "Bitte prüfe Version 4.")
				}).Title("Mit Assistent bearbeiten"),
				mgmt.Assistant.Button(wnd, AssistantOptions{Label: "Assistent"}),
			)
		})
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

	// nobody signed in: no assistant, and Open does nothing
	anon := app.Open(t, nil, "page")
	anon.Find(nagotest.Text("available: false"))
	anon.Click(anon.Find(nagotest.Text("Mit Assistent bearbeiten")))
	if opened {
		t.Fatal("Open must report an unavailable assistant")
	}

	w := app.Open(t, cfg.SysUser(), "page")
	w.Find(nagotest.Text("available: true"))
	w.FindAll(nagotest.Label("Nachricht")).None() // the panel is closed

	w.Click(w.Find(nagotest.Text("Mit Assistent bearbeiten")))
	if !opened {
		t.Fatal("Open must report the available assistant")
	}

	field := w.Find(nagotest.Label("Nachricht")).Node().Component.(*proto.TextField)
	if field.Value != "Bitte prüfe Version 4." {
		t.Fatalf("the draft is not in the input field: %q", field.Value)
	}

	if got := llm.requested(); len(got) != 0 {
		t.Fatalf("the draft must not be sent, got requests %v", got)
	}

	// an operator hides the assistant
	if err := std.Must(cfg.SettingsManagement()).UseCases.StoreGlobal(user.SU(), AssistantSettings{Hidden: true}); err != nil {
		t.Fatal(err)
	}

	app.Open(t, cfg.SysUser(), "page").Find(nagotest.Text("available: false"))
}

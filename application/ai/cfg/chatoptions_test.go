// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"errors"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// An embedded chat gets the same model and safety settings as the floating button.
func TestChatOptions(t *testing.T) {
	llm := newFakeOpenAI(t, "Alpha", "alpha-1")

	var mgmt Management
	var opts uicompletion.ChatOptions
	var optsErr error
	app := nagotest.New(t, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.chatoptionstest")
		mgmt = std.Must(Enable(c))
		c.RootView("chat", func(wnd core.Window) core.View {
			opts, optsErr = mgmt.Assistant.ChatOptions(wnd, AssistantOptions{
				Title:  "Hilfe",
				Agents: []uicompletion.Agent{{Name: "a"}, {Name: "pinned", Model: "own", MaxTokens: 7}},
			})
			return ui.Text("chat")
		})
		c.RootView("marked", func(wnd core.Window) core.View {
			opts, optsErr = mgmt.Assistant.ChatOptions(wnd, AssistantOptions{Confirmation: ConfirmationMarked})
			return ui.Text("marked")
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

	store := func(s AssistantSettings) {
		t.Helper()
		if err := std.Must(cfg.SettingsManagement()).UseCases.StoreGlobal(user.SU(), s); err != nil {
			t.Fatal(err)
		}
	}

	app.Open(t, nil, "chat")
	if !errors.Is(optsErr, ErrAssistantNotSignedIn) {
		t.Fatalf("expected not signed in, got %v", optsErr)
	}

	// the window context comes from the subject, and only a subject of the application carries its services
	store(AssistantSettings{ReadOnly: true})
	app.Open(t, cfg.SysUser(), "chat")
	if optsErr != nil {
		t.Fatal(optsErr)
	}

	if !opts.ReadOnly || !opts.ConfirmMutations || opts.Title != "Hilfe" || opts.Completions == nil || opts.Provider == nil {
		t.Fatalf("the operator settings are missing: %+v", opts)
	}

	if a := opts.Agents[0]; a.Model != "alpha-1" || a.MaxTokens != DefaultAssistantMaxTokens {
		t.Fatalf("the agent did not get the operator's model: %+v", a)
	}

	if a := opts.Agents[1]; a.Model != "own" || a.MaxTokens != 7 {
		t.Fatalf("a pinned agent must keep its model: %+v", a)
	}

	store(AssistantSettings{SkipConfirmation: true})
	app.Open(t, cfg.SysUser(), "chat")
	if optsErr != nil || opts.ConfirmMutations || opts.ReadOnly {
		t.Fatalf("the confirmation must follow the operator: %+v %v", opts, optsErr)
	}

	// an assistant which confirms only marked tools ignores the operator's default of confirming everything
	store(AssistantSettings{})
	app.Open(t, cfg.SysUser(), "marked")
	if optsErr != nil || opts.ConfirmMutations || !opts.ConfirmMarked {
		t.Fatalf("expected only marked tools to be confirmed: %+v %v", opts, optsErr)
	}

	// the operator may confirm only marked tools for every assistant which follows the global settings
	store(AssistantSettings{ConfirmMarkedOnly: true})
	app.Open(t, cfg.SysUser(), "chat")
	if optsErr != nil || opts.ConfirmMutations || !opts.ConfirmMarked {
		t.Fatalf("expected the operator's choice of marked tools only: %+v %v", opts, optsErr)
	}

	store(AssistantSettings{Hidden: true})
	app.Open(t, cfg.SysUser(), "chat")
	if !errors.Is(optsErr, ErrAssistantHidden) {
		t.Fatalf("expected hidden, got %v", optsErr)
	}
}

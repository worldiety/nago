// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command ai-example puts an AI assistant on top of an ordinary nago application.
//
// The entry point lives under cmd/ and does what an entry point does: it bootstraps the framework, enables
// the contexts and decides how they are reached. Every piece of business meaning is behind app/library.
package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgai "go.wdy.de/nago/application/ai/cfg"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	_ "go.wdy.de/nago/application/ai/provider/anthropic"
	_ "go.wdy.de/nago/application/ai/provider/mistralai"
	_ "go.wdy.de/nago/application/ai/provider/openai"
	aispeclink "go.wdy.de/nago/application/speclink/ai"
	cfgspeclink "go.wdy.de/nago/application/speclink/cfg"
	ailibrary "go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library/ai"
	cfglibrary "go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library/cfg"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_113")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		// The context wires itself: store, use cases, its role and its screens. Nothing about the library
		// leaks into this file.
		lib := option.Must(cfglibrary.Enable(cfg))

		modAI := option.Must(cfgai.Enable(cfg))

		// The requirement catalogue of this binary, with ready-made tools on top. Nothing about the library
		// is repeated here: spec.Declare filled the catalogue during package initialisation, and the tools
		// read it through use cases that audit the acting subject like every other one.
		specMod := option.Must(cfgspeclink.Enable(cfg))

		// Build the tools once. They receive the acting subject per call, so there is nothing per-user about
		// them and nothing to rebuild per turn.
		tools := append(ailibrary.Tools(lib.UseCases), aispeclink.Tools(specMod.UseCases)...)

		scaffold := cfg.NewScaffold().
			Login(true).
			MenuEntry().Title("Bestand").Forward(lib.Pages.Books).Private().
			Decorator()

		// The assistant hangs on the decorator rather than on a page, so it is genuinely on every screen -
		// including the administration pages the framework brings along. When it cannot run (no provider, no
		// model, hidden by the operator, missing role) Decorate returns the view untouched and logs why.
		cfg.SetDecorator(func(wnd core.Window, view core.View) core.View {
			return modAI.Assistant.Decorate(wnd, scaffold(wnd, view), cfgai.AssistantOptions{
				Title: "Bibliotheks-Assistent",
				Label: "Assistent",
				// One tag for the whole application: the assistant follows the user from screen to screen, so
				// a conversation started on one page and continued on another is one conversation.
				Tags:             []string{"tutorial113:assistant"},
				History:          true,
				AskUser:          true,
				ConfirmMutations: true,
				MaxTurns:         32,
				Agents: []uicompletion.Agent{{
					ID:    "librarian",
					Name:  "Bibliotheks-Assistent",
					Tools: tools,
					// Rebuilt on every question, because both derived halves move: the requirement index grows
					// with the project, and the situational half changes with every navigation. Only the
					// domain half is a constant, and it is the only one written by hand.
					SystemPromptFunc: func() string {
						return ailibrary.SystemPrompt + "\n\n" +
							aispeclink.Index(wnd.Subject(), specMod.UseCases) + "\n" +
							uicompletion.WindowContext(wnd)
					},
				}},
			})
		})
	}).Run()
}

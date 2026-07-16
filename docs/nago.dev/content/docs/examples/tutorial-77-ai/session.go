// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application/ai"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// sessionChat demonstrates the generic [uicompletion.Chat] component with persisted history. The whole chat -
// message list, input, progress, the restore-a-previous-conversation dialog and lazy session creation - is
// provided by the reusable component; the example only wires it to the first usable provider and enables
// History. Compare this with the low-level examples (stateless.go, agentic.go), which drive the completion API
// by hand.
func sessionChat(wnd core.Window, uc ai.UseCases, sessions session.UseCases) core.View {
	prov, comps, err := firstCompletionProvider(wnd.Subject(), uc, false)
	if err != nil {
		return alert.BannerError(err)
	}

	chat := uicompletion.Chat(wnd, uicompletion.ChatOptions{
		Sessions:    sessions,
		Completions: comps,
		Provider:    prov,
		Title:       "Persistente Session",
		History:     true,
		// Tags scope the restore dialog. All sessions of this demo page share one tag, so the history button
		// only lists conversations started here.
		Tags:   []string{"tutorial-77:session"},
		Agents: []uicompletion.Agent{{SystemPrompt: "You are a helpful assistant. Keep answers concise."}},
	})

	return ui.VStack(
		ui.Text("Persistente Session (uicompletion.Chat mit History)").Font(ui.Title),
		ui.Text("Der komplette Chat inkl. Verlauf-Wiederherstellung kommt aus der wiederverwendbaren uicompletion-Komponente."),
		chat,
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

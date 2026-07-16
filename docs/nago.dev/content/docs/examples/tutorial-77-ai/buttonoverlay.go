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

// buttonOverlayPage demonstrates the floating [uicompletion.ChatButton]: a page with arbitrary content that
// gets an assistant button anchored to the bottom-right corner. Clicking it toggles the same chat panel used
// by the embedded variant. The button is configured in the nago value style (Corner/Label), the chat itself
// via ChatOptions.
func buttonOverlayPage(wnd core.Window, uc ai.UseCases, sessions session.UseCases) core.View {
	prov, comps, err := firstCompletionProvider(wnd.Subject(), uc, true)
	if err != nil {
		return alert.BannerError(err)
	}

	button := uicompletion.ChatButton(uicompletion.ChatOptions{
		Sessions:    sessions,
		Completions: comps,
		Provider:    prov,
		Title:       "KI-Assistent",
		History:     true,
		FileUpload:  true,
		Tags:        []string{"tutorial-77:overlay"},
		Agents:      []uicompletion.Agent{{SystemPrompt: "You are a friendly assistant that helps the user navigate this demo application. When the user attaches a file, use its content in your answer."}},
	}).Corner(uicompletion.CornerBottomRight).Label("Assistent")

	return ui.VStack(
		ui.Text("Floating Chat-Button").Font(ui.Title),
		ui.Text("Diese Seite hat beliebigen Inhalt. Unten rechts sitzt der wiederverwendbare uicompletion.ChatButton, der das Chat-Panel ein- und ausklappt."),
		ui.Text("Der Button ist per Position: Fixed an der Bildschirmecke verankert und überlagert die Seite unabhängig vom Scrollzustand."),
		button,
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

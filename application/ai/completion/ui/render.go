// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package uicompletion provides generic, reusable chat UI on top of the stateless [completion] API and the
// persistable [session] use cases.
//
// It offers two entry points:
//
//   - [Chat]: an embeddable chat view, configured entirely by code via [ChatOptions] (with/without persisted
//     history, with/without an agent picker, with/without file upload, with/without tools and the built-in
//     ask_user clarification tool).
//   - [ChatButton]: a floating button that sits in a screen corner and toggles the same [Chat] panel.
//
// The design deliberately keeps every domain-specific concern out: agents are a plain, caller-populated
// [Agent] slice, tools are supplied as per-turn factories, and the persisted history is scoped only by the
// opaque [ChatOptions.Tags]. Downstream contexts can therefore build their own assistant experiences directly
// on top of this package.
package uicompletion

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// scrollAnchorID is the stable id of the invisible element the conversation scrolls to after every update, so
// the newest message is always in view.
const scrollAnchorID = "uicompletion-end-of-history"

// conversationView renders the given history as scrollable chat bubbles plus a stable scroll anchor. emptyHint
// is shown (as a muted line) when the history contains no visible message yet.
//
// The scroll anchor must live at a fixed index inside a fixed-size parent. nago's UiStack.vue renders children
// with a keyless v-for and computes the DOM id once (non-reactive) at setup. If the anchor were a sibling of
// the growing bubble list, Vue would reuse the anchor's instance for new bubbles on every history update,
// leaking a stale id onto them and breaking the scroll. Keeping the anchor as a stable second child of a
// two-child stack avoids the instance reuse entirely.
func conversationView(history []completion.Message, emptyHint string, height ui.Length) core.View {
	bubbles := renderHistory(history)
	if len(bubbles) == 0 && emptyHint != "" {
		bubbles = append(bubbles, ui.Text(emptyHint).Font(ui.BodySmall))
	}

	return ui.ScrollView(
		ui.VStack(
			ui.VStack(bubbles...).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
			ui.VStack().ID(scrollAnchorID).Frame(ui.Frame{}.Size(ui.L2, ui.L2)),
		).FullWidth().Alignment(ui.Leading),
	).Axis(ui.ScrollViewAxisVertical).
		ScrollToView(scrollAnchorID, ui.ScrollAnimationSmooth).
		Frame(ui.Frame{Height: height, Width: ui.Full})
}

// renderHistory turns the stateless message history into chat bubbles. Tool calls are shown as muted hints;
// tool results (which live inside follow-up user messages) are omitted.
func renderHistory(history []completion.Message) []core.View {
	var views []core.View
	for _, m := range history {
		for _, c := range m.Content {
			switch v := c.(type) {
			case completion.Text:
				if strings.TrimSpace(v.Text) == "" {
					continue
				}
				views = append(views, chatBubble(m.Role, v.Text))
			case completion.ToolCall:
				views = append(views, ui.HStack(
					ui.Text("→ Tool: "+v.Name).Font(ui.Small),
				).FullWidth().Alignment(ui.Leading))
			}
		}
	}
	return views
}

func chatBubble(role completion.Role, text string) core.View {
	bg := ui.M3
	if role == completion.User {
		bg = ui.M4
	}

	bubble := ui.VStack(markdown.RichText(text)).
		Alignment(ui.Leading).
		BackgroundColor(bg).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{MaxWidth: "85%"})

	if role == completion.User {
		return ui.HStack(ui.Spacer(), bubble).FullWidth()
	}
	return ui.HStack(bubble, ui.Spacer()).FullWidth()
}

// chatFrame wraps the given body into the chat panel chrome (header with title, optional header actions and a
// close button, card styling, fixed width). It is used by the floating [ChatButton] panel.
func chatFrame(body core.View, title string, actions core.View, open *core.State[bool]) core.View {
	header := ui.HStack(
		ui.Text(title).Font(ui.TitleSmall),
		ui.Spacer(),
		ui.If(actions != nil, actions),
		ui.TertiaryButton(func() {
			open.Set(false)
		}).PreIcon(icons.Close).AccessibilityLabel("Schließen"),
	).Gap(ui.L4).FullWidth().Alignment(ui.Center)

	return ui.VStack(
		header,
		body,
	).Gap(ui.L8).
		Alignment(ui.Leading).
		BackgroundColor(ui.M1).
		Border(ui.Border{}.Radius(ui.L16).Color(ui.M4).Width(ui.L1).Shadow(ui.L8)).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{Width: ui.L560})
}

// thinkingLabel builds the progress line shown while the model is composing the next turn. From the second
// turn onward it appends the (1-based) step number so the user can tell a long, multi-turn run is still making
// progress. turn is the zero-based loop index reported by [completion.Progress].
func thinkingLabel(turn int) string {
	if turn > 0 {
		return fmt.Sprintf("… die KI denkt nach (Schritt %d)", turn+1)
	}
	return "… die KI denkt nach"
}

// toolLabel builds the progress line shown while a single tool call is executing.
func toolLabel(name string) string {
	if name == "" {
		return "… die KI führt ein Werkzeug aus"
	}
	return fmt.Sprintf("… die KI führt das Werkzeug „%s“ aus", name)
}

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
	"encoding/json"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accordion"
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
func conversationView(wnd core.Window, history []completion.Message, emptyHint string, height ui.Length, hv historyView) core.View {
	bubbles := renderHistory(wnd, history, hv)
	if len(bubbles) == 0 && emptyHint != "" {
		bubbles = append(bubbles, ui.Text(emptyHint).Font(ui.BodySmall))
	}

	return ui.ScrollView(
		ui.VStack(
			ui.VStack(bubbles...).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
			ui.VStack().ID(hv.idPrefix+scrollAnchorID).Frame(ui.Frame{}.Size(ui.L2, ui.L2)),
		).FullWidth().Alignment(ui.Leading),
	).Axis(ui.ScrollViewAxisVertical).
		ScrollToView(hv.idPrefix+scrollAnchorID, ui.ScrollAnimationSmooth).
		ScrollBehavior(ui.ScrollBehaviorAuto).
		Frame(ui.Frame{Height: height, Width: ui.Full, MinHeight: "0dp"})
}

// historyView configures [renderHistory].
type historyView struct {
	// idPrefix keeps the ids of stateful elements unique when a second conversation (a sub-agent transcript) is
	// rendered in the same window.
	idPrefix string

	// openChild, when set, offers to open the persisted transcript of a sub-agent.
	openChild func(id session.ID)
}

// renderHistory turns the stateless message history into chat bubbles. Consecutive tool calls are shown as one
// collapsed section and reasoning as a collapsed section; tool results (which live inside follow-up user messages) and the prompts
// the loop injects on its own (see [completion.IsLoopPrompt]) are omitted. Calls of the delegation tools are
// shown as a collapsible list of their sub tasks.
func renderHistory(wnd core.Window, history []completion.Message, hv historyView) []core.View {
	// ask_user calls are a dialog with the user, so question and answer are shown as regular bubbles.
	askCalls := map[string]bool{}

	// Delegation calls are rendered together with their results, which live in a later message, and with the
	// newest state of their background tasks.
	results := map[string]completion.ToolResult{}
	for _, m := range history {
		for _, c := range m.Content {
			if r, ok := c.(completion.ToolResult); ok {
				results[r.ToolCallID] = r
			}
		}
	}
	latest := latestTasks(history)

	// An agentic run may call dozens of tools before it answers. Consecutive tool calls, and the reasoning between
	// and after them, are therefore folded into one collapsed section, which any other visible message ends.
	var views []core.View
	var group []groupItem
	flush := func() {
		if len(group) > 0 {
			views = append(views, groupView(wnd, group)...)
			group = nil
		}
	}
	add := func(v core.View) {
		flush()
		views = append(views, v)
	}

	for i, m := range history {
		if completion.IsLoopPrompt(m) {
			continue
		}
		for j, c := range m.Content {
			id := fmt.Sprintf("%suicompletion-%d-%d", hv.idPrefix, i, j)
			switch v := c.(type) {
			case completion.Thinking:
				if strings.TrimSpace(v.Text) == "" {
					continue
				}
				group = append(group, groupItem{id: id, thinking: v.Text})
			case completion.Text:
				if strings.TrimSpace(v.Text) == "" {
					continue
				}
				add(chatBubble(m.Role, v.Text))
			case completion.ToolCall:
				if v.Name == askUserToolName {
					if q, ok := askQuestion(v); ok {
						askCalls[v.ID] = true
						add(chatBubble(completion.Assistant, q))
						continue
					}
				}
				if completion.IsDelegationTool(v.Name) {
					res, done := results[v.ID]
					if view := tasksView(wnd, hv, v, res, done, latest); view != nil {
						add(view)
						continue
					}
				}
				group = append(group, groupItem{id: id, tool: v.Name})
			case completion.ToolResult:
				if !askCalls[v.ToolCallID] || v.IsError {
					continue
				}
				if a, ok := askAnswer(v); ok {
					add(chatBubble(completion.User, a))
				}
			}
		}
	}

	flush()
	return views
}

// groupItem is a tool call or a reasoning block of a run of consecutive ones, see [groupView].
type groupItem struct {
	id       string // stable while the history grows, it keeps the section open or closed
	tool     string
	thinking string
}

// groupView renders consecutive tool calls as a single collapsed section. Reasoning without any tool call keeps
// its own collapsed section.
func groupView(wnd core.Window, items []groupItem) []core.View {
	var tools []string
	for _, it := range items {
		if it.tool != "" {
			tools = append(tools, it.tool)
		}
	}

	if len(tools) == 0 {
		var views []core.View
		for _, it := range items {
			views = append(views, thinkingView(wnd, it.id+"-thinking", it.thinking))
		}
		return views
	}

	title := fmt.Sprintf("%d Werkzeugaufrufe", len(tools))
	if len(tools) == 1 {
		title = "Werkzeug: " + tools[0]
	}

	// the section is collapsed already, so the reasoning is shown inline instead of in a nested section
	body := make([]core.View, 0, len(items))
	for _, it := range items {
		if it.tool != "" {
			body = append(body, ui.Text("→ "+it.tool).Font(ui.Small))
			continue
		}
		body = append(body, ui.Text(strings.TrimSpace(it.thinking)).Font(ui.Small).Color(ui.M8))
	}

	open := core.StateOf[bool](wnd, items[0].id+"-tools")
	return []core.View{ui.HStack(
		accordion.Accordion(
			ui.Text(title).Font(ui.Small),
			ui.VStack(body...).Gap(ui.L4).Alignment(ui.Leading).FullWidth(),
			open,
		).Small().HideSeparator().Frame(ui.Frame{MaxWidth: "85%"}),
		ui.Spacer(),
	).FullWidth()}
}

// askQuestion extracts the question of an ask_user call.
func askQuestion(call completion.ToolCall) (string, bool) {
	var in struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal(call.Arguments, &in); err != nil || strings.TrimSpace(in.Question) == "" {
		return "", false
	}
	return in.Question, true
}

// askAnswer extracts the user's answer from an ask_user tool result.
func askAnswer(res completion.ToolResult) (string, bool) {
	for _, c := range res.Content {
		t, ok := c.(completion.Text)
		if !ok {
			continue
		}
		var out struct {
			Answer string `json:"answer"`
		}
		if err := json.Unmarshal([]byte(t.Text), &out); err == nil && strings.TrimSpace(out.Answer) != "" {
			return out.Answer, true
		}
	}
	return "", false
}

// thinkingView renders a reasoning block as a collapsed, muted section the user can expand on demand.
func thinkingView(wnd core.Window, id, text string) core.View {
	open := core.StateOf[bool](wnd, id)
	return ui.HStack(
		accordion.Accordion(
			ui.Text("Gedankengang").Font(ui.Small),
			ui.VStack(markdown.RichText(text)).Alignment(ui.Leading).Font(ui.Small),
			open,
		).Small().HideSeparator().Frame(ui.Frame{MaxWidth: "85%"}),
		ui.Spacer(),
	).FullWidth()
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
		Frame(ui.Frame{Width: ui.L560, MaxWidth: panelMaxWidth, MaxHeight: panelMaxHeight})
}

const (
	// panelMaxWidth keeps the floating panel on a narrow screen, e.g. of a phone, with a gutter on both sides.
	panelMaxWidth ui.Length = "calc(100vw - 4rem)"

	// panelMaxHeight keeps the floating panel below the app bar and above its button. The conversation shrinks
	// instead, see conversationView.
	panelMaxHeight ui.Length = "calc(100dvh - 13rem)"
)

// busyLine shows the progress of a run behind a small spinning ring, so that a long run visibly keeps working.
func busyLine(label string) core.View {
	ring := ui.Border{}.Circle().Width(ui.L2).Color(ui.M5)
	ring.TopColor = ui.I0

	return ui.HStack(
		ui.VStack().
			Animation(ui.AnimateSpin).
			Border(ring).
			Frame(ui.Frame{}.Size(ui.L12, ui.L12)),
		ui.Text(label).Font(ui.BodySmall),
	).Gap(ui.L8).Alignment(ui.Center)
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

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// renderDecisions renders the decisions a suspended run waits on: clarifying questions (predefined options
// plus a free-text answer) and approvals of mutating calls. Once every pending call is decided, onResolve
// receives all resolutions at once. onDismiss drops the pending decisions, e.g. when the user moves on.
//
// The panel is rendered from the continuation alone, which comes from the persisted session (or the transient
// chat's view state). Nothing blocks while it is shown, so the user may close the chat or navigate away and
// answer later.
func renderDecisions(wnd core.Window, cont *completion.Continuation, revision int, busy bool, onResolve func([]completion.Resolution), onDismiss func()) core.View {
	// Decisions collected so far, keyed by this very suspension so a newer one never inherits older choices.
	key := suspensionKey(cont, revision)
	decided := core.StateOf[map[string]completion.Resolution](wnd, "uicompletion-decisions-"+key)

	decide := func(r completion.Resolution) {
		all := make(map[string]completion.Resolution, len(cont.Pending))
		for k, v := range decided.Get() {
			all[k] = v
		}
		all[r.CallID] = r
		decided.Set(all)

		if len(all) < len(cont.Pending) {
			return
		}

		resolutions := make([]completion.Resolution, 0, len(all))
		for _, pc := range cont.Pending {
			resolutions = append(resolutions, all[pc.Call.ID])
		}
		onResolve(resolutions)
	}

	var items []core.View
	for _, pc := range cont.Pending {
		if _, done := decided.Get()[pc.Call.ID]; done {
			continue
		}

		switch pc.Kind {
		case completion.PendingQuestion:
			items = append(items, questionPanel(wnd, pc, key, busy, decide))
		case completion.PendingApproval:
			items = append(items, approvalPanel(wnd, pc, busy, decide))
		}
	}

	return ui.VStack(
		ui.VStack(items...).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
		ui.HStack(
			ui.Spacer(),
			ui.TertiaryButton(onDismiss).Title("Verwerfen und neues Thema beginnen").Enabled(!busy),
		).FullWidth(),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
}

// suspensionKey identifies one suspension: the revision alone is not enough for a transient chat, which has
// none, but the ids of the pending calls are unique per model turn.
func suspensionKey(cont *completion.Continuation, revision int) string {
	ids := make([]string, 0, len(cont.Pending))
	for _, pc := range cont.Pending {
		ids = append(ids, pc.Call.ID)
	}
	return fmt.Sprintf("%d-%s", revision, strings.Join(ids, "_"))
}

// questionPanel renders the answer input of one clarifying question. The question itself is already shown as a
// chat bubble (see [renderHistory]).
func questionPanel(wnd core.Window, pc completion.PendingCall, key string, busy bool, decide func(completion.Resolution)) core.View {
	free := core.StateOf[string](wnd, "uicompletion-answer-"+key+"-"+pc.Call.ID)
	answer := func(text string) {
		if strings.TrimSpace(text) == "" || busy {
			return
		}
		decide(completion.Resolution{CallID: pc.Call.ID, Answer: text})
	}

	var options []core.View
	for _, opt := range pc.Options {
		opt := opt
		options = append(options, ui.HStack(ui.Text(opt)).Action(func() {
			answer(opt)
		}).FullWidth().Border(ui.Border{}.Color(ui.I0).Width(ui.L1).Radius(ui.L8)).Padding(ui.Padding{}.All(ui.L8)))
	}

	return ui.VStack(
		ui.Text("Antwort auf die Rückfrage").Font(ui.TitleSmall),
		ui.VStack(options...).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
		ui.TextField("Eigene Antwort", free.Get()).
			InputValue(free).
			FullWidth().
			Disabled(busy).
			KeydownEnter(func() { answer(free.Get()) }),
		ui.HStack(
			ui.Spacer(),
			ui.PrimaryButton(func() { answer(free.Get()) }).
				Title("Antworten").
				Enabled(!busy && strings.TrimSpace(free.Get()) != ""),
		).FullWidth(),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))
}

// approvalPanel renders the review of one mutating call. The arguments are shown verbatim, because a summary
// written by the model is exactly what must not be trusted here; an approval executes exactly these.
func approvalPanel(wnd core.Window, pc completion.PendingCall, busy bool, decide func(completion.Resolution)) core.View {
	subject := wnd.Subject()
	effect := pc.Effect
	if effect == "" {
		effect = fmt.Sprintf("%s()", pc.Call.Name)
	}

	return ui.VStack(
		ui.Text(StrConfirmTitle.Get(subject)).Font(ui.TitleSmall),
		ui.Text(StrConfirmIntro.Get(subject)),
		ui.Text(effect).Font(ui.BodyMedium),
		ui.Text(pc.Call.Name+" "+prettyArguments(pc.Call.Arguments)).Font(ui.Monospace),
		ui.HStack(
			ui.Spacer(),
			ui.SecondaryButton(func() {
				decide(completion.Resolution{CallID: pc.Call.ID, Message: StrConfirmDeclined.Get(subject)})
			}).Title(StrConfirmDecline.Get(subject)).Enabled(!busy),
			ui.PrimaryButton(func() {
				decide(completion.Resolution{CallID: pc.Call.ID, Approved: true})
			}).Title(StrConfirmApprove.Get(subject)).Enabled(!busy),
		).Gap(ui.L8).FullWidth(),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))
}

// optimisticAnswers renders the user's decisions as the tool-result turn the run will append, so the answers
// show up as bubbles immediately while the model is still working.
func optimisticAnswers(cont *completion.Continuation, resolutions []completion.Resolution) completion.Message {
	isQuestion := map[string]bool{}
	for _, pc := range cont.Pending {
		isQuestion[pc.Call.ID] = pc.Kind == completion.PendingQuestion
	}

	var content []completion.Content
	for _, r := range resolutions {
		if !isQuestion[r.CallID] || r.Dismissed {
			continue
		}
		buf, _ := json.Marshal(struct {
			Answer string `json:"answer"`
		}{r.Answer})
		content = append(content, completion.ToolResult{ToolCallID: r.CallID, Content: []completion.Content{completion.Text{Text: string(buf)}}})
	}

	return completion.Message{Role: completion.User, Content: content}
}

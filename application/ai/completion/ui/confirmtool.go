// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrConfirmTitle = i18n.MustString("nago.ai.confirm.title", i18n.Values{
		language.German:  "Änderung bestätigen",
		language.English: "Confirm change",
	})
	StrConfirmIntro = i18n.MustString("nago.ai.confirm.intro", i18n.Values{
		language.German:  "Der Assistent möchte etwas ändern. Prüfen Sie, was passieren soll:",
		language.English: "The assistant wants to make a change. Check what is about to happen:",
	})
	StrConfirmApprove = i18n.MustString("nago.ai.confirm.approve", i18n.Values{
		language.German:  "Ausführen",
		language.English: "Execute",
	})
	StrConfirmDecline = i18n.MustString("nago.ai.confirm.decline", i18n.Values{
		language.German:  "Ablehnen",
		language.English: "Decline",
	})
	StrConfirmDeclined = i18n.MustString("nago.ai.confirm.declined", i18n.Values{
		language.German:  "Der Nutzer hat diese Änderung abgelehnt. Führe sie nicht aus und frage nach, wie stattdessen vorgegangen werden soll.",
		language.English: "The user declined this change. Do not perform it and ask how to proceed instead.",
	})
)

// pendingConfirm is a mutating tool call waiting for the user's decision. approve is a buffered channel the
// (background) run goroutine blocks on until the user decided in the UI - the same mechanism [pendingAsk]
// uses for clarifying questions.
type pendingConfirm struct {
	// Tool is the name of the tool about to run.
	Tool string
	// Effect is the tool's own description of what it changes ([completion.Tool.Confirm]).
	Effect string
	// Arguments is the pretty-printed JSON the model chose.
	Arguments string

	approve chan bool
}

// confirmMutationGate returns a [completion.BeforeToolCallFunc] that holds every mutating tool call until the
// user approves it.
//
// The point of doing this here rather than in the system prompt is that it cannot be talked around. A prompt
// saying "always ask before writing" is a request to the model; this is a gate in front of the function call.
// Non-mutating tools pass through untouched, so a conversation that only reads never stops to ask.
func confirmMutationGate(wnd core.Window, confirm *core.State[*pendingConfirm]) completion.BeforeToolCallFunc {
	return func(_ auth.Subject, tool completion.Tool, call completion.ToolCall) error {
		if !tool.Mutating {
			return nil
		}

		ch := make(chan bool, 1)
		pc := &pendingConfirm{
			Tool:      tool.Def.Name,
			Effect:    tool.Confirm,
			Arguments: prettyArguments(call.Arguments),
			approve:   ch,
		}

		wnd.Post(func() { confirm.Set(pc) })

		if <-ch {
			return nil
		}

		// Reported to the model as a tool error rather than aborting the run, so it can propose something
		// else instead of the whole turn collapsing on a perfectly reasonable "no".
		return errors.New(StrConfirmDeclined.Get(wnd.Subject()))
	}
}

// prettyArguments formats the raw tool arguments for human review. The arguments are shown verbatim on
// purpose: a summary written by the model is exactly the thing that must not be trusted here.
func prettyArguments(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}

	return buf.String()
}

// decideConfirm delivers the user's decision to the blocked run goroutine and clears the pending request.
func decideConfirm(confirm *core.State[*pendingConfirm], pc *pendingConfirm, approved bool) {
	if pc == nil {
		return
	}

	select {
	case pc.approve <- approved:
	default:
	}

	confirm.Set(nil)
}

// renderConfirm renders the pending mutation for review.
func renderConfirm(wnd core.Window, confirm *core.State[*pendingConfirm], pc *pendingConfirm) core.View {
	effect := pc.Effect
	if effect == "" {
		effect = fmt.Sprintf("%s()", pc.Tool)
	}

	return ui.VStack(
		ui.Text(StrConfirmTitle.Get(wnd.Subject())).Font(ui.TitleSmall),
		ui.Text(StrConfirmIntro.Get(wnd.Subject())),
		ui.Text(effect).Font(ui.BodyMedium),
		ui.Text(pc.Tool+" "+pc.Arguments).Font(ui.Monospace),
		ui.HStack(
			ui.Spacer(),
			ui.SecondaryButton(func() {
				decideConfirm(confirm, pc, false)
			}).Title(StrConfirmDecline.Get(wnd.Subject())),
			ui.PrimaryButton(func() {
				decideConfirm(confirm, pc, true)
			}).Title(StrConfirmApprove.Get(wnd.Subject())),
		).Gap(ui.L8).FullWidth(),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))
}

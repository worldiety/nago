// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// historyDialog renders the "restore a previous conversation" dialog when present is set. It lists the
// subject's persisted sessions carrying all of the given tags (most recently updated first) as clickable
// cards; picking one invokes onPick with the full session so the caller can load its history back into the
// panel and continue it. The dialog is cancelable and returns nil while not presented. Errors while listing
// are surfaced inline so the user can still cancel out.
func historyDialog(wnd core.Window, sessionUC session.UseCases, tags []string, present *core.State[bool], onPick func(session.Session)) core.View {
	if !present.Get() {
		return nil
	}

	var sessions []session.Session
	for s, err := range sessionUC.FindAll(wnd.Subject(), session.FindAllOptions{Tags: tags}) {
		if err != nil {
			return alert.Dialog("Gespeicherte Verläufe", alert.BannerError(err), present, alert.Closeable())
		}
		sessions = append(sessions, s)
	}

	// Most recently updated first, so the conversation the user most likely wants to resume is on top.
	slices.SortFunc(sessions, func(a, b session.Session) int {
		return int(b.UpdatedAt) - int(a.UpdatedAt)
	})

	var body core.View
	if len(sessions) == 0 {
		body = ui.Text("Es gibt noch keine gespeicherten Verläufe.").Font(ui.BodySmall)
	} else {
		rows := make([]core.View, 0, len(sessions))
		for _, s := range sessions {
			s := s
			rows = append(rows, historyCard(wnd, s, func() {
				present.Set(false)
				onPick(s)
			}))
		}
		body = ui.VStack(rows...).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
	}

	return alert.Dialog("Gespeicherte Verläufe", body, present, alert.Closeable(), alert.Larger())
}

// historyCard renders one selectable conversation in the history dialog: a short preview (session title or
// first user message via session.String()), the last-update timestamp and the number of messages. Clicking
// the card restores the conversation via onPick.
func historyCard(wnd core.Window, s session.Session, onPick func()) core.View {
	when := s.UpdatedAt.Time(wnd.Location()).Format("2006-01-02 15:04")

	return ui.VStack(
		ui.Text(s.String()).Font(ui.Title).Frame(ui.Frame{MaxWidth: ui.Full}),
		ui.HStack(
			ui.Text(when).Font(ui.Small),
			ui.Spacer(),
			ui.Text(fmt.Sprintf("%d Nachrichten", len(s.Messages))).Font(ui.Small),
		).FullWidth().Alignment(ui.Center),
	).Gap(ui.L4).
		FullWidth().
		Alignment(ui.Leading).
		Action(onPick).
		BackgroundColor(ui.M2).
		Border(ui.Border{}.Radius(ui.L8).Color(ui.M4).Width(ui.L1)).
		Padding(ui.Padding{}.All(ui.L12))
}

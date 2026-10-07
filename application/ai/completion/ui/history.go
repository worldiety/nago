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
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// historyDialog renders the "restore a previous conversation" dialog when present is set. It lists the
// subject's persisted sessions carrying all of the given tags (most recently updated first) as clickable
// cards; picking one invokes onPick with the full session so the caller can load its history back into the
// panel and continue it. The dialog is cancelable and returns nil while not presented. Errors while listing
// are surfaced inline so the user can still cancel out.
func historyDialog(wnd core.Window, sessionUC session.UseCases, tags []string, present *core.State[bool], onPick func(session.Session), onDeleted func(session.ID)) core.View {
	if !present.Get() {
		return nil
	}

	// the conversation to delete, after the user confirmed it
	toDelete := core.StateOf[session.ID](wnd, "uicompletion-history-delete")
	confirmDelete := core.StateOf[bool](wnd, "uicompletion-history-delete-confirm")
	deleting := core.StateOf[bool](wnd, "uicompletion-history-deleting")

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
		rows := make([]core.View, 0, len(sessions)+1)
		rows = append(rows, deleteDialog(wnd, sessionUC, toDelete, confirmDelete, deleting, onDeleted))
		for _, s := range sessions {
			s := s
			rows = append(rows, ui.HStack(
				historyCard(wnd, s, func() {
					present.Set(false)
					onPick(s)
				}),
				// beside the card, a click on it would pick the conversation as well
				ui.TertiaryButton(func() {
					toDelete.Set(s.ID)
					confirmDelete.Set(true)
				}).PreIcon(icons.TrashBin).AccessibilityLabel("Verlauf löschen").Enabled(!deleting.Get()),
			).Gap(ui.L4).FullWidth().Alignment(ui.Center))
		}
		body = ui.VStack(rows...).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
	}

	return alert.Dialog("Gespeicherte Verläufe", body, present, alert.Closeable(), alert.Larger())
}

// historyCard renders one selectable conversation in the history dialog: a short preview (session title or
// first user message via session.String()), the last-update timestamp and the number of messages. Clicking
// the card restores the conversation via onPick.
// deleteDialog asks before a conversation is deleted. The deletion runs in the background, because it also deletes
// what the application and the provider keep for the conversation, see [session.OnDelete].
func deleteDialog(wnd core.Window, sessionUC session.UseCases, toDelete *core.State[session.ID], present, deleting *core.State[bool], onDeleted func(session.ID)) core.View {
	if !present.Get() || toDelete.Get() == "" {
		return nil
	}

	id := toDelete.Get()
	return alert.Dialog(
		"Verlauf löschen",
		ui.Text("Soll dieser Verlauf mit allen Nachrichten und Anhängen gelöscht werden? Das lässt sich nicht rückgängig machen."),
		present,
		alert.Cancel(nil),
		alert.Delete(func() {
			deleting.Set(true)
			subject := wnd.Subject()
			go func() {
				err := sessionUC.Delete(subject, id)
				wnd.Post(func() {
					deleting.Set(false)
					toDelete.Set("")
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					if onDeleted != nil {
						onDeleted(id)
					}
				})
			}()
		}),
	)
}

func historyCard(wnd core.Window, s session.Session, onPick func()) core.View {
	when := s.UpdatedAt.Time(wnd.Location()).Format("2006-01-02 15:04")

	return ui.VStack(
		ui.Text(s.String()).Font(ui.Title).Frame(ui.Frame{MaxWidth: ui.Full}),
		ui.HStack(
			ui.Text(when).Font(ui.Small),
			ui.If(s.Pending != nil, ui.Text("• wartet auf deine Antwort").Font(ui.Small).Color(ui.I0)),
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

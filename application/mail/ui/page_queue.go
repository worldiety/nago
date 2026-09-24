// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/tags"
)

type queueFilter struct {
	key    string
	title  string
	status []mail.Status
	stuck  bool
}

var queueFilters = []queueFilter{
	{key: "", title: "Alle"},
	{key: string(mail.StatusQueued), title: "Wartet", status: []mail.Status{mail.StatusQueued, mail.StatusUndefined}},
	{key: string(mail.StatusError) + "," + string(mail.StatusFailed) + "," + string(mail.StatusSuppressed), title: "Alle Fehler", status: []mail.Status{mail.StatusError, mail.StatusFailed, mail.StatusSuppressed}},
	{key: string(mail.StatusError), title: "Wird wiederholt", status: []mail.Status{mail.StatusError}},
	{key: string(mail.StatusFailed), title: "Endgültig fehlgeschlagen", status: []mail.Status{mail.StatusFailed}},
	{key: string(mail.StatusSuppressed), title: "Unterdrückt", status: []mail.Status{mail.StatusSuppressed}},
	{key: string(mail.StatusSendSuccess), title: "Versendet", status: []mail.Status{mail.StatusSendSuccess}},
	{key: "stuck", title: "Hängend", stuck: true},
}

// QueuePage lists the outgoing mails. The query parameters status (comma separated), server and stuck narrow
// the result.
func QueuePage(wnd core.Window, pages Pages, uc mail.UseCases) core.View {
	statusParam := wnd.Values()["status"]
	server := wnd.Values()["server"]
	stuck := wnd.Values()["stuck"] == "1"

	filter := mail.OutgoingFilter{Server: server}
	if stuck {
		filter.Stuck = time.Hour
	}

	for _, s := range strings.Split(statusParam, ",") {
		if s = strings.TrimSpace(s); s != "" {
			filter.Status = append(filter.Status, mail.Status(s))
		}
	}

	if slices.Contains(filter.Status, mail.StatusQueued) {
		filter.Status = append(filter.Status, mail.StatusUndefined)
	}

	activeKey := statusParam
	if stuck {
		activeKey = "stuck"
	}

	var filterButtons []core.View
	for _, f := range queueFilters {
		values := core.Values{}
		if server != "" {
			values["server"] = server
		}

		switch {
		case f.stuck:
			values["stuck"] = "1"
		case f.key != "":
			values["status"] = f.key
		}

		action := func() { wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, values) }
		if f.key == activeKey {
			filterButtons = append(filterButtons, ui.PrimaryButton(action).Title(f.title))
		} else {
			filterButtons = append(filterButtons, ui.SecondaryButton(action).Title(f.title))
		}
	}

	if server != "" {
		filterButtons = append(filterButtons, ui.HStack(
			tags.StatusBadge(ui.SV0, "Server: "+server),
			ui.TertiaryButton(func() {
				values := wnd.Values().Clone().Delete("server")
				wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, values)
			}).Title("Filter entfernen"),
		).Gap(ui.L4))
	}

	return ui.VStack(
		header(wnd, pages, tabQueue, "Warteschlange", []crumb{{title: "Warteschlange", action: func() {
			wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, nil)
		}}}),
		ui.HStack(filterButtons...).Gap(ui.L8).Wrap(true).Alignment(ui.Leading),
		dataview.FromData(wnd, dataview.Data[mail.Outgoing, mail.ID]{
			ID:      "nago-mail-queue",
			FindAll: uc.FindOutgoingIDs(wnd.Subject(), filter),
			FindByID: func(id mail.ID) (option.Opt[mail.Outgoing], error) {
				return uc.FindOutgoingByID(wnd.Subject(), id)
			},
			Fields: queueFields(wnd),
		}).
			Search(true).
			Action(func(e mail.Outgoing) {
				wnd.Navigation().ForwardTo(pages.OutgoingMail, core.Values{"id": string(e.ID)})
			}).
			NextActionIndicator(true).
			SelectOptions(
				dataview.SelectOption[mail.ID]{
					Icon: heroOutline.ArrowPath,
					Name: "Erneut versuchen",
					Action: func(selected []mail.ID) error {
						if err := uc.RetryOutgoing(wnd.Subject(), selected...); err != nil {
							return err
						}

						alert.ShowBannerMessage(wnd, alert.Message{Title: "Erneut eingereiht", Message: fmt.Sprintf("%d Mails werden beim nächsten Durchlauf erneut versendet.", len(selected)), Intent: alert.IntentOk})
						return nil
					},
				},
				dataview.NewSelectOptionDelete(wnd, func(selected []mail.ID) error {
					for _, id := range selected {
						if err := uc.DeleteOutgoingByID(wnd.Subject(), id); err != nil {
							return err
						}
					}

					return nil
				}),
			),
	).Alignment(ui.TopLeading).Gap(ui.L16).FullWidth()
}

func queueFields(wnd core.Window) []dataview.Field[mail.Outgoing] {
	return []dataview.Field[mail.Outgoing]{
		{
			ID:   "receiver",
			Name: "Empfänger",
			Map: func(obj mail.Outgoing) core.View {
				return ui.Text(xstrings.EllipsisEnd(obj.Receiver, 40))
			},
			Comparator: func(a, b mail.Outgoing) int { return strings.Compare(a.Receiver, b.Receiver) },
		},
		{
			ID:   "subject",
			Name: "Betreff",
			Map: func(obj mail.Outgoing) core.View {
				return ui.Text(xstrings.EllipsisEnd(obj.Subject, 50))
			},
			Comparator: func(a, b mail.Outgoing) int { return strings.Compare(a.Subject, b.Subject) },
			Visible:    dataview.MinSizeMedium(),
		},
		{
			ID:   "status",
			Name: "Status",
			Map: func(obj mail.Outgoing) core.View {
				return statusPill(obj.Status)
			},
			Comparator: func(a, b mail.Outgoing) int { return strings.Compare(string(a.Status), string(b.Status)) },
		},
		{
			ID:   "attempts",
			Name: "Versuche",
			Map: func(obj mail.Outgoing) core.View {
				return ui.Text(fmt.Sprint(obj.Attempted()))
			},
			Comparator: func(a, b mail.Outgoing) int { return a.Attempted() - b.Attempted() },
			Visible:    dataview.MinSizeLarge(),
		},
		{
			ID:   "server",
			Name: "Server",
			Map: func(obj mail.Outgoing) core.View {
				return ui.Text(obj.ServerName)
			},
			Comparator: func(a, b mail.Outgoing) int { return strings.Compare(a.ServerName, b.ServerName) },
			Visible:    dataview.MinSizeXL(),
		},
		{
			ID:   "error",
			Name: "Letzter Fehler",
			Map: func(obj mail.Outgoing) core.View {
				if obj.LastError == "" {
					return ui.Text("–")
				}
				return ui.Text(xstrings.EllipsisEnd(obj.LastError, 50)).Font(ui.MonoSmall)
			},
			Visible: dataview.MinSizeXL(),
		},
		{
			ID:   "queued",
			Name: "Eingereiht",
			Map: func(obj mail.Outgoing) core.View {
				return ui.Text(formatTime(wnd, obj.QueuedAt))
			},
			Comparator: func(a, b mail.Outgoing) int { return a.QueuedAt.Compare(b.QueuedAt) },
			Visible:    dataview.MinSizeMedium(),
		},
	}
}

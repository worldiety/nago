// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisms

import (
	"strings"

	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/sms"
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	StrSMSQueue  = i18n.MustString("nago.sms.queue.title", i18n.Values{language.English: "SMS Queue", language.German: "SMS Ausgangsliste"})
	StrSMSQueued = i18n.MustString("nago.sms.queue.queued", i18n.Values{language.English: "queued", language.German: "wartet auf Versand"})
	StrSMSSent   = i18n.MustString("nago.sms.queue.sent", i18n.Values{language.English: "sent", language.German: "versendet"})
	StrSMSFailed = i18n.MustString("nago.sms.queue.failed", i18n.Values{language.English: "failed", language.German: "Fehler beim Versand"})
)

func PageQueue(wnd core.Window, uc sms.UseCases) core.View {
	return ui.VStack(
		ui.H1(StrSMSQueue.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[message.SMS, message.ID]{
			FindAll: uc.FindAllMessageIDs(wnd.Subject()),
			FindByID: func(id message.ID) (option.Opt[message.SMS], error) {
				return uc.FindMessageByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[message.SMS]{
				{
					ID:   "phone",
					Name: rstring.LabelReceiver.Get(wnd),
					Map: func(obj message.SMS) core.View {
						return ui.Text(obj.Recipient.String())
					},
				},
				{
					ID:   "body",
					Name: rstring.LabelText.Get(wnd),
					Map: func(obj message.SMS) core.View {
						return ui.Text(xstrings.EllipsisEnd(obj.Body, 30))
					},
				},
				{
					ID:   "status",
					Name: rstring.LabelState.Get(wnd),
					Map: func(obj message.SMS) core.View {
						switch obj.Status {
						case message.StatusSent:
							return ui.Text(StrSMSSent.Get(wnd))
						case message.StatusFailed:
							return ui.Text(StrSMSFailed.Get(wnd))
						case message.StatusQueued:
							return ui.Text(StrSMSQueued.Get(wnd))
						default:
							return ui.Text(string(obj.Status))
						}
					},
					Comparator: func(a, b message.SMS) int {
						return strings.Compare(string(a.Status), string(b.Status))
					},
				},
				{
					ID:   "created",
					Name: rstring.LabelCreatedAt.Get(wnd),
					Map: func(obj message.SMS) core.View {
						return ui.Text(date.Format(wnd.Locale(), date.TimeMinute, obj.CreatedAt.Time(wnd.Location())))
					},
					Comparator: func(a, b message.SMS) int {
						return int(a.CreatedAt - b.CreatedAt)
					},
					Visible: func(ctx dataview.FieldContext) bool {
						return ctx.Window.Info().SizeClass > core.SizeClassMedium || ctx.Style == dataview.Card
					},
				},
				{
					ID:   "sent",
					Name: rstring.LabelSentAt.Get(wnd),
					Map: func(obj message.SMS) core.View {
						return ui.Text(date.Format(wnd.Locale(), date.TimeMinute, obj.CreatedAt.Time(wnd.Location())))
					},
					Comparator: func(a, b message.SMS) int {
						return int(a.SendAt - b.SendAt)
					},
					Visible: func(ctx dataview.FieldContext) bool {
						return ctx.Window.Info().SizeClass > core.SizeClassMedium || ctx.Style == dataview.Card
					},
				},
			},
		}).SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(selected []message.ID) error {
				for _, id := range selected {
					if err := uc.DeleteMessageByID(wnd.Subject(), id); err != nil {
						return err
					}
				}

				return nil
			}),
		),
	).
		Alignment(ui.Leading).
		FullWidth()
}

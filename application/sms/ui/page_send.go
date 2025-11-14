// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisms

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/sms"
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrSendSMS       = i18n.MustString("nago.sms.test.title", i18n.Values{language.English: "Send SMS", language.German: "SMS senden"})
	StrTestSMSSend   = i18n.MustString("nago.sms.admin.send_msg_title", i18n.Values{language.English: "Send SMS", language.German: "SMS versenden"})
	StrTestSMSQueued = i18n.MustString("nago.sms.admin.send_msg_desc", i18n.Values{language.English: "SMS was queued.", language.German: "SMS wurde in die Warteschlange eingefügt."})
)

func PageSend(wnd core.Window, uc sms.UseCases) core.View {
	text := core.AutoState[string](wnd).Init(func() string {
		return "This is a test SMS."
	})
	phone := core.AutoState[string](wnd)

	originator := core.AutoState[string](wnd).Init(func() string {
		return "Test GmbH"
	})

	return ui.VStack(
		ui.H1(StrSendSMS.Get(wnd)),
		form.Card(
			ui.TextField(rstring.LabelReceiver.Get(wnd), phone.Get()).InputValue(phone).FullWidth(),
			ui.TextField(rstring.LabelSender.Get(wnd), originator.Get()).InputValue(originator).FullWidth(),
			ui.TextField(rstring.LabelText.Get(wnd), text.Get()).InputValue(text).Lines(3).FullWidth(),
		).Gap(ui.L8).
			Frame(ui.Frame{}.FullWidth()),
		ui.HLine(),
		ui.HStack(
			ui.PrimaryButton(func() {
				num, err := message.NewMSISDN(phone.Get())
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				id, err := uc.Send(wnd.Subject(), message.SendRequested{
					Recipient:  num,
					Originator: message.Originator(originator.Get()),
					Body:       text.Get(),
				}, sms.SendOptions{})

				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				slog.Info("sms send or queued", "id", id)
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   StrSMSSent.Get(wnd),
					Message: StrTestSMSQueued.Get(wnd),
					Intent:  alert.IntentOk,
				})
			}).Title(rstring.ActionSend.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing),
	).Alignment(ui.Leading).
		Gap(ui.L8).
		Frame(ui.Frame{}.Larger())
}

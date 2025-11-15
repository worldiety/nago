// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uichatbot

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/chatbot"
	"go.wdy.de/nago/application/chatbot/message"
	"go.wdy.de/nago/application/chatbot/user"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrSendChatMsg       = i18n.MustString("nago.chatbot.test.title", i18n.Values{language.English: "Send Message", language.German: "Nachricht senden"})
	StrTestChatMsgSend   = i18n.MustString("nago.chatbot.admin.send_msg_title", i18n.Values{language.English: "Send Message", language.German: "Nachricht versenden"})
	StrTestChatMsgQueued = i18n.MustString("nago.chatbot.admin.send_msg_desc", i18n.Values{language.English: "Message was queued.", language.German: "Message wurde in die Warteschlange eingefügt."})
)

func PageSend(wnd core.Window, uc chatbot.UseCases) core.View {
	text := core.AutoState[string](wnd).Init(func() string {
		return "This is a test message."
	})
	mail := core.AutoState[string](wnd)

	return ui.VStack(
		ui.H1(StrSendChatMsg.Get(wnd)),
		form.Card(
			ui.TextField(rstring.LabelReceiver.Get(wnd), mail.Get()).InputValue(mail).FullWidth(),
			ui.TextField(rstring.LabelText.Get(wnd), text.Get()).InputValue(text).Lines(3).FullWidth(),
		).Gap(ui.L8).
			Frame(ui.Frame{}.FullWidth()),
		ui.HLine(),
		ui.HStack(
			ui.PrimaryButton(func() {

				id, err := uc.Send(wnd.Subject(), message.SendRequested{
					RecipientByMail: user.Email(mail.Get()),
					Text:            text.Get(),
				}, chatbot.SendOptions{})

				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				slog.Info("chatbot message send or queued", "id", id)
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   StrTestChatMsgSend.Get(wnd),
					Message: StrTestChatMsgQueued.Get(wnd),
					Intent:  alert.IntentOk,
				})
			}).Title(rstring.ActionSend.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing),
	).Alignment(ui.Leading).
		Gap(ui.L8).
		Frame(ui.Frame{}.Larger())
}

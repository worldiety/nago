// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"os"

	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	StrMessages  = i18n.MustString("nago.ai.admin.messages", i18n.Values{language.English: "Messages", language.German: "Nachrichten"})
	StrRole      = i18n.MustString("nago.ai.admin.role", i18n.Values{language.English: "Role", language.German: "Rolle"})
	StrContent   = i18n.MustString("nago.ai.admin.content", i18n.Values{language.English: "Content", language.German: "Inhalt"})
	StrCreatedAt = i18n.MustString("nago.ai.admin.created_at", i18n.Values{language.English: "Created at", language.German: "Erstellt am"})
	StrOpenChat  = i18n.MustString("nago.ai.admin.open_chat", i18n.Values{language.English: "Open Chat", language.German: "Chat öffnen"})
)

func PageConversation(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()

	optConv := prov.Conversations()
	if optConv.IsNone() {
		return alert.BannerError(err)
	}

	conv := optConv.Unwrap()

	cid := conversation.ID(wnd.Values()["conversation"])
	optRefConv, err := conv.FindByID(wnd.Subject(), cid)
	if err != nil {
		return alert.BannerError(err)
	}

	if optRefConv.IsNone() {
		return alert.BannerError(fmt.Errorf("conversation not found: %s: %w", cid, os.ErrNotExist))
	}

	refConv := optRefConv.Unwrap()

	messages := conv.Conversation(wnd.Subject(), cid)

	return ui.VStack(
		breadcrumb.Breadcrumbs(ui.TertiaryButton(func() {
			wnd.Navigation().Back()
		}).Title(rstring.ActionBack.Get(wnd))).ClampLeading(),

		ui.H1(refConv.Name),
		ui.Text(refConv.Description),
		ui.Space(ui.L48),
		msgTable(wnd, prov, messages),
	).FullWidth().Alignment(ui.Leading)
}

func msgTable(wnd core.Window, prov provider.Provider, messages provider.Conversation) core.View {
	msgs := messages

	loadedMsgs := core.AutoState[[]message.Message](wnd).AsyncInit(func() []message.Message {
		v, err := xslices.Collect2(msgs.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	return ui.VStack(
		ui.H2(StrMessages.Get(wnd)),
		ui.If(!loadedMsgs.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),

		ui.IfFunc(loadedMsgs.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedMsgs.Get(), []dataview.Field[dataview.Element[message.Message]]{
				{
					Name: StrCreatedAt.Get(wnd),
					Map: func(obj dataview.Element[message.Message]) core.View {
						return ui.Text(date.Format(wnd.Subject().Language(), date.TimeMinute, obj.Value.CreatedAt.Time(wnd.Location())))
					},
				},
				{
					Name: StrRole.Get(wnd),
					Map: func(obj dataview.Element[message.Message]) core.View {
						return ui.Text(string(obj.Value.Role))
					},
				},
				{
					Name: StrContent.Get(wnd),
					Map: func(obj dataview.Element[message.Message]) core.View {
						return renderContent(wnd, obj.Value)
					},
				},
			}).CreateActionView(ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo("admin/ai/chat", wnd.Values().Put("provider", string(prov.Identity())).Put("conversation", string(messages.Identity())))
			}).Title(StrOpenChat.Get(wnd)))
		}),
	).FullWidth().Alignment(ui.Leading)

}

func renderContent(wnd core.Window, msg message.Message) core.View {
	switch {
	case msg.MessageInput.IsSome():
		return ui.Text(msg.MessageInput.Unwrap())
	case msg.MessageOutput.IsSome():
		return ui.Text(msg.MessageOutput.Unwrap())
	case msg.ToolExecution.IsSome():
		msg := msg.ToolExecution.Unwrap()
		return ui.Text(msg.Type + ": " + msg.Arguments)
	case msg.File.IsSome():
		msg := msg.File.Unwrap()
		return ui.Text("File: " + msg.Name + " (" + string(msg.ID) + ")\n" + string(msg.MimeType))
	case msg.DocumentURL.IsSome():
		msg := msg.DocumentURL.Unwrap()
		return ui.Text("Document-URL: " + msg.Name + " (" + string(msg.URL) + ")\n")
	default:
		return ui.Text(fmt.Sprintf("rendering not implemented: %#v", msg))
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrNewChat            = i18n.MustString("nago.ai.chats.new_chat", i18n.Values{language.English: "New Chat", language.German: "Neuer Chat"})
	StrChats              = i18n.MustString("nago.ai.chats.chats", i18n.Values{language.English: "Chats", language.German: "Chats"})
	StrDeleteChatTitle    = i18n.MustString("nago.ai.chats.delete", i18n.Values{language.English: "Delete Chat?", language.German: "Chat löschen?"})
	StrDeleteChatMessageX = i18n.MustVarString("nago.ai.chats.delete_msg_x", i18n.Values{language.English: "Do you want to delete the chat `{x}`?", language.German: "Soll der Chat `{x}` wirklich gelöscht werden?"})
)

type TChats struct {
	selected *core.State[conversation.ID]
	frame    ui.Frame
	provider provider.Provider
}

func Chats(provider provider.Provider, selected *core.State[conversation.ID]) TChats {
	return TChats{selected: selected, provider: provider}
}

func (c TChats) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	deleteState := core.DerivedState[conversation.Conversation](c.selected, "-selected-delete")
	deletePresented := core.DerivedState[bool](c.selected, "-delete-presented")

	optConvs := c.provider.Conversations()
	if optConvs.IsNone() {
		return alert.BannerError(fmt.Errorf("provider does not support conversations: %s", c.provider.Identity())).Render(ctx)
	}

	convs := optConvs.Unwrap()
	conversations, err := xslices.Collect2(convs.All(wnd.Subject()))
	if err != nil {
		return alert.BannerError(err).Render(ctx)
	}

	return ui.VStack(
		c.deleteDialog(deleteState, deletePresented),
		ui.TertiaryButton(func() {
			c.selected.Set("")
		}).PreIcon(icons.Pen).Title(StrNewChat.Get(wnd)),
		ui.HStack(ui.Text(StrChats.Get(wnd)).Color(ui.ColorIconsMuted).Font(ui.BodySmall)).Padding(ui.Padding{Left: ui.L16}),
	).
		Append(
			ui.ForEach(conversations, func(conv conversation.Conversation) core.View {

				title := xstrings.EllipsisEnd(c.title(conv), 20)
				var btnView core.View
				if conv.ID == c.selected.Get() {
					btnView = ui.SecondaryButton(nil).Title(title)
				} else {
					btnView = ui.TertiaryButton(func() {
						c.selected.Set(conv.ID)
						c.selected.Notify()
					}).Title(title)
				}

				return ui.HStack(
					btnView,
					ui.Spacer(),
					ui.Menu(
						ui.TertiaryButton(nil).PreIcon(icons.DotsVertical),
						ui.MenuGroup(
							ui.MenuItem(func() {
								deleteState.Set(conv)
								deletePresented.Set(true)
							}, ui.Text(rstring.ActionDelete.Get(wnd))),
						),
					),
				).Gap(ui.L4).FullWidth()
			})...,
		).BackgroundColor(ui.M2).
		Alignment(ui.TopLeading).
		Gap(ui.L4).
		Border(ui.Border{RightColor: ui.M5, RightWidth: ui.L1}).
		Padding(ui.Padding{}.All(ui.L4)).
		Frame(c.frame).
		Render(ctx)
}

func (c TChats) Frame(frame ui.Frame) TChats {
	c.frame = frame
	return c
}

func (c TChats) title(conv conversation.Conversation) string {
	if conv.Name != "" {
		return conv.Name
	}

	return string(conv.ID)
}

func (c TChats) deleteDialog(selected *core.State[conversation.Conversation], presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		StrDeleteChatTitle.Get(selected.Window()),
		ui.Text(StrDeleteChatMessageX.Get(selected.Window(), i18n.String("x", c.title(selected.Get())))),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {

			if err := c.provider.Conversations().Unwrap().Delete(selected.Window().Subject(), selected.Get().ID); err != nil {
				alert.ShowBannerError(selected.Window(), err)
				return
			}

			// try not to re-render absent chat
			if c.selected.Get() == selected.Get().ID {
				c.selected.Set("")
			}
		}),
	)
}
